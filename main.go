// SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metric names exposed by the exporter; shared with tests so the
// producer/consumer contract lives in one place.
const (
	metricSecondsToMidnight = "doomsdayclock_seconds_to_midnight"
	metricScrapeSuccess     = "doomsdayclock_scrape_success"
	metricCacheStale        = "doomsdayclock_cache_stale"
)

var version = "unknown"

var (
	secondsDesc = prometheus.NewDesc(
		metricSecondsToMidnight,
		"Current Doomsday Clock setting in seconds to midnight",
		nil, nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		metricScrapeSuccess,
		"Whether the last scrape of the Bulletin was successful (1=ok, 0=fail)",
		nil, nil,
	)
	cacheStaleDesc = prometheus.NewDesc(
		metricCacheStale,
		"Whether the served value is past its TTL (1=stale/degraded, 0=fresh)",
		nil, nil,
	)
)

var secondsRe = regexp.MustCompile(`(\d+)\s+seconds?\s+to\s+midnight`)

type bulletinPage struct {
	Excerpt struct {
		Rendered string `json:"rendered"`
	} `json:"excerpt"`
}

// cache holds the last successfully scraped value with thread-safe access.
type cache struct {
	mu        sync.RWMutex
	value     float64
	fetchedAt time.Time
	hasValue  bool
}

func (c *cache) store(v float64) {
	c.mu.Lock()
	c.value = v
	c.fetchedAt = time.Now()
	c.hasValue = true
	c.mu.Unlock()
}

func (c *cache) load() (float64, time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value, c.fetchedAt, c.hasValue
}

// scraper fetches the Doomsday Clock value in the background and caches it.
type scraper struct {
	client   *http.Client
	url      string
	ttl      time.Duration
	interval time.Duration
	// Fields rather than constants, so the retry cadence runs at test speed.
	initialBackoff time.Duration
	maxBackoff     time.Duration
	cache          cache
}

// newScraper builds the production scraper: retry after 2s, doubling to a 60s ceiling.
func newScraper(cfg config) *scraper {
	return &scraper{
		client:         &http.Client{Timeout: cfg.fetchTimeout},
		url:            cfg.scrapeURL,
		ttl:            cfg.cacheTTL,
		interval:       cfg.scrapeInterval,
		initialBackoff: 2 * time.Second,
		maxBackoff:     60 * time.Second,
	}
}

// drainBudget is the window in-flight requests get to finish during a shutdown.
func drainBudget() time.Duration {
	return 10 * time.Second
}

func (s *scraper) Describe(ch chan<- *prometheus.Desc) {
	ch <- secondsDesc
	ch <- scrapeSuccessDesc
	ch <- cacheStaleDesc
}

// staleAt reports whether a value fetched at fetchedAt is past its TTL at now.
func staleAt(now, fetchedAt time.Time, ttl time.Duration) bool {
	return now.Sub(fetchedAt) > ttl
}

func (s *scraper) Collect(ch chan<- prometheus.Metric) {
	value, fetchedAt, hasValue := s.cache.load()

	if !hasValue {
		ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, 0)
		ch <- prometheus.MustNewConstMetric(cacheStaleDesc, prometheus.GaugeValue, 1)
		return
	}

	isStale := staleAt(time.Now(), fetchedAt, s.ttl)

	ch <- prometheus.MustNewConstMetric(secondsDesc, prometheus.GaugeValue, value)

	if isStale {
		ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, 0)
		ch <- prometheus.MustNewConstMetric(cacheStaleDesc, prometheus.GaugeValue, 1)
	} else {
		ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, 1)
		ch <- prometheus.MustNewConstMetric(cacheStaleDesc, prometheus.GaugeValue, 0)
	}
}

// run starts the background scrape loop. Fetches immediately, then at
// s.interval. On failure, applies exponential backoff up to maxBackoff.
// It returns when ctx is cancelled so the loop unwinds during shutdown.
func (s *scraper) run(ctx context.Context) {
	// wait blocks for d unless ctx is cancelled first; it returns false on
	// cancellation so the loop can exit instead of sleeping through shutdown.
	wait := func(d time.Duration) bool {
		t := time.NewTimer(d)
		defer t.Stop()
		select {
		case <-ctx.Done():
			return false
		case <-t.C:
			return true
		}
	}

	backoff := time.Duration(0)

	for {
		value, err := s.fetch(ctx)
		if err != nil {
			// First failure waits initialBackoff, then doubles up to maxBackoff.
			backoff = min(max(backoff*2, s.initialBackoff), s.maxBackoff)

			slog.Error("scrape failed", "error", err, "retry_in", backoff)
			if !wait(backoff) {
				slog.Info("scraper stopping", "reason", ctx.Err())
				return
			}
			continue
		}

		s.cache.store(value)
		slog.Info("scrape succeeded", "seconds", value)
		backoff = 0
		if !wait(s.interval) {
			slog.Info("scraper stopping", "reason", ctx.Err())
			return
		}
	}
}

func (s *scraper) fetch(ctx context.Context) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "doomsday-clock-exporter/"+version)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status %d", resp.StatusCode)
	}

	var page bulletinPage
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return 0, fmt.Errorf("decode: %w", err)
	}

	matches := secondsRe.FindStringSubmatch(page.Excerpt.Rendered)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no match in excerpt: %q", page.Excerpt.Rendered)
	}

	seconds, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("parse %q: %w", matches[1], err)
	}

	return seconds, nil
}

func main() {
	os.Exit(realMain(os.Args[1:]))
}

// config is the exporter runtime configuration, read once from the environment.
type config struct {
	port           string
	scrapeURL      string
	cacheTTL       time.Duration
	scrapeInterval time.Duration
	fetchTimeout   time.Duration
}

// loadConfig reads the DOOMSDAY_CLOCK_EXPORTER_* environment, defaulting anything unset.
func loadConfig() config {
	return config{
		port: envOr("DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT", "8080"),
		scrapeURL: envOr("DOOMSDAY_CLOCK_EXPORTER_SCRAPE_URL",
			"https://thebulletin.org/wp-json/wp/v2/pages/10305"),
		cacheTTL:       time.Duration(envIntOr("DOOMSDAY_CLOCK_EXPORTER_CACHE_TTL", 86400)) * time.Second,
		scrapeInterval: time.Duration(envIntOr("DOOMSDAY_CLOCK_EXPORTER_SCRAPE_INTERVAL", 3600)) * time.Second,
		fetchTimeout:   time.Duration(envIntOr("DOOMSDAY_CLOCK_EXPORTER_FETCH_TIMEOUT", 30)) * time.Second,
	}
}

// newServer builds the HTTP server with the read/write/idle contract.
func newServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// serve waits for a shutdown signal or a fatal server error, drains up to drain, and returns the exit code.
func serve(ctx context.Context, srv *http.Server, serverErr <-chan error, drain time.Duration) int {
	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received, draining HTTP requests and scraper")
		drainCtx, cancel := context.WithTimeout(context.Background(), drain)
		defer cancel()
		if err := srv.Shutdown(drainCtx); err != nil {
			slog.Error("graceful shutdown failed", "error", err)
			return 1
		}
	case err := <-serverErr:
		// ErrServerClosed is expected when Shutdown() stops the server; any
		// other error means the server never served correctly.
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			return 1
		}
	}
	return 0
}

// realMain runs the exporter and returns the process exit code. Keeping
// os.Exit in main (with no preceding defers) satisfies gocritic's
// exitAfterDefer so every defer below runs before the process ends.
func realMain(args []string) int {
	fs := flag.NewFlagSet("doomsday-clock-exporter", flag.ExitOnError)
	showVersion := fs.Bool("version", false, "Print version and exit")
	// ExitOnError prints the usage and exits on a bad flag, so Parse never returns an error.
	_ = fs.Parse(args)

	if *showVersion {
		fmt.Println(version)
		return 0
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel(os.Getenv("B19_VERBOSITY"))})))

	// shutdownCtx is cancelled on SIGINT/SIGTERM so the HTTP server and the
	// background scraper unwind together instead of being killed mid-request.
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := loadConfig()

	slog.Info("starting",
		"port", cfg.port,
		"version", version,
		"cache_ttl", cfg.cacheTTL,
		"scrape_interval", cfg.scrapeInterval,
		"fetch_timeout", cfg.fetchTimeout,
	)

	s := newScraper(cfg)

	// Registering on the default registry exposes the doomsdayclock_* gauges
	// alongside the auto-registered Go runtime/process self-observability
	// metrics (go_*, process_*) served by promhttp.Handler below.
	prometheus.MustRegister(s)
	go s.run(shutdownCtx)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	srv := newServer(":"+cfg.port, mux)

	// Run ListenAndServe in its own goroutine so realMain can select between
	// an incoming shutdown signal and a fatal server error (e.g. port in use).
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.ListenAndServe()
	}()

	return serve(shutdownCtx, srv, serverErr, drainBudget())
}

// logLevel maps B19_VERBOSITY to a slog level; unset or unknown is warn.
func logLevel(verbosity string) slog.Level {
	switch verbosity {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "error":
		return slog.LevelError
	}
	return slog.LevelWarn
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
		slog.Warn("invalid env value, using default", "key", key, "value", v, "default", fallback)
	}
	return fallback
}

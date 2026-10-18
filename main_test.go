// SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>
//
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_model/go"
)

// ninetySecondsExcerpt is the canonical Bulletin wording for the current
// 90-seconds-to-midnight setting; reused across fetch/collect fixtures.
const ninetySecondsExcerpt = "It is now 90 seconds to midnight"

func TestSecondsRe(t *testing.T) {
	cases := []struct {
		input  string
		want   float64
		wantOK bool
	}{
		{ninetySecondsExcerpt, 90, true},
		{"It is now 1 second to midnight", 1, true},
		{"It is now 100 seconds to midnight.", 100, true},
		{"no clock info here", 0, false},
		{"", 0, false},
		{"almost 90 secs to midnight", 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			matches := secondsRe.FindStringSubmatch(tc.input)
			gotOK := len(matches) >= 2
			if gotOK != tc.wantOK {
				t.Fatalf("match=%v want=%v", gotOK, tc.wantOK)
			}
			if gotOK {
				var got float64
				fmt.Sscanf(matches[1], "%f", &got)
				if got != tc.want {
					t.Fatalf("value=%v want=%v", got, tc.want)
				}
			}
		})
	}
}

func TestCacheStoreLoad(t *testing.T) {
	var c cache

	_, _, has := c.load()
	if has {
		t.Fatal("empty cache should report hasValue=false")
	}

	c.store(42.5)
	val, fetchedAt, has := c.load()
	if !has {
		t.Fatal("stored cache should report hasValue=true")
	}
	if val != 42.5 {
		t.Fatalf("value=%v want=42.5", val)
	}
	if fetchedAt.IsZero() {
		t.Fatal("fetchedAt should not be zero")
	}
}

func TestCacheConcurrentAccess(t *testing.T) {
	var c cache
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.store(float64(n))
		}(i)
	}
	wg.Wait()

	val, _, has := c.load()
	if !has {
		t.Fatal("cache should have a value after concurrent writes")
	}
	if val < 0 || val > 99 {
		t.Fatalf("unexpected value=%v", val)
	}
}

func TestFetchSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if accept := r.Header.Get("Accept"); accept != "application/json" {
			t.Errorf("Accept header=%q want=application/json", accept)
		}
		if ua := r.Header.Get("User-Agent"); !strings.HasPrefix(ua, "doomsday-clock-exporter/") {
			t.Errorf("User-Agent=%q want prefix doomsday-clock-exporter/", ua)
		}

		page := bulletinPage{}
		page.Excerpt.Rendered = ninetySecondsExcerpt
		json.NewEncoder(w).Encode(page)
	}))
	defer srv.Close()

	s := &scraper{
		client: srv.Client(),
		url:    srv.URL,
		ttl:    time.Hour,
	}

	val, err := s.fetch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 90 {
		t.Fatalf("value=%v want=90", val)
	}
}

func TestFetchHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := &scraper{
		client: srv.Client(),
		url:    srv.URL,
		ttl:    time.Hour,
	}

	_, err := s.fetch(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestFetchBadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	s := &scraper{
		client: srv.Client(),
		url:    srv.URL,
		ttl:    time.Hour,
	}

	_, err := s.fetch(context.Background())
	if err == nil {
		t.Fatal("expected error for bad JSON")
	}
}

func TestFetchNoMatchInExcerpt(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		page := bulletinPage{}
		page.Excerpt.Rendered = "something unrelated"
		json.NewEncoder(w).Encode(page)
	}))
	defer srv.Close()

	s := &scraper{
		client: srv.Client(),
		url:    srv.URL,
		ttl:    time.Hour,
	}

	_, err := s.fetch(context.Background())
	if err == nil {
		t.Fatal("expected error when excerpt has no clock value")
	}
}

func TestFetchConnectionRefused(t *testing.T) {
	s := &scraper{
		client: &http.Client{Timeout: time.Second},
		url:    "http://127.0.0.1:1",
		ttl:    time.Hour,
	}

	_, err := s.fetch(context.Background())
	if err == nil {
		t.Fatal("expected error for connection refused")
	}
}

func TestCollectNoValue(t *testing.T) {
	s := &scraper{ttl: time.Hour}

	reg := prometheus.NewRegistry()
	reg.Register(s)

	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}

	names := make(map[string]bool)
	for _, mf := range families {
		names[mf.GetName()] = true
	}

	for _, want := range []string{
		metricScrapeSuccess,
		metricCacheStale,
	} {
		if !names[want] {
			t.Errorf("missing metric %q", want)
		}
	}
	if names[metricSecondsToMidnight] {
		t.Error("should not emit seconds metric when no value cached")
	}
}

func TestCollectFreshValue(t *testing.T) {
	s := &scraper{ttl: time.Hour}
	s.cache.store(89)

	reg := prometheus.NewRegistry()
	reg.Register(s)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}

	vals := metricValues(mfs)

	if vals[metricSecondsToMidnight] != 89 {
		t.Errorf("seconds=%v want=89", vals[metricSecondsToMidnight])
	}
	if vals[metricScrapeSuccess] != 1 {
		t.Errorf("success=%v want=1", vals[metricScrapeSuccess])
	}
	if vals[metricCacheStale] != 0 {
		t.Errorf("stale=%v want=0", vals[metricCacheStale])
	}
}

func TestCollectStaleValue(t *testing.T) {
	s := &scraper{ttl: 1 * time.Nanosecond}
	s.cache.store(100)
	// Wait for cache to become stale
	time.Sleep(10 * time.Millisecond)

	reg := prometheus.NewRegistry()
	reg.Register(s)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}

	vals := metricValues(mfs)

	if vals[metricSecondsToMidnight] != 100 {
		t.Errorf("seconds=%v want=100", vals[metricSecondsToMidnight])
	}
	if vals[metricScrapeSuccess] != 0 {
		t.Errorf("success=%v want=0 (stale)", vals[metricScrapeSuccess])
	}
	if vals[metricCacheStale] != 1 {
		t.Errorf("stale=%v want=1", vals[metricCacheStale])
	}
}

func TestDescribe(t *testing.T) {
	s := &scraper{ttl: time.Hour}

	ch := make(chan *prometheus.Desc, 3)
	s.Describe(ch)
	close(ch)

	var descs []*prometheus.Desc
	for d := range ch {
		descs = append(descs, d)
	}
	if len(descs) != 3 {
		t.Fatalf("got %d descs, want 3", len(descs))
	}
}

func TestScraperRegistration(t *testing.T) {
	s := &scraper{ttl: time.Hour}
	reg := prometheus.NewRegistry()

	if err := reg.Register(s); err != nil {
		t.Fatalf("Register: %v", err)
	}
}

func TestEnvOr(t *testing.T) {
	const key = "TEST_DCE_ENVOR"

	os.Unsetenv(key)
	if got := envOr(key, "fallback"); got != "fallback" {
		t.Errorf("unset: got=%q want=fallback", got)
	}

	os.Setenv(key, "custom")
	t.Cleanup(func() { os.Unsetenv(key) })
	if got := envOr(key, "fallback"); got != "custom" {
		t.Errorf("set: got=%q want=custom", got)
	}
}

func TestEnvIntOr(t *testing.T) {
	const key = "TEST_DCE_ENVINTOR"

	os.Unsetenv(key)
	if got := envIntOr(key, 42); got != 42 {
		t.Errorf("unset: got=%d want=42", got)
	}

	os.Setenv(key, "99")
	t.Cleanup(func() { os.Unsetenv(key) })
	if got := envIntOr(key, 42); got != 99 {
		t.Errorf("set: got=%d want=99", got)
	}

	os.Setenv(key, "notanumber")
	if got := envIntOr(key, 42); got != 42 {
		t.Errorf("invalid: got=%d want=42", got)
	}
}

func TestScraperIntegrationWithTestServer(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		page := bulletinPage{}
		switch callCount {
		case 1:
			page.Excerpt.Rendered = ninetySecondsExcerpt
		case 2:
			w.WriteHeader(http.StatusBadGateway)
			return
		case 3:
			page.Excerpt.Rendered = "It is now 100 seconds to midnight"
		default:
			page.Excerpt.Rendered = "It is now 89 seconds to midnight"
		}
		json.NewEncoder(w).Encode(page)
	}))
	defer srv.Close()

	s := &scraper{
		client: srv.Client(),
		url:    srv.URL,
		ttl:    time.Hour,
	}

	val, err := s.fetch(context.Background())
	if err != nil || val != 90 {
		t.Fatalf("first fetch: val=%v err=%v", val, err)
	}

	_, err = s.fetch(context.Background())
	if err == nil {
		t.Fatal("second fetch should fail")
	}

	val, err = s.fetch(context.Background())
	if err != nil || val != 100 {
		t.Fatalf("third fetch: val=%v err=%v", val, err)
	}
}

// TestRunStopsOnContextCancel verifies the scraper loop unwinds promptly when
// its context is cancelled, rather than blocking inside an interval/backoff
// sleep — the property that lets graceful shutdown terminate the goroutine.
func TestRunStopsOnContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		page := bulletinPage{}
		page.Excerpt.Rendered = ninetySecondsExcerpt
		json.NewEncoder(w).Encode(page)
	}))
	defer srv.Close()

	s := &scraper{
		client:   srv.Client(),
		url:      srv.URL,
		ttl:      time.Hour,
		interval: time.Hour,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		s.run(ctx)
		close(done)
	}()

	// Wait for at least one successful scrape to reach the cache, proving the
	// loop has moved into the interval wait before we cancel.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, _, ok := s.cache.load(); ok {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if _, _, ok := s.cache.load(); !ok {
		t.Fatal("no value cached before cancel")
	}

	cancel()
	select {
	case <-done:
		// run() returned promptly after cancellation.
	case <-time.After(2 * time.Second):
		t.Fatal("run() did not stop within 2s of context cancellation")
	}
}

func TestMetricsEndpointIntegration(t *testing.T) {
	s := &scraper{ttl: time.Hour}
	s.cache.store(90)

	reg := prometheus.NewRegistry()
	reg.Register(s)

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d want=200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, metricSecondsToMidnight) {
		t.Error("response missing doomsdayclock_seconds_to_midnight metric")
	}
	if !strings.Contains(body, "90") {
		t.Error("response missing value 90")
	}
}

func TestCollectUsesDesc(t *testing.T) {
	s := &scraper{ttl: time.Hour}
	s.cache.store(90)

	ch := make(chan prometheus.Metric, 10)
	s.Collect(ch)
	close(ch)

	var metrics []prometheus.Metric
	for m := range ch {
		metrics = append(metrics, m)
	}
	if len(metrics) != 3 {
		t.Fatalf("got %d metrics, want 3", len(metrics))
	}

	descs := make(map[string]bool)
	for _, m := range metrics {
		desc := m.Desc()
		// The desc string contains the metric name
		for _, name := range []string{
			metricSecondsToMidnight,
			metricScrapeSuccess,
			metricCacheStale,
		} {
			if strings.Contains(desc.String(), name) {
				descs[name] = true
			}
		}
	}

	for _, name := range []string{
		metricSecondsToMidnight,
		metricScrapeSuccess,
		metricCacheStale,
	} {
		if !descs[name] {
			t.Errorf("missing metric desc for %q", name)
		}
	}
}

func metricValues(mfs []*io_prometheus_client.MetricFamily) map[string]float64 {
	out := make(map[string]float64, len(mfs))
	for _, mf := range mfs {
		for _, m := range mf.GetMetric() {
			out[mf.GetName()] = m.GetGauge().GetValue()
		}
	}
	return out
}

func TestLogLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"":      slog.LevelWarn,
		"bogus": slog.LevelWarn,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
		"info":  slog.LevelInfo,
		"debug": slog.LevelDebug,
	}
	for in, want := range cases {
		if got := logLevel(in); got != want {
			t.Errorf("logLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

// The drain window is a published behaviour, so it is pinned by this test.
func TestDrainBudget(t *testing.T) {
	if got := drainBudget(); got != 10*time.Second {
		t.Errorf("drainBudget()=%v want=10s", got)
	}
}

func TestStaleAt(t *testing.T) {
	fetchedAt := time.Date(2026, time.October, 18, 12, 0, 0, 0, time.UTC)
	ttl := 24 * time.Hour

	cases := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"fresh", fetchedAt, false},
		{"exactly at ttl", fetchedAt.Add(ttl), false},
		{"one nanosecond past ttl", fetchedAt.Add(ttl).Add(time.Nanosecond), true},
		{"well past ttl", fetchedAt.Add(72 * time.Hour), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := staleAt(tc.now, fetchedAt, ttl); got != tc.want {
				t.Errorf("staleAt=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestFetchBuildRequestError(t *testing.T) {
	s := &scraper{
		client: &http.Client{Timeout: time.Second},
		url:    "http://\x7f-invalid/",
		ttl:    time.Hour,
	}

	_, err := s.fetch(context.Background())
	if err == nil {
		t.Fatal("expected error for an unparsable URL")
	}
	if !strings.Contains(err.Error(), "build request") {
		t.Errorf("error=%v want it to name the request build step", err)
	}
}

// A clock value with more digits than float64 can hold must error, not gauge.
func TestFetchValueOutOfFloat64Range(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		page := bulletinPage{}
		page.Excerpt.Rendered = "It is now " + strings.Repeat("9", 400) + " seconds to midnight"
		json.NewEncoder(w).Encode(page)
	}))
	defer srv.Close()

	s := &scraper{client: srv.Client(), url: srv.URL, ttl: time.Hour}

	if _, err := s.fetch(context.Background()); err == nil {
		t.Fatal("expected error for an out-of-range clock value")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	for _, key := range []string{
		"DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT",
		"DOOMSDAY_CLOCK_EXPORTER_SCRAPE_URL",
		"DOOMSDAY_CLOCK_EXPORTER_CACHE_TTL",
		"DOOMSDAY_CLOCK_EXPORTER_SCRAPE_INTERVAL",
		"DOOMSDAY_CLOCK_EXPORTER_FETCH_TIMEOUT",
	} {
		t.Setenv(key, "")
	}

	want := config{
		port:           "8080",
		scrapeURL:      "https://thebulletin.org/wp-json/wp/v2/pages/10305",
		cacheTTL:       86400 * time.Second,
		scrapeInterval: 3600 * time.Second,
		fetchTimeout:   30 * time.Second,
	}
	if got := loadConfig(); got != want {
		t.Errorf("config=%+v want=%+v", got, want)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT", "9999")
	t.Setenv("DOOMSDAY_CLOCK_EXPORTER_SCRAPE_URL", "http://bulletin.test/page/1")
	t.Setenv("DOOMSDAY_CLOCK_EXPORTER_CACHE_TTL", "5")
	t.Setenv("DOOMSDAY_CLOCK_EXPORTER_SCRAPE_INTERVAL", "7")
	t.Setenv("DOOMSDAY_CLOCK_EXPORTER_FETCH_TIMEOUT", "11")

	want := config{
		port:           "9999",
		scrapeURL:      "http://bulletin.test/page/1",
		cacheTTL:       5 * time.Second,
		scrapeInterval: 7 * time.Second,
		fetchTimeout:   11 * time.Second,
	}
	if got := loadConfig(); got != want {
		t.Errorf("config=%+v want=%+v", got, want)
	}
}

func TestLoadConfigRejectsNonNumericDurations(t *testing.T) {
	t.Setenv("DOOMSDAY_CLOCK_EXPORTER_CACHE_TTL", "notanumber")
	t.Setenv("DOOMSDAY_CLOCK_EXPORTER_SCRAPE_INTERVAL", "24h")
	t.Setenv("DOOMSDAY_CLOCK_EXPORTER_FETCH_TIMEOUT", " ")

	cfg := loadConfig()
	if cfg.cacheTTL != 86400*time.Second {
		t.Errorf("cacheTTL=%v want=86400s", cfg.cacheTTL)
	}
	if cfg.scrapeInterval != 3600*time.Second {
		t.Errorf("scrapeInterval=%v want=3600s", cfg.scrapeInterval)
	}
	if cfg.fetchTimeout != 30*time.Second {
		t.Errorf("fetchTimeout=%v want=30s", cfg.fetchTimeout)
	}
}

func TestNewScraper(t *testing.T) {
	cfg := config{
		port:           "8080",
		scrapeURL:      "http://bulletin.test/page/1",
		cacheTTL:       2 * time.Hour,
		scrapeInterval: 3 * time.Hour,
		fetchTimeout:   4 * time.Second,
	}

	s := newScraper(cfg)

	if s.url != cfg.scrapeURL {
		t.Errorf("url=%q want=%q", s.url, cfg.scrapeURL)
	}
	if s.ttl != cfg.cacheTTL {
		t.Errorf("ttl=%v want=%v", s.ttl, cfg.cacheTTL)
	}
	if s.interval != cfg.scrapeInterval {
		t.Errorf("interval=%v want=%v", s.interval, cfg.scrapeInterval)
	}
	if s.client.Timeout != cfg.fetchTimeout {
		t.Errorf("client timeout=%v want=%v", s.client.Timeout, cfg.fetchTimeout)
	}
	if s.initialBackoff != 2*time.Second || s.maxBackoff != 60*time.Second {
		t.Errorf("backoff bounds=(%v,%v) want=(2s,60s)", s.initialBackoff, s.maxBackoff)
	}
}

func TestNewServerTimeouts(t *testing.T) {
	srv := newServer(":1234", http.NotFoundHandler())

	if srv.Addr != ":1234" {
		t.Errorf("addr=%q want=:1234", srv.Addr)
	}
	if srv.Handler == nil {
		t.Error("handler is nil")
	}
	if srv.ReadTimeout != 5*time.Second {
		t.Errorf("read timeout=%v want=5s", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 10*time.Second {
		t.Errorf("write timeout=%v want=10s", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 120*time.Second {
		t.Errorf("idle timeout=%v want=120s", srv.IdleTimeout)
	}
}

// freePort returns a loopback port that was free a moment ago.
func freePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer ln.Close()
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("split %s: %v", ln.Addr(), err)
	}
	return port
}

// waitForListener blocks until addr accepts a TCP connection.
func waitForListener(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("nothing listening on %s", addr)
}

func TestServeDrainsOnShutdown(t *testing.T) {
	addr := "127.0.0.1:" + freePort(t)
	srv := newServer(addr, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "ok")
	}))

	serverErr := make(chan error, 1)
	go func() { serverErr <- srv.ListenAndServe() }()
	waitForListener(t, addr)

	ctx, cancel := context.WithCancel(context.Background())
	code := make(chan int, 1)
	go func() { code <- serve(ctx, srv, serverErr, drainBudget()) }()
	cancel()

	select {
	case got := <-code:
		if got != 0 {
			t.Errorf("exit code=%d want=0", got)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("serve did not return after shutdown")
	}

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			t.Errorf("server error=%v want=%v", err, http.ErrServerClosed)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server did not stop")
	}

	client := &http.Client{Timeout: time.Second}
	if _, err := client.Get("http://" + addr); err == nil {
		t.Error("server still accepts connections after shutdown")
	}
}

func TestServeAcceptsServerClosedAsCleanExit(t *testing.T) {
	serverErr := make(chan error, 1)
	serverErr <- http.ErrServerClosed

	srv := newServer("127.0.0.1:0", http.NotFoundHandler())
	if got := serve(context.Background(), srv, serverErr, drainBudget()); got != 0 {
		t.Errorf("exit code=%d want=0 for ErrServerClosed", got)
	}
}

func TestServeAcceptsClosedErrorChannelAsCleanExit(t *testing.T) {
	serverErr := make(chan error)
	close(serverErr)

	srv := newServer("127.0.0.1:0", http.NotFoundHandler())
	if got := serve(context.Background(), srv, serverErr, drainBudget()); got != 0 {
		t.Errorf("exit code=%d want=0 for a nil server error", got)
	}
}

func TestServeReportsFatalServerError(t *testing.T) {
	serverErr := make(chan error, 1)
	serverErr <- errors.New("listen tcp :8080: bind: address already in use")

	srv := newServer("127.0.0.1:0", http.NotFoundHandler())
	if got := serve(context.Background(), srv, serverErr, drainBudget()); got != 1 {
		t.Errorf("exit code=%d want=1 for a fatal server error", got)
	}
}

// A request still in flight when the drain window closes fails the shutdown.
func TestServeFailsWhenDrainExpires(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{})
	addr := "127.0.0.1:" + freePort(t)

	srv := newServer(addr, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		fmt.Fprint(w, "late")
	}))
	defer srv.Close()

	serverErr := make(chan error, 1)
	go func() { serverErr <- srv.ListenAndServe() }()
	waitForListener(t, addr)

	client := &http.Client{Timeout: 10 * time.Second}
	go client.Get("http://" + addr)
	<-started

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got := serve(ctx, srv, serverErr, 20*time.Millisecond); got != 1 {
		t.Errorf("exit code=%d want=1 when the drain window expires", got)
	}

	close(release)
}

func TestRealMainPrintsVersionAndExits(t *testing.T) {
	code := 0
	out := captureStdout(t, func() { code = realMain([]string{"-version"}) })

	if code != 0 {
		t.Errorf("exit code=%d want=0", code)
	}
	if out != version+"\n" {
		t.Errorf("stdout=%q want=%q", out, version+"\n")
	}
}

// TestRealMainServesUntilSignal is the end-to-end path a container runs.
func TestRealMainServesUntilSignal(t *testing.T) {
	bulletin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		page := bulletinPage{}
		page.Excerpt.Rendered = ninetySecondsExcerpt
		json.NewEncoder(w).Encode(page)
	}))
	defer bulletin.Close()

	port := freePort(t)
	t.Setenv("DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT", port)
	t.Setenv("DOOMSDAY_CLOCK_EXPORTER_SCRAPE_URL", bulletin.URL)
	t.Setenv("DOOMSDAY_CLOCK_EXPORTER_CACHE_TTL", "60")
	t.Setenv("DOOMSDAY_CLOCK_EXPORTER_SCRAPE_INTERVAL", "3600")
	t.Setenv("DOOMSDAY_CLOCK_EXPORTER_FETCH_TIMEOUT", "5")

	restoreDefaultLogger := slog.Default()
	t.Cleanup(func() { slog.SetDefault(restoreDefaultLogger) })

	code := make(chan int, 1)
	go func() { code <- realMain(nil) }()

	base := "http://127.0.0.1:" + port
	client := &http.Client{Timeout: 5 * time.Second}

	if body := waitForBody(t, client, base+"/health", "ok"); body != "ok" {
		t.Errorf("/health body=%q want=ok", body)
	}

	body := waitForBody(t, client, base+"/metrics", metricSecondsToMidnight)
	for _, want := range []string{metricSecondsToMidnight, metricScrapeSuccess, metricCacheStale, "90"} {
		if !strings.Contains(body, want) {
			t.Errorf("/metrics missing %q", want)
		}
	}

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("signal self: %v", err)
	}

	select {
	case got := <-code:
		if got != 0 {
			t.Errorf("exit code=%d want=0", got)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("realMain did not exit on SIGTERM")
	}
}

// waitForBody polls url until the response body contains want.
func waitForBody(t *testing.T, client *http.Client, url, want string) string {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	var last string
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			last = string(body)
			if strings.Contains(last, want) {
				return last
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("%s never served %q (last body: %q)", url, want, last)
	return ""
}

// captureStdout returns everything fn wrote to os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	orig := os.Stdout
	os.Stdout = w
	fn()
	os.Stdout = orig
	w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	r.Close()
	return buf.String()
}

// hitRecorder timestamps upstream requests so the retry cadence can be asserted.
type hitRecorder struct {
	mu    sync.Mutex
	hits  []time.Time
	calls atomic.Int64
}

func (h *hitRecorder) record() int64 {
	h.mu.Lock()
	h.hits = append(h.hits, time.Now())
	h.mu.Unlock()
	return h.calls.Add(1)
}

func (h *hitRecorder) count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.hits)
}

// gaps returns the n-1 intervals between the first n recorded hits.
func (h *hitRecorder) gaps(n int) []time.Duration {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]time.Duration, 0, n-1)
	for i := 1; i < n && i < len(h.hits); i++ {
		out = append(out, h.hits[i].Sub(h.hits[i-1]))
	}
	return out
}

// runUntilHits runs s until the upstream has answered n times, then stops it.
func runUntilHits(t *testing.T, s *scraper, rec *hitRecorder, n int) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { defer close(done); s.run(ctx) }()

	deadline := time.After(30 * time.Second)
	for rec.count() < n {
		select {
		case <-deadline:
			cancel()
			<-done
			t.Fatalf("run() reached %d scrapes, want %d", rec.count(), n)
		case <-time.After(2 * time.Millisecond):
		}
	}
	cancel()
	<-done
}

// TestRunBackoffGrowsAndResets scripts fail, fail, ok, fail so each gap shows the policy.
func TestRunBackoffGrowsAndResets(t *testing.T) {
	const (
		initial = 30 * time.Millisecond
		maximum = 400 * time.Millisecond
		every   = 60 * time.Millisecond
	)

	rec := &hitRecorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if n := rec.record(); n == 1 || n == 2 || n == 4 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		page := bulletinPage{}
		page.Excerpt.Rendered = ninetySecondsExcerpt
		json.NewEncoder(w).Encode(page)
	}))
	defer srv.Close()

	s := &scraper{
		client:         srv.Client(),
		url:            srv.URL,
		ttl:            time.Hour,
		interval:       every,
		initialBackoff: initial,
		maxBackoff:     maximum,
	}

	runUntilHits(t, s, rec, 5)
	gaps := rec.gaps(5)

	if len(gaps) != 4 {
		t.Fatalf("got %d gaps, want 4", len(gaps))
	}
	if gaps[0] < 20*time.Millisecond {
		t.Errorf("gap before the first retry=%v, want at least the initial backoff", gaps[0])
	}
	if gaps[1] < 45*time.Millisecond {
		t.Errorf("gap before the second retry=%v, want a doubled backoff", gaps[1])
	}
	if gaps[3] < 20*time.Millisecond {
		t.Errorf("gap after a success=%v, want the backoff reset to the initial value", gaps[3])
	}
	if gaps[3] > 4*initial {
		t.Errorf("gap after a success=%v, want the backoff reset and not grown", gaps[3])
	}
}

// TestRunBackoffIsCappedAtMax keeps a dead upstream from growing the delay without bound.
func TestRunBackoffIsCappedAtMax(t *testing.T) {
	const (
		initial = 20 * time.Millisecond
		maximum = 30 * time.Millisecond
	)

	rec := &hitRecorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		rec.record()
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	s := &scraper{
		client:         srv.Client(),
		url:            srv.URL,
		ttl:            time.Hour,
		interval:       time.Hour,
		initialBackoff: initial,
		maxBackoff:     maximum,
	}

	runUntilHits(t, s, rec, 5)
	gaps := rec.gaps(5)

	if len(gaps) != 4 {
		t.Fatalf("got %d gaps, want 4", len(gaps))
	}
	if gaps[0] < 15*time.Millisecond {
		t.Errorf("gap before the first retry=%v, want at least the initial backoff", gaps[0])
	}
	// Uncapped, the fourth gap would already be eight times the initial backoff.
	if last := gaps[3]; last > 2*maximum {
		t.Errorf("gap between the last retries=%v, want no more than the %v ceiling", last, maximum)
	}
}

// TestRunStopsDuringBackoffWait proves cancellation interrupts the retry sleep.
func TestRunStopsDuringBackoffWait(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	s := &scraper{
		client:         srv.Client(),
		url:            srv.URL,
		ttl:            time.Hour,
		interval:       time.Hour,
		initialBackoff: 30 * time.Second,
		maxBackoff:     time.Minute,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { defer close(done); s.run(ctx) }()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("run() did not stop while waiting out a backoff")
	}

	if _, _, has := s.cache.load(); has {
		t.Error("a failing upstream must not populate the cache")
	}
}

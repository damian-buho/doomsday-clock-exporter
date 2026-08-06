// SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
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
	const key = "TEST_O9S_DCE_ENVOR"

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
	const key = "TEST_O9S_DCE_ENVINTOR"

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

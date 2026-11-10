// SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// testListPage is the fixture title of the page holding the launch authority table.
const testListPage = "Nukes"

// launchTableFixture mirrors the shape of the live Launch authority table, refs and multi-line rows included.
const launchTableFixture = `== Launch authority ==
Some prose.<ref name="UCS" />

{| class="wikitable"
|+Launch authority of nuclear states
|-
! scope="col" width=130px | Country
! scope="col"| Authority
! scope="col"| Notes
|-
| scope="row"| {{flagicon|United States}} [[Nuclear weapons of the United States|United States]]
| [[President of the United States]]|| See the [[Nuclear football|Presidential Emergency Satchel]].<ref name="UCS" />
|-
| scope="row"| {{flagicon|Pakistan}} [[Pakistan and weapons of mass destruction|Pakistan]]
| [[National Command Authority (Pakistan)|National Command Authority]]|| Requires a consensus.<ref name="UCS" />
|-
| scope="row"| {{flagicon|North Korea}} [[North Korea and weapons of mass destruction|North Korea]]
| [[President of the State Affairs Commission|President of the State Affairs]]|| The ultimate decisionmaker.
|}

== Nuclear weapons sharing ==
{| class="wikitable"
|-
| scope="row"| [[Belgium]] || [[King of the Belgians]]
|}
`

// wikiFixture is a minimal Wikipedia: the list page, three offices and three people.
func wikiFixture() map[string]string {
	return map[string]string{
		testListPage:                                launchTableFixture,
		"President of the United States":            "{{Infobox official post\n| incumbent = [[Donald Trump]]\n}}",
		"National Command Authority (Pakistan)":     "{{Infobox government agency\n| chief1_name = [[Lieutenant general (Pakistan)|Lt-Gen.]] [[Yusuf Jamal]]\n}}\nWith the [[Prime Minister of Pakistan]] as its chairperson, the NCA acts.",
		"Prime Minister of Pakistan":                "| incumbent = [[Shehbaz Sharif]]",
		"President of the State Affairs Commission": "| incumbent = [[Kim Jong Un]] ''(since 2016)''",
		"Donald Trump":                              "| birth_date      = {{birth date and age|1946|6|14}}",
		"Shehbaz Sharif":                            "| birth_date         = {{birth date and age|df=y|1951|9|23}}",
		"Kim Jong Un":                               "| birth_date         = 8 January {{circa|1982–1984}} (age {{age|1984|1|8}}–{{age|1982|1|8}})",
	}
}

// fakeWiki serves pages as the MediaWiki query API does, answering missing for unknown titles.
func fakeWiki(t *testing.T, pages map[string]string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Query().Get("titles")
		if ua := r.Header.Get("User-Agent"); !strings.HasPrefix(ua, "doomsday-clock-exporter/") {
			t.Errorf("User-Agent=%q want prefix doomsday-clock-exporter/", ua)
		}
		page := map[string]any{"title": title, "missing": true}
		if text, ok := pages[title]; ok {
			page = map[string]any{"title": title, "revisions": []any{map[string]any{"slots": map[string]any{"main": map[string]any{"content": text}}}}}
		}
		json.NewEncoder(w).Encode(map[string]any{"query": map[string]any{"pages": []any{page}}})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func testLeaders(srv *httptest.Server) *leaders {
	return &leaders{client: srv.Client(), api: srv.URL, page: testListPage}
}

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestLaunchAuthorities(t *testing.T) {
	got, err := launchAuthorities(launchTableFixture)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []authority{
		{"United States", "President of the United States"},
		{"Pakistan", "National Command Authority (Pakistan)"},
		{"North Korea", "President of the State Affairs Commission"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rows %+v want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d=%+v want=%+v", i, got[i], want[i])
		}
	}
}

func TestLaunchAuthoritiesRejectsMissingOrEmptyTable(t *testing.T) {
	for name, text := range map[string]string{
		"no heading":  "{|\n|-\n| scope=\"row\"| [[A]] || [[B]]\n|}",
		"no rows":     "== Launch authority ==\n{|\n|-\n! Country\n|}",
		"no links":    "== Launch authority ==\n{|\n|-\n| scope=\"row\"| A || B\n|}",
		"single cell": "== Launch authority ==\n{|\n|-\n| scope=\"row\"| [[A]]\n|}",
	} {
		if _, err := launchAuthorities(text); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

func TestBirthDate(t *testing.T) {
	for name, tc := range map[string]struct {
		text string
		want time.Time
	}{
		"plain":       {"| birth_date = {{birth date and age|1946|6|14}}", date(1946, 6, 14)},
		"df first":    {"| birth_date = {{birth date and age|df=yes|1952|10|7}}", date(1952, 10, 7)},
		"df last":     {"| birth_date = {{birth date and age|1970|1|7|df=y}}", date(1970, 1, 7)},
		"zero padded": {"| birth_date = {{Birth date and age|df=yes|1953|06|15}}", date(1953, 6, 15)},
		"bda":         {"| birth_date = {{bda|1977|12|21}}", date(1977, 12, 21)},
		"range":       {wikiFixture()["Kim Jong Un"], date(1983, 1, 8)},
	} {
		got, err := birthDate(tc.text)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", name, err)
			continue
		}
		if !got.Equal(tc.want) {
			t.Errorf("%s: got %s want %s", name, got.Format(time.DateOnly), tc.want.Format(time.DateOnly))
		}
	}
}

func TestBirthDateRejectsMissingDate(t *testing.T) {
	for name, text := range map[string]string{
		"no field":     "| death_date = {{death date|2000|1|1}}",
		"no template":  "| birth_date = 1950s",
		"too few args": "| birth_date = {{birth date|1950|5}}",
	} {
		if _, err := birthDate(text); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}

func TestAgeYearsAroundBirthday(t *testing.T) {
	birth := date(1980, 3, 10)
	if got := ageYears(birth, date(2020, 3, 10)); got != 40 {
		t.Errorf("on birthday=%v want=40", got)
	}
	if got := ageYears(birth, date(2020, 3, 9)); got >= 40 || got < 39.99 {
		t.Errorf("day before=%v want just under 40", got)
	}
	if got := ageYears(birth, date(2020, 3, 11)); got <= 40 || got > 40.01 {
		t.Errorf("day after=%v want just over 40", got)
	}
}

func TestAgeYearsLeapDayBirth(t *testing.T) {
	if got := ageYears(date(2000, 2, 29), date(2001, 3, 1)); got != 1 {
		t.Errorf("leap birth on 1 March=%v want=1", got)
	}
}

func TestAverageAge(t *testing.T) {
	now := date(2020, 6, 1)
	if got := averageAge([]time.Time{date(1980, 6, 1), date(1960, 6, 1)}, now); got != 50 {
		t.Errorf("average=%v want=50", got)
	}
}

func TestMedianAge(t *testing.T) {
	now := date(2020, 6, 1)
	odd := []time.Time{date(1940, 6, 1), date(1980, 6, 1), date(1970, 6, 1)}
	if got := medianAge(odd, now); got != 50 {
		t.Errorf("odd median=%v want=50", got)
	}
	even := []time.Time{date(1940, 6, 1), date(2000, 6, 1), date(1980, 6, 1), date(1970, 6, 1)}
	if got := medianAge(even, now); got != 45 {
		t.Errorf("even median=%v want=45", got)
	}
	if got := medianAge(odd[:1], now); got != 80 {
		t.Errorf("single median=%v want=80", got)
	}
}

func TestYearsAndDays(t *testing.T) {
	if y, d := yearsAndDays(70.5); y != 70 || d != 182 {
		t.Errorf("70.5 -> %dy %dd want 70y 182d", y, d)
	}
	if y, d := yearsAndDays(0); y != 0 || d != 0 {
		t.Errorf("0 -> %dy %dd want 0y 0d", y, d)
	}
}

func TestChairWindowEdge(t *testing.T) {
	fill := func(n int) string { return "[[Office]]" + strings.Repeat("x", n) + " chair" }
	if !chairRe.MatchString(fill(39)) {
		t.Error("40-char gap should match")
	}
	if chairRe.MatchString(fill(40)) {
		t.Error("41-char gap should not match")
	}
}

func TestLeadersFetchFollowsTheChain(t *testing.T) {
	births, err := testLeaders(fakeWiki(t, wikiFixture())).fetch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []time.Time{date(1946, 6, 14), date(1951, 9, 23), date(1983, 1, 8)}
	if len(births) != len(want) {
		t.Fatalf("got %d births want %d", len(births), len(want))
	}
	for i := range want {
		if !births[i].Equal(want[i]) {
			t.Errorf("birth %d=%s want=%s", i, births[i].Format(time.DateOnly), want[i].Format(time.DateOnly))
		}
	}
}

// chairChain builds offices O0…On-1 each chaired by the next, the last naming a person.
func chairChain(n int) map[string]string {
	pages := map[string]string{"Person": "| birth_date = {{bda|1950|1|1}}"}
	for i := range n - 1 {
		pages["O"+string(rune('0'+i))] = "The [[O" + string(rune('1'+i)) + "]] is its chair."
	}
	pages["O"+string(rune('0'+n-1))] = "| incumbent = [[Person]]"
	return pages
}

func TestIncumbentHopLimitEdge(t *testing.T) {
	if got, err := testLeaders(fakeWiki(t, chairChain(maxOfficeHops))).incumbent(context.Background(), "O0"); err != nil || got != "Person" {
		t.Errorf("%d offices: got %q err %v want Person", maxOfficeHops, got, err)
	}
	if _, err := testLeaders(fakeWiki(t, chairChain(maxOfficeHops+1))).incumbent(context.Background(), "O0"); err == nil {
		t.Errorf("%d offices: want error", maxOfficeHops+1)
	}
}

func TestLeadersFetchFailsWhenAnyLinkBreaks(t *testing.T) {
	for _, drop := range []string{testListPage, "Prime Minister of Pakistan", "Kim Jong Un"} {
		pages := wikiFixture()
		delete(pages, drop)
		if _, err := testLeaders(fakeWiki(t, pages)).fetch(context.Background()); err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("without %q: err=%v want not found", drop, err)
		}
	}
	pages := wikiFixture()
	pages["Donald Trump"] = "no infobox"
	if _, err := testLeaders(fakeWiki(t, pages)).fetch(context.Background()); err == nil {
		t.Error("person without birth_date: want error")
	}
	pages = wikiFixture()
	pages["President of the United States"] = "vacant"
	if _, err := testLeaders(fakeWiki(t, pages)).fetch(context.Background()); err == nil {
		t.Error("office without incumbent or chair: want error")
	}
	pages = wikiFixture()
	pages[testListPage] = "no table"
	if _, err := testLeaders(fakeWiki(t, pages)).fetch(context.Background()); err == nil {
		t.Error("list page without table: want error")
	}
}

func TestWikitextTransportErrors(t *testing.T) {
	for name, h := range map[string]http.HandlerFunc{
		"status": func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTooManyRequests) },
		"json":   func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("{")) },
		"empty":  func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte(`{"query":{"pages":[]}}`)) },
	} {
		srv := httptest.NewServer(h)
		if _, err := testLeaders(srv).wikitext(context.Background(), "X"); err == nil {
			t.Errorf("%s: want error", name)
		}
		srv.Close()
	}
	if _, err := (&leaders{client: http.DefaultClient, api: "http://127.0.0.1:1"}).wikitext(context.Background(), "X"); err == nil {
		t.Error("refused: want error")
	}
	if _, err := (&leaders{client: http.DefaultClient, api: "://bad"}).wikitext(context.Background(), "X"); err == nil {
		t.Error("bad url: want error")
	}
}

func TestLeadersCollect(t *testing.T) {
	l := &leaders{}
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(l)
	if mfs, err := reg.Gather(); err != nil || len(mfs) != 0 {
		t.Fatalf("no data: got %d families err %v want none", len(mfs), err)
	}
	l.births = []time.Time{time.Now().AddDate(-60, 0, 0)}
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, name := range []string{metricLeadersAverageAge, metricLeadersMedianAge} {
		if got := metricValues(mfs)[name]; math.Abs(got-60) > 0.01 {
			t.Errorf("%s=%v want≈60", name, got)
		}
	}
}

func TestLeadersRunStoresBirthsAndStops(t *testing.T) {
	l := testLeaders(fakeWiki(t, wikiFixture()))
	l.interval, l.initialBackoff, l.maxBackoff = time.Hour, time.Millisecond, time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); l.run(ctx) }()
	deadline := time.Now().Add(30 * time.Second)
	for {
		l.mu.RLock()
		n := len(l.births)
		l.mu.RUnlock()
		if n == 3 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("births never stored, have %d", n)
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	<-done
}

func TestNewLeaders(t *testing.T) {
	cfg := config{leadersURL: "http://wiki.test/api.php", leadersPage: "P", scrapeInterval: 7 * time.Second, fetchTimeout: 11 * time.Second}
	l := newLeaders(cfg)
	if l.api != cfg.leadersURL || l.page != cfg.leadersPage || l.interval != cfg.scrapeInterval || l.client.Timeout != cfg.fetchTimeout {
		t.Errorf("leaders=%+v does not mirror config %+v", l, cfg)
	}
}

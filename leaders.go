// SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// leadersDisabled is the DOOMSDAY_CLOCK_EXPORTER_LEADERS_URL value that turns the source off.
const leadersDisabled = "off"

// maxOfficeHops bounds how many office pages are followed to reach a person.
const maxOfficeHops = 3

const metricLeadersAverageAge = "doomsdayclock_nuclear_leaders_average_age_years"

var leadersAgeDesc = prometheus.NewDesc(
	metricLeadersAverageAge,
	"Average age in years of the people holding nuclear launch authority, as listed on Wikipedia",
	nil, nil,
)

var (
	// launchTableRe captures the body of the first table under the Launch authority heading.
	launchTableRe = regexp.MustCompile(`(?s)==\s*Launch authority\s*==.*?\{\|(.*?)\n\|\}`)
	// cellSplitRe splits a wikitable row into cells on line-leading and inline separators.
	cellSplitRe = regexp.MustCompile(`\n\||\|\|`)
	// linkRe matches a wikilink, capturing its target and optional label.
	linkRe = regexp.MustCompile(`\[\[([^\]|#]+)(?:#[^\]|]*)?(?:\|([^\]]*))?\]\]`)
	// incumbentRe captures the value of an infobox incumbent field.
	incumbentRe = regexp.MustCompile(`(?m)^\s*\|\s*incumbent\s*=(.*)$`)
	// chairRe captures a link described as the chair of the body within the same clause.
	chairRe = regexp.MustCompile(`\[\[([^\]|#]+)[^\]]*\]\][^.\n\[]{0,40}\bchair`)
	// birthFieldRe captures the value of an infobox birth_date field.
	birthFieldRe = regexp.MustCompile(`(?m)^\s*\|\s*birth_date\s*=(.*)$`)
	// dateTemplateRe captures the parameters of a birth date or age template.
	dateTemplateRe = regexp.MustCompile(`(?i)\{\{\s*(?:birth date and age|birth date|bda|age)\s*\|([^{}]*)\}\}`)
)

// authority is one row of the launch authority table.
type authority struct {
	country string
	office  string
}

// wikiResponse is the slice of a MediaWiki query API answer the source reads.
type wikiResponse struct {
	Query struct {
		Pages []struct {
			Title     string `json:"title"`
			Missing   bool   `json:"missing"`
			Revisions []struct {
				Slots struct {
					Main struct {
						Content string `json:"content"`
					} `json:"main"`
				} `json:"slots"`
			} `json:"revisions"`
		} `json:"pages"`
	} `json:"query"`
}

// leaders follows Wikipedia from the launch authority table to each incumbent’s birth date.
type leaders struct {
	client         *http.Client
	api            string
	page           string
	interval       time.Duration
	initialBackoff time.Duration
	maxBackoff     time.Duration
	mu             sync.RWMutex
	births         []time.Time
}

// newLeaders builds the production Wikipedia source with the Bulletin scraper’s cadence.
func newLeaders(cfg config) *leaders {
	return &leaders{
		client:         &http.Client{Timeout: cfg.fetchTimeout},
		api:            cfg.leadersURL,
		page:           cfg.leadersPage,
		interval:       cfg.scrapeInterval,
		initialBackoff: 2 * time.Second,
		maxBackoff:     60 * time.Second,
	}
}

func (l *leaders) Describe(ch chan<- *prometheus.Desc) {
	ch <- leadersAgeDesc
}

func (l *leaders) Collect(ch chan<- prometheus.Metric) {
	l.mu.RLock()
	births := l.births
	l.mu.RUnlock()
	if len(births) == 0 {
		return
	}
	ch <- prometheus.MustNewConstMetric(leadersAgeDesc, prometheus.GaugeValue, averageAge(births, time.Now()))
}

// run refreshes the birth dates every interval until ctx is cancelled.
func (l *leaders) run(ctx context.Context) {
	retryLoop(ctx, "wikipedia", l.interval, l.initialBackoff, l.maxBackoff, func(ctx context.Context) error {
		births, err := l.fetch(ctx)
		if err != nil {
			return err
		}
		l.mu.Lock()
		l.births = births
		l.mu.Unlock()
		years, days := yearsAndDays(averageAge(births, time.Now()))
		slog.Info("scrape succeeded", "source", "wikipedia", "leaders", len(births), "average_years", years, "average_days", days)
		return nil
	})
}

// fetch resolves every launch authority to a person and returns their birth dates.
func (l *leaders) fetch(ctx context.Context) ([]time.Time, error) {
	text, err := l.wikitext(ctx, l.page)
	if err != nil {
		return nil, err
	}
	authorities, err := launchAuthorities(text)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", l.page, err)
	}
	births := make([]time.Time, 0, len(authorities))
	for _, a := range authorities {
		person, err := l.incumbent(ctx, a.office)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", a.country, err)
		}
		bio, err := l.wikitext(ctx, person)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", a.country, err)
		}
		birth, err := birthDate(bio)
		if err != nil {
			return nil, fmt.Errorf("%s: %s: %w", a.country, person, err)
		}
		slog.Debug("leader resolved", "country", a.country, "office", a.office, "person", person, "birth", birth.Format(time.DateOnly))
		births = append(births, birth)
	}
	return births, nil
}

// incumbent follows office pages until one names its current holder.
func (l *leaders) incumbent(ctx context.Context, office string) (string, error) {
	for hop := range maxOfficeHops {
		text, err := l.wikitext(ctx, office)
		if err != nil {
			return "", err
		}
		if m := incumbentRe.FindStringSubmatch(text); m != nil {
			if person, _, ok := firstLink(m[1]); ok {
				slog.Debug("incumbent found", "office", office, "person", person, "hop", hop)
				return person, nil
			}
		}
		m := chairRe.FindStringSubmatch(text)
		if m == nil {
			return "", fmt.Errorf("%s: no incumbent and no chair", office)
		}
		slog.Debug("office chaired by another office", "office", office, "chair", m[1], "hop", hop)
		office = m[1]
	}
	return "", fmt.Errorf("%s: no incumbent within %d hops", office, maxOfficeHops)
}

// wikitext returns the current wikitext of a page, following redirects.
func (l *leaders) wikitext(ctx context.Context, title string) (string, error) {
	q := url.Values{
		"action": {"query"}, "prop": {"revisions"}, "rvprop": {"content"}, "rvslots": {"main"},
		"format": {"json"}, "formatversion": {"2"}, "redirects": {"1"}, "titles": {title},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l.api+"?"+q.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "doomsday-clock-exporter/"+version+" (https://kiota.ch/damian-buho/doomsday-clock-exporter)")
	req.Header.Set("Accept", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", title, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch %s: status %d", title, resp.StatusCode)
	}

	var w wikiResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return "", fmt.Errorf("decode %s: %w", title, err)
	}
	if len(w.Query.Pages) == 0 || w.Query.Pages[0].Missing || len(w.Query.Pages[0].Revisions) == 0 {
		return "", fmt.Errorf("page %q not found", title)
	}
	return w.Query.Pages[0].Revisions[0].Slots.Main.Content, nil
}

// launchAuthorities reads the country and authority link of every row of the launch authority table.
func launchAuthorities(text string) ([]authority, error) {
	m := launchTableRe.FindStringSubmatch(text)
	if m == nil {
		return nil, errors.New("no launch authority table")
	}
	var out []authority
	for _, row := range strings.Split(m[1], "\n|-") {
		if !strings.Contains(row, `scope="row"`) {
			continue
		}
		cells := cellSplitRe.Split(row, -1)
		if len(cells) < 3 {
			slog.Warn("launch authority row skipped", "cells", len(cells), "row", row)
			continue
		}
		_, country, okCountry := firstLink(cells[1])
		office, _, okOffice := firstLink(cells[2])
		if !okCountry || !okOffice {
			slog.Warn("launch authority row without links", "row", row)
			continue
		}
		out = append(out, authority{country: country, office: office})
	}
	if len(out) == 0 {
		return nil, errors.New("launch authority table has no rows")
	}
	return out, nil
}

// firstLink returns the target and the shown label of the first wikilink in s.
func firstLink(s string) (target, label string, ok bool) {
	m := linkRe.FindStringSubmatch(s)
	if m == nil {
		return "", "", false
	}
	target = strings.TrimSpace(m[1])
	if label = strings.TrimSpace(m[2]); label == "" {
		label = target
	}
	return target, label, true
}

// birthDate reads the infobox birth date, taking the midpoint when the field gives a range.
func birthDate(text string) (time.Time, error) {
	field := birthFieldRe.FindStringSubmatch(text)
	if field == nil {
		return time.Time{}, errors.New("no birth_date field")
	}
	var earliest, latest time.Time
	for _, tpl := range dateTemplateRe.FindAllStringSubmatch(field[1], -1) {
		d, ok := templateDate(tpl[1])
		if !ok {
			continue
		}
		if earliest.IsZero() || d.Before(earliest) {
			earliest = d
		}
		if latest.IsZero() || d.After(latest) {
			latest = d
		}
	}
	if earliest.IsZero() {
		return time.Time{}, fmt.Errorf("no date in birth_date %q", strings.TrimSpace(field[1]))
	}
	return earliest.Add(latest.Sub(earliest) / 2).Truncate(24 * time.Hour), nil
}

// templateDate builds a date from the first three numeric positional template parameters.
func templateDate(params string) (time.Time, bool) {
	var ymd []int
	for _, p := range strings.Split(params, "|") {
		if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
			ymd = append(ymd, n)
		}
	}
	if len(ymd) < 3 {
		return time.Time{}, false
	}
	return time.Date(ymd[0], time.Month(ymd[1]), ymd[2], 0, 0, 0, 0, time.UTC), true
}

// ageYears is the age at now in whole years plus the elapsed fraction of the current birthday year.
func ageYears(birth, now time.Time) float64 {
	years := now.Year() - birth.Year()
	if birth.AddDate(years, 0, 0).After(now) {
		years--
	}
	last, next := birth.AddDate(years, 0, 0), birth.AddDate(years+1, 0, 0)
	return float64(years) + float64(now.Sub(last))/float64(next.Sub(last))
}

// averageAge is the mean ageYears of births at now.
func averageAge(births []time.Time, now time.Time) float64 {
	var sum float64
	for _, b := range births {
		sum += ageYears(b, now)
	}
	return sum / float64(len(births))
}

// yearsAndDays splits a fractional age into whole years and whole days of a 365.2425-day year.
func yearsAndDays(age float64) (int, int) {
	years := int(age)
	return years, int((age - float64(years)) * 365.2425)
}

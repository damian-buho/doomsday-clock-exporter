<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

# Cached scraping with graceful degradation

- A background scraper fetches the Doomsday Clock value from the Bulletin of Atomic Scientists at configurable intervals.
- Scraped values are cached with a configurable TTL; Prometheus scrapes always return the cached value without blocking.
- Serves stale values with a degradation gauge (`doomsdayclock_cache_stale=1`) when the upstream source is unavailable.
- Scrape failures trigger exponential backoff with retry, up to a 60-second ceiling.

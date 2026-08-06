<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

# Environment variable configuration

- HTTP listen port: `O9S_DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT` (default `8080`).
- Upstream scrape URL: `O9S_DOOMSDAY_CLOCK_EXPORTER_SCRAPE_URL` (defaults to the Bulletin’s REST API).
- Upstream fetch timeout: `O9S_DOOMSDAY_CLOCK_EXPORTER_FETCH_TIMEOUT` (default `30` seconds).
- Scrape interval: `O9S_DOOMSDAY_CLOCK_EXPORTER_SCRAPE_INTERVAL` (default `3600` seconds).
- Cache TTL: `O9S_DOOMSDAY_CLOCK_EXPORTER_CACHE_TTL` (default `86400` seconds).

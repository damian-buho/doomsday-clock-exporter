<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

# Environment variable configuration

- HTTP listen port: `DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT` (default `8080`).
- Upstream scrape URL: `DOOMSDAY_CLOCK_EXPORTER_SCRAPE_URL` (defaults to the Bulletin’s REST API).
- Upstream fetch timeout: `DOOMSDAY_CLOCK_EXPORTER_FETCH_TIMEOUT` (default `30` seconds).
- Scrape interval: `DOOMSDAY_CLOCK_EXPORTER_SCRAPE_INTERVAL` (default `3600` seconds).
- Cache TTL: `DOOMSDAY_CLOCK_EXPORTER_CACHE_TTL` (default `86400` seconds).
- Wikipedia API for the nuclear leaders’ age: `DOOMSDAY_CLOCK_EXPORTER_LEADERS_URL` (default `https://en.wikipedia.org/w/api.php`, `off` disables).
- Wikipedia page holding the launch authority table: `DOOMSDAY_CLOCK_EXPORTER_LEADERS_PAGE` (default `List of states with nuclear weapons`).

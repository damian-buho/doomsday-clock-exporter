<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

# Налаштування змінними середовища

- Порт прослуховування HTTP: `O9S_DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT` (типово `8080`).
- URL збору з upstream: `O9S_DOOMSDAY_CLOCK_EXPORTER_SCRAPE_URL` (типово REST API Bulletin).
- Таймаут звернення до upstream: `O9S_DOOMSDAY_CLOCK_EXPORTER_FETCH_TIMEOUT` (типово `30` секунд).
- Інтервал збору: `O9S_DOOMSDAY_CLOCK_EXPORTER_SCRAPE_INTERVAL` (типово `3600` секунд).
- TTL кешу: `O9S_DOOMSDAY_CLOCK_EXPORTER_CACHE_TTL` (типово `86400` секунд).

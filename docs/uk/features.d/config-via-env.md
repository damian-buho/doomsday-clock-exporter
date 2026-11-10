<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

<!-- textlint-disable terminology,common-misspellings -->

# Налаштування змінними середовища

- Порт прослуховування HTTP: `DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT` (типово `8080`).
- URL збору з upstream: `DOOMSDAY_CLOCK_EXPORTER_SCRAPE_URL` (типово REST API Bulletin).
- Таймаут звернення до upstream: `DOOMSDAY_CLOCK_EXPORTER_FETCH_TIMEOUT` (типово `30` секунд).
- Інтервал збору: `DOOMSDAY_CLOCK_EXPORTER_SCRAPE_INTERVAL` (типово `3600` секунд).
- TTL кешу: `DOOMSDAY_CLOCK_EXPORTER_CACHE_TTL` (типово `86400` секунд).
- API Вікіпедії для віку ядерних лідерів: `DOOMSDAY_CLOCK_EXPORTER_LEADERS_URL` (типово `https://en.wikipedia.org/w/api.php`, `off` вимикає).
- Сторінка Вікіпедії з таблицею повноважень на пуск: `DOOMSDAY_CLOCK_EXPORTER_LEADERS_PAGE` (типово `List of states with nuclear weapons`).

<!-- textlint-enable -->

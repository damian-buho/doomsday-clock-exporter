<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

<!-- textlint-disable terminology,common-misspellings -->

# Configuración mediante variables de entorno

- Puerto de escucha HTTP: `DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT` (por defecto `8080`).
- URL de recolección upstream: `DOOMSDAY_CLOCK_EXPORTER_SCRAPE_URL` (por defecto, la API REST del Bulletin).
- Tiempo de espera de la obtención upstream: `DOOMSDAY_CLOCK_EXPORTER_FETCH_TIMEOUT` (por defecto `30` segundos).
- Intervalo de recolección: `DOOMSDAY_CLOCK_EXPORTER_SCRAPE_INTERVAL` (por defecto `3600` segundos).
- TTL de la caché: `DOOMSDAY_CLOCK_EXPORTER_CACHE_TTL` (por defecto `86400` segundos).
- API de Wikipedia para la edad de los líderes nucleares: `DOOMSDAY_CLOCK_EXPORTER_LEADERS_URL` (por defecto `https://en.wikipedia.org/w/api.php`, `off` la desactiva).
- Página de Wikipedia con la tabla de autoridad de lanzamiento: `DOOMSDAY_CLOCK_EXPORTER_LEADERS_PAGE` (por defecto `List of states with nuclear weapons`).

<!-- textlint-enable -->

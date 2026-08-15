<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

# Configuración mediante variables de entorno

- Puerto de escucha HTTP: `O9S_DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT` (por defecto `8080`).
- URL de recolección upstream: `O9S_DOOMSDAY_CLOCK_EXPORTER_SCRAPE_URL` (por defecto, la API REST del Bulletin).
- Tiempo de espera de la obtención upstream: `O9S_DOOMSDAY_CLOCK_EXPORTER_FETCH_TIMEOUT` (por defecto `30` segundos).
- Intervalo de recolección: `O9S_DOOMSDAY_CLOCK_EXPORTER_SCRAPE_INTERVAL` (por defecto `3600` segundos).
- TTL de la caché: `O9S_DOOMSDAY_CLOCK_EXPORTER_CACHE_TTL` (por defecto `86400` segundos).

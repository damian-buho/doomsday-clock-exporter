<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

# Self-observability metrics

- The scraper registers on the Prometheus default registry, so the `/metrics` endpoint exposes the custom `doomsdayclock_*` gauges alongside the auto-registered Go runtime and process collectors.
- `go_*` metrics (goroutines, GC, memory, etc.) and `process_*` metrics (RSS, open FDs, CPU) let the operator monitor the exporter itself, not just the Doomsday Clock value.
- No extra wiring is required — `promhttp.Handler()` serves the default gatherer, which already carries these collectors.

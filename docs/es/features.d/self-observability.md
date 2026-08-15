<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

# Métricas de autoobservabilidad

- El recolector se registra en el registro por defecto de Prometheus, de modo que el endpoint `/metrics` expone los gauges personalizados `doomsdayclock_*` junto con los recolectores de runtime de Go y de proceso registrados automáticamente.
- Las métricas `go_*` (goroutines, GC, memoria, etc.) y `process_*` (RSS, descriptores abiertos, CPU) permiten al operador monitorizar el propio exportador, no solo el valor del Reloj del Juicio Final.
- No hace falta cableado adicional — `promhttp.Handler()` sirve el gatherer por defecto, que ya lleva estos recolectores.

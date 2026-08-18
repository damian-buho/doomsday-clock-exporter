<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

<!-- textlint-disable terminology,common-misspellings -->

# Recolección con caché y degradación elegante

- Un recolector en segundo plano obtiene el valor del Reloj del Juicio Final desde el Bulletin of Atomic Scientists en intervalos configurables.
- Los valores recolectados se guardan en caché con un TTL configurable; las recolecciones de Prometheus siempre devuelven el valor en caché sin bloquearse.
- Sirve valores obsoletos marcando un gauge de degradación (`doomsdayclock_cache_stale=1`) cuando la fuente original no está disponible.
- Los fallos de recolección disparan reintentos con retroceso exponencial, con un techo de 60 segundos.

<!-- textlint-enable -->

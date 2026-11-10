<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

<!-- textlint-disable terminology,common-misspellings -->

# Edad media y mediana de quienes tienen autoridad de lanzamiento

- Exporta `doomsdayclock_nuclear_leaders_average_age_years` y `doomsdayclock_nuclear_leaders_median_age_years`: la edad media y la mediana de quien puede ordenar un lanzamiento nuclear en cada Estado con armas nucleares.
- La imagen no incluye ninguna lista de líderes ni de Estados: cada recolección lee en Wikipedia la tabla de autoridad de lanzamiento, luego el titular actual de cada cargo y, por último, su fecha de nacimiento.
- Un cambio de líder aparece en la siguiente recolección, sin recompilar ni publicar una versión.
- Las edades cuentan los años completos más la parte transcurrida del año en curso, así que el valor crece entre recolecciones.
- `DOOMSDAY_CLOCK_EXPORTER_LEADERS_URL=off` la desactiva.

<!-- textlint-enable -->

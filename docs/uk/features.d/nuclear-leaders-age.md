<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

<!-- textlint-disable terminology,common-misspellings -->

# Середній і медіанний вік тих, хто має право на ядерний пуск

- Експортує `doomsdayclock_nuclear_leaders_average_age_years` і `doomsdayclock_nuclear_leaders_median_age_years`: середній і медіанний вік осіб, які можуть віддати наказ про ядерний пуск у кожній ядерній державі.
- Образ не містить жодного списку лідерів чи держав: кожен збір читає у Вікіпедії таблицю повноважень на пуск, потім чинного обіймача кожної посади, а тоді дату народження цієї особи.
- Зміна лідера з’являється під час наступного збору, без перезбирання чи випуску.
- Вік рахується як повні роки плюс частка поточного року, що минула, тож значення зростає між зборами.
- `DOOMSDAY_CLOCK_EXPORTER_LEADERS_URL=off` вимикає цю функцію.

<!-- textlint-enable -->

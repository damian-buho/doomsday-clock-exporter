<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

# Average and median age of the people with launch authority

- Exports `doomsdayclock_nuclear_leaders_average_age_years` and `doomsdayclock_nuclear_leaders_median_age_years`: the average and median age of whoever can order a nuclear launch in each nuclear-armed state.
- No list of leaders or states ships with the image: each scrape reads the launch authority table on Wikipedia, then each office’s current incumbent, then that person’s birth date.
- A change of leader shows up at the next scrape, with no rebuild or release.
- Ages count whole years plus the elapsed share of the current year, so the value grows between scrapes.
- Set `DOOMSDAY_CLOCK_EXPORTER_LEADERS_URL=off` to turn it off.

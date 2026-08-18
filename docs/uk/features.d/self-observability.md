<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

<!-- textlint-disable terminology,common-misspellings -->

# Метрики самоспостереження

- Збирач реєструється в типовому реєстрі Prometheus, тож кінцева точка `/metrics` викладає власні gauge-метрики `doomsdayclock_*` поруч із автоматично зареєстрованими колекторами Go-рантайму та процесу.
- Метрики `go_*` (горутини, GC, пам’ять тощо) і `process_*` (RSS, відкриті FD, CPU) дозволяють оператору стежити за самим експортером, а не лише за значенням Годинника Судного дня.
- Додаткове з’єднання не потрібне — `promhttp.Handler()` обслуговує типовий gatherer, який уже несе ці колектори.

<!-- textlint-enable -->

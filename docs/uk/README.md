<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>
SPDX-License-Identifier: MIT
pf-cli-managed: yes
-->

<!-- textlint-disable terminology,common-misspellings -->

[English](../../README.md) · [Español](../es/README.md)

# Doomsday Clock Exporter

Експортер Prometheus для значення Годинника Судного дня

[![Stand with Ukraine](https://raw.githubusercontent.com/vshymanskyy/StandWithUkraine/main/badges/StandWithUkraine.svg)](https://damian-buho.github.io/support-ukraine/) [![License](https://img.shields.io/static/v1?label=license&message=MIT&color=4c1&style=flat-square)](LICENSE) ![Commit style](https://img.shields.io/static/v1?label=commits&message=conventional&color=blue&style=flat-square) ![Workflow](https://img.shields.io/static/v1?label=workflow&message=git-flow&color=blue&style=flat-square) ![Versioning](https://img.shields.io/static/v1?label=versioning&message=semantic&color=blue&style=flat-square) [![PRs welcome](https://img.shields.io/static/v1?label=PRs&message=welcome&color=4c1&style=flat-square)](CONTRIBUTING.md) [![Citation](https://img.shields.io/static/v1?label=citation&message=cff&color=blue&style=flat-square)](CITATION.cff) [![REUSE compliance](https://api.reuse.software/badge/codeberg.org/o9s/doomsday-clock-exporter)](https://api.reuse.software/info/codeberg.org/o9s/doomsday-clock-exporter)

![Project status](https://img.shields.io/static/v1?label=status&message=maintained&color=1d63ed&style=flat-square) [![Last commit](https://img.shields.io/gitea/last-commit/o9s/doomsday-clock-exporter?gitea_url=https://codeberg.org&style=flat-square)](https://codeberg.org/o9s/doomsday-clock-exporter)

[![Build status on kiota.ch](https://kiota.ch/o9s/doomsday-clock-exporter/badges/workflows/published.yaml/badge.svg)](https://kiota.ch/o9s/doomsday-clock-exporter/actions)

## Можливості

- Кешований збір із плавною деградацією
- Налаштування змінними середовища
- Перевірка стану, специфічна для сервісу
- Метрики самоспостереження
- Persistent APT cache across builds
- Service process management with log routing (b19-exec)
- Cached artifact downloads with integrity verification (b19-fetch)
- Timed command execution with failure reporting (b19-run)
- Run-once initialization (bootstrap.d)
- Modular build hooks (build.d)
- Automatic CPU count detection (NUMPROCS)
- Declarative dependency management (b19-deps)
- Pluggable startup system (entrypoint.d)
- Feature toggles for all subsystems
- Built-in health monitoring (healthcheck.d)
- Multilingual shell output (b19-i18n)
- Image lineage tracking
- Structured, level-filtered logging (b19-log)
- Non-root container by default
- Air-gapped / offline build and runtime support
- Runtime overlay injection
- Reproducible base image (pinned by digest)
- Port validation
- Unified lifecycle runner family
- Docker secrets auto-loading (secrets)
- Interactive shell hooks (shell.d)
- Graceful signal handling
- Jinja2 configuration templates (minijinja-cli)
- Built-in test framework (test.d)
- Pre-installed utility tools
- XDG Base Directory paths

Див. [FEATURES.md](FEATURES.md), щоб переглянути повний перелік.

## Що надає цей проєкт

- **Виконуваний файл** `dist/doomsday-clock-exporter`
- **Образ контейнера** `ghcr.io/damian-buho/o9s/doomsday-clock-exporter:latest`
- **Образ контейнера** `docker.io/damianbuho/o9s-doomsday-clock-exporter:latest`

## Встановлення

Завантажте опублікований образ контейнера:

```sh
docker pull ghcr.io/damian-buho/o9s/doomsday-clock-exporter:latest
docker pull docker.io/damianbuho/o9s-doomsday-clock-exporter:latest
```

Якщо наведені вище реєстри недоступні, завантажте з джерела:

```sh
docker pull kiota.ch/o9s/doomsday-clock-exporter:latest
```

## Використання

Запустіть стек локально:

```sh
make dc-up
make dc-logs
make dc-down
```

## Збирання

- [Довідник із Makefile](../MAKEFILE.md)

Точки входу конвеєра:

- `make analyze` — Run the heavy analysis sweep (mutation testing, benchmarks)
- `make audited` — Re-scan the pinned dependencies and published artifacts for new vulnerabilities
- `make check-outdated` — Report every pinned dependency that lags upstream
- `make ready-to-publish` — Run the pseudo-CI pipeline locally — build, test and scan, without publishing

Виконайте `make` без аргументів для типової цілі; виконайте `make help`, щоб переглянути всі цілі.

Для локального циклу розробки `make dev-container` піднімає dev-container.

## Політики

- [Як зробити внесок](CONTRIBUTING.md)
- [Політика безпеки](SECURITY.md)
- [Як отримати підтримку](SUPPORT.md)
- [Кодекс поведінки](CODE_OF_CONDUCT.md)

## Посилання

### Проєкт

- [Специфікація Projectfile](https://projectfile.org)
- [Doomsday Clock Exporter на Codeberg](https://codeberg.org/o9s/doomsday-clock-exporter)
- [Doomsday Clock Exporter на GitHub](https://github.com/damian-buho/o9s-doomsday-clock-exporter)
- [Doomsday Clock Exporter на kiota.ch](https://kiota.ch/o9s/doomsday-clock-exporter)
- [Issues на Codeberg](https://codeberg.org/o9s/doomsday-clock-exporter/issues)
- [Issues на GitHub](https://github.com/damian-buho/o9s-doomsday-clock-exporter/issues)

### Інше

- [Від автора](https://dbuho.me)

## Ліцензія

Цей проєкт ліцензовано на умовах MIT — див. файл [LICENSE](LICENSE) для подробиць.

*Згенеровано з projectfile ([дізнатися як](https://projectfile.org/how-to/readme))*
<!-- textlint-enable -->

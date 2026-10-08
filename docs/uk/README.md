<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>
SPDX-License-Identifier: MIT
pf-cli-managed: yes
-->

<!-- textlint-disable terminology,common-misspellings -->

[English](../../README.md) · [Español](../es/README.md)

# Doomsday Clock Exporter

Експортер Prometheus, який збирає значення Годинника Судного дня від Bulletin of the Atomic Scientists і надає його як gauge секунд до опівночі. Фоновий збирач кешує значення з налаштовуваним TTL, віддає застарілі значення з gauge деградації при відмові джерела та повторює спроби з експоненційним відкатом.

[![Stand with Ukraine](https://raw.githubusercontent.com/vshymanskyy/StandWithUkraine/main/badges/StandWithUkraine.svg)](https://damian-buho.github.io/support-ukraine/) [![Projectfile inside](https://badges.kiota.ch/static/v1?label=projectfile&message=inside&labelColor=0d0d0d&color=8c6723&style=flat-square)](https://projectfile.org) [![License](https://badges.kiota.ch/static/v1?label=license&message=MIT&color=1e5913&style=flat-square)](LICENSE) [![Cosign](https://badges.kiota.ch/static/v1?label=cosign&message=enabled&color=1e5913&style=flat-square)](https://docs.sigstore.dev/cosign/verifying/verify/) ![ClamAV scanned](https://badges.kiota.ch/static/v1?label=clamav&message=scanned&color=1877aa&style=flat-square) [![PRs welcome](https://badges.kiota.ch/static/v1?label=PRs&message=welcome&color=1e5913&style=flat-square)](CONTRIBUTING.md) [![REUSE compliance](https://api.reuse.software/badge/codeberg.org/damian-buho/doomsday-clock-exporter)](https://api.reuse.software/info/codeberg.org/damian-buho/doomsday-clock-exporter)

![Project status](https://badges.kiota.ch/static/v1?label=status&message=maintained&color=1d63ed&style=flat-square) [![Last commit on kiota.ch](https://badges.kiota.ch/gitea/last-commit/damian-buho/doomsday-clock-exporter?gitea_url=https://kiota.ch&label=last%20commit%20on%20kiota.ch&style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter) [![Last commit on Codeberg](https://badges.kiota.ch/gitea/last-commit/damian-buho/doomsday-clock-exporter?gitea_url=https://codeberg.org&label=last%20commit%20on%20Codeberg&style=flat-square)](https://codeberg.org/damian-buho/doomsday-clock-exporter) [![Last commit on GitHub](https://badges.kiota.ch/github/last-commit/damian-buho/doomsday-clock-exporter?label=last%20commit%20on%20GitHub&style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter)

[![Publish pipeline on GitHub](https://github.com/damian-buho/doomsday-clock-exporter/actions/workflows/published.yaml/badge.svg?style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter/actions) [![Vulnerability audit on GitHub](https://github.com/damian-buho/doomsday-clock-exporter/actions/workflows/audited.yaml/badge.svg?style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter/actions) [![Dependency freshness on GitHub](https://github.com/damian-buho/doomsday-clock-exporter/actions/workflows/check-outdated.yaml/badge.svg?style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter/actions) [![Analysis sweep on GitHub](https://github.com/damian-buho/doomsday-clock-exporter/actions/workflows/analyzed.yaml/badge.svg?style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter/actions)

[![Publish pipeline on kiota.ch](https://kiota.ch/damian-buho/doomsday-clock-exporter/badges/workflows/published.yaml/badge.svg?style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter/actions) [![Vulnerability audit on kiota.ch](https://kiota.ch/damian-buho/doomsday-clock-exporter/badges/workflows/audited.yaml/badge.svg?style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter/actions) [![Dependency freshness on kiota.ch](https://kiota.ch/damian-buho/doomsday-clock-exporter/badges/workflows/check-outdated.yaml/badge.svg?style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter/actions) [![Analysis sweep on kiota.ch](https://kiota.ch/damian-buho/doomsday-clock-exporter/badges/workflows/analyzed.yaml/badge.svg?style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter/actions)

## Можливості

- Кешований збір із плавною деградацією
- Налаштування змінними середовища
- Перевірка стану, специфічна для сервісу
- Метрики самоспостереження

Також успадковує можливості B19 / Ubuntu — повний перелік див. у [Можливості](FEATURES.md).

## Швидкий старт

Збережіть це як `compose.yaml`:

```yaml
---
services:
  doomsday-clock-exporter:
    image: docker.io/damianbuho/doomsday-clock-exporter:latest
    ports:
      - "8080:8080"
    cap_drop: [ALL]
    security_opt: [no-new-privileges:true]
    restart: unless-stopped
```

Потім запустіть його командою `docker compose up --detach`.

## Що надає цей проєкт

- **Служба** `exporter` — слухає на `8080 (metrics)` — Prometheus metrics endpoint
- **Виконуваний файл** `doomsday-clock-exporter` — команда `doomsday-clock-exporter`
- **Образ контейнера** `ghcr.io/damian-buho/doomsday-clock-exporter:latest`
- **Образ контейнера** `damianbuho/doomsday-clock-exporter:latest`

## Встановлення

### Образ контейнера

Завантажте опублікований образ контейнера:

#### Завантажити з GHCR — linux/amd64, linux/arm64, linux/riscv64

```sh
docker pull ghcr.io/damian-buho/doomsday-clock-exporter:latest
```

#### Завантажити з DockerHub — linux/amd64

```sh
docker pull damianbuho/doomsday-clock-exporter:latest
```

Стабільні випуски також публікують теґи `X.Y.Z`, `X.Y` і `X` — завантажте той рівень точності, який хочете зафіксувати.

Якщо наведені вище реєстри недоступні, завантажте з джерела:

#### Завантажити з Kiota — linux/amd64

```sh
docker pull kiota.ch/damian-buho/doomsday-clock-exporter:latest
```

### Готовий бінарний файл

Завантажте готовий бінарний файл для своєї платформи з випусків на GitHub:

```sh
mkdir -p ~/.local/bin
curl --fail --location --output ~/.local/bin/doomsday-clock-exporter https://github.com/damian-buho/doomsday-clock-exporter/releases/latest/download/doomsday-clock-exporter-linux-$(uname -m)
chmod +x ~/.local/bin/doomsday-clock-exporter
~/.local/bin/doomsday-clock-exporter --help
```

Опубліковано для: `linux/amd64`, `linux/arm64`, `linux/riscv64`

## Використання

Запустіть сервіс у фоновому режимі, опублікувавши його порти:

### З GHCR

```sh
docker run --detach --publish 8080:8080/tcp ghcr.io/damian-buho/doomsday-clock-exporter:latest
```

### З DockerHub

```sh
docker run --detach --publish 8080:8080/tcp damianbuho/doomsday-clock-exporter:latest
```

Потім перевірте, що він відповідає:

```sh
curl http://localhost:8080/metrics
```

## Збирання

Клонуйте репозиторій разом із підмодулями:

```sh
git clone --recurse-submodules https://codeberg.org/damian-buho/doomsday-clock-exporter doomsday-clock-exporter && cd doomsday-clock-exporter
```

Зберіть образ контейнера локально:

```sh
make container-build
```

- [Довідник із Makefile](../how-to/MAKEFILE.md)

Виконайте `make` без аргументів для типової цілі; виконайте `make help`, щоб переглянути всі цілі.

Для локального циклу розробки `make dev-container` піднімає dev-container.

Точки входу конвеєра:

- `make analyzed` — Запускає важкий аналіз (мутаційне тестування, бенчмарки)
- `make audited` — Повторно сканує закріплені залежності й опубліковані артефакти на нові вразливості
- `make check-outdated` — Звітує про кожну закріплену залежність, що відстає від upstream
- `make ready-to-publish` — Запускає псевдо-CI локально — збирає, тестує й сканує без публікації

## Політики

- [Як зробити внесок](CONTRIBUTING.md)
- [Політика безпеки](SECURITY.md)
- [Як отримати підтримку](SUPPORT.md)
- [Кодекс поведінки](CODE_OF_CONDUCT.md)
- [Політика щодо ШІ та LLM](AI_POLICY.md)

## Посилання

- [Специфікація Projectfile](https://projectfile.org)

## Ліцензія

Цей проєкт ліцензовано на умовах MIT — див. файл [LICENSE](LICENSE) для подробиць.

<!-- textlint-enable -->

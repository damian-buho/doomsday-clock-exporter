<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>
SPDX-License-Identifier: MIT
pf-cli-managed: yes
-->

[Español](docs/es/README.md) · [Українська](docs/uk/README.md)

# Doomsday Clock Exporter

Prometheus exporter that scrapes the Bulletin of the Atomic Scientists Doomsday Clock value and exposes it as a gauge of seconds to midnight. A background scraper caches the value with a configurable TTL, serves stale values with a degradation gauge on upstream failure, and retries with exponential backoff.

[![Stand with Ukraine](https://raw.githubusercontent.com/vshymanskyy/StandWithUkraine/main/badges/StandWithUkraine.svg)](https://damian-buho.github.io/support-ukraine/) [![Projectfile inside](https://badges.kiota.ch/static/v1?label=projectfile&message=inside&labelColor=0d0d0d&color=8c6723&style=flat-square)](https://projectfile.org) [![License](https://badges.kiota.ch/static/v1?label=license&message=MIT&color=1e5913&style=flat-square)](LICENSE) [![PRs welcome](https://badges.kiota.ch/static/v1?label=PRs&message=welcome&color=1e5913&style=flat-square)](CONTRIBUTING.md) [![REUSE compliance](https://api.reuse.software/badge/codeberg.org/damian-buho/doomsday-clock-exporter)](https://api.reuse.software/info/codeberg.org/damian-buho/doomsday-clock-exporter)

![Project status](https://badges.kiota.ch/static/v1?label=status&message=maintained&color=1d63ed&style=flat-square) [![Last commit on kiota.ch](https://badges.kiota.ch/gitea/last-commit/damian-buho/doomsday-clock-exporter?gitea_url=https://kiota.ch&label=last%20commit%20on%20kiota.ch&style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter) [![Last commit on Codeberg](https://badges.kiota.ch/gitea/last-commit/damian-buho/doomsday-clock-exporter?gitea_url=https://codeberg.org&label=last%20commit%20on%20Codeberg&style=flat-square)](https://codeberg.org/damian-buho/doomsday-clock-exporter) [![Last commit on GitHub](https://badges.kiota.ch/github/last-commit/damian-buho/doomsday-clock-exporter?label=last%20commit%20on%20GitHub&style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter)

[![Publish pipeline on GitHub](https://github.com/damian-buho/doomsday-clock-exporter/actions/workflows/published.yaml/badge.svg?style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter/actions) [![Vulnerability audit on GitHub](https://github.com/damian-buho/doomsday-clock-exporter/actions/workflows/audited.yaml/badge.svg?style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter/actions) [![Dependency freshness on GitHub](https://github.com/damian-buho/doomsday-clock-exporter/actions/workflows/check-outdated.yaml/badge.svg?style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter/actions) [![Analysis sweep on GitHub](https://github.com/damian-buho/doomsday-clock-exporter/actions/workflows/analyzed.yaml/badge.svg?style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter/actions)

[![Publish pipeline on kiota.ch](https://kiota.ch/damian-buho/doomsday-clock-exporter/badges/workflows/published.yaml/badge.svg?style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter/actions) [![Vulnerability audit on kiota.ch](https://kiota.ch/damian-buho/doomsday-clock-exporter/badges/workflows/audited.yaml/badge.svg?style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter/actions) [![Dependency freshness on kiota.ch](https://kiota.ch/damian-buho/doomsday-clock-exporter/badges/workflows/check-outdated.yaml/badge.svg?style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter/actions) [![Analysis sweep on kiota.ch](https://kiota.ch/damian-buho/doomsday-clock-exporter/badges/workflows/analyzed.yaml/badge.svg?style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter/actions)

## Features

- Cached scraping with graceful degradation
- Environment variable configuration
- Service-specific healthcheck
- Self-observability metrics

It also inherits the features of B19 / Ubuntu — see [Features](docs/FEATURES.md) for the full list.

## Quick Start

Save this as `compose.yaml`:

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

Then start it with `docker compose up --detach`.

## What this provides

- **Service** `exporter` — listens on `8080 (metrics)` — Prometheus metrics endpoint
- **Executable** `doomsday-clock-exporter` — command `doomsday-clock-exporter`
- **Container image** `ghcr.io/damian-buho/doomsday-clock-exporter:latest`
- **Container image** `damianbuho/doomsday-clock-exporter:latest`

## Installation

### Container image

Pull the published container image:

#### Pull from GHCR — linux/amd64, linux/arm64, linux/riscv64

```sh
docker pull ghcr.io/damian-buho/doomsday-clock-exporter:latest
```

#### Pull from DockerHub — linux/amd64

```sh
docker pull damianbuho/doomsday-clock-exporter:latest
```

Stable releases also publish `X.Y.Z`, `X.Y` and `X` tags — pull the precision you want to pin.

If the registries above are unreachable, pull from the origin instead:

#### Pull from Kiota — linux/amd64

```sh
docker pull kiota.ch/damian-buho/doomsday-clock-exporter:latest
```

### Prebuilt binary

Download the prebuilt binary for your platform from the latest GitHub release:

```sh
curl --fail --location --output doomsday-clock-exporter https://github.com/damian-buho/doomsday-clock-exporter/releases/latest/download/doomsday-clock-exporter-$(uname -s | tr A-Z a-z)-$(uname -m | sed -e s/x86_64/amd64/ -e s/aarch64/arm64/) && chmod +x doomsday-clock-exporter
./doomsday-clock-exporter --help
```

Published for: `linux/amd64`, `linux/arm64`, `linux/riscv64`

## Usage

Run the service in the background, publishing its ports:

### From GHCR

```sh
docker run --detach --publish 8080:8080/tcp ghcr.io/damian-buho/doomsday-clock-exporter:latest
```

### From DockerHub

```sh
docker run --detach --publish 8080:8080/tcp damianbuho/doomsday-clock-exporter:latest
```

Then check that it answers:

```sh
curl http://localhost:8080/metrics
```

## Building

Clone the repository with its submodules:

```sh
git clone --recurse-submodules https://codeberg.org/damian-buho/doomsday-clock-exporter doomsday-clock-exporter && cd doomsday-clock-exporter
```

Build the container image locally:

```sh
make container-build
```

- [Makefile reference](docs/how-to/MAKEFILE.md)

Run `make` with no arguments for the default target; run `make help` to list every target.

For the local dev loop, `make dev-container` brings up the dev-container.

Pipeline entry points:

- `make analyzed` — Run the heavy analysis sweep (mutation testing, benchmarks)
- `make audited` — Re-scan the pinned dependencies and published artifacts for new vulnerabilities
- `make check-outdated` — Report every pinned dependency that lags upstream
- `make ready-to-publish` — Run the pseudo-CI pipeline locally — build, test and scan, without publishing

## Policies

- [How to contribute](CONTRIBUTING.md)
- [Security policy](SECURITY.md)
- [Getting support](SUPPORT.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [AI and LLM Policy](AI_POLICY.md)

## Links

- [Projectfile Specification](https://projectfile.org)

## License

This project is licensed under MIT — see the [LICENSE](LICENSE) file for details.

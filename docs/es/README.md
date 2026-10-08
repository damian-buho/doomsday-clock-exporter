<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>
SPDX-License-Identifier: MIT
pf-cli-managed: yes
-->

<!-- textlint-disable terminology,common-misspellings -->

[English](../../README.md) · [Українська](../uk/README.md)

# Doomsday Clock Exporter

Exportador de Prometheus que extrae el valor del Reloj del Juicio Final del Bulletin of the Atomic Scientists y lo expone como un indicador de segundos hasta la medianoche. Un raspador en segundo plano almacena el valor en caché con un TTL configurable, sirve valores obsoletos con un indicador de degradación ante fallos del origen, y reintenta con retroceso exponencial.

[![Stand with Ukraine](https://raw.githubusercontent.com/vshymanskyy/StandWithUkraine/main/badges/StandWithUkraine.svg)](https://damian-buho.github.io/support-ukraine/) [![Projectfile inside](https://badges.kiota.ch/static/v1?label=projectfile&message=inside&labelColor=0d0d0d&color=8c6723&style=flat-square)](https://projectfile.org) [![License](https://badges.kiota.ch/static/v1?label=license&message=MIT&color=1e5913&style=flat-square)](LICENSE) [![Cosign](https://badges.kiota.ch/static/v1?label=cosign&message=enabled&color=1e5913&style=flat-square)](https://docs.sigstore.dev/cosign/verifying/verify/) ![ClamAV scanned](https://badges.kiota.ch/static/v1?label=clamav&message=scanned&color=1877aa&style=flat-square) [![PRs welcome](https://badges.kiota.ch/static/v1?label=PRs&message=welcome&color=1e5913&style=flat-square)](CONTRIBUTING.md) [![REUSE compliance](https://api.reuse.software/badge/codeberg.org/damian-buho/doomsday-clock-exporter)](https://api.reuse.software/info/codeberg.org/damian-buho/doomsday-clock-exporter)

![Project status](https://badges.kiota.ch/static/v1?label=status&message=maintained&color=1d63ed&style=flat-square) [![Last commit on kiota.ch](https://badges.kiota.ch/gitea/last-commit/damian-buho/doomsday-clock-exporter?gitea_url=https://kiota.ch&label=last%20commit%20on%20kiota.ch&style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter) [![Last commit on Codeberg](https://badges.kiota.ch/gitea/last-commit/damian-buho/doomsday-clock-exporter?gitea_url=https://codeberg.org&label=last%20commit%20on%20Codeberg&style=flat-square)](https://codeberg.org/damian-buho/doomsday-clock-exporter) [![Last commit on GitHub](https://badges.kiota.ch/github/last-commit/damian-buho/doomsday-clock-exporter?label=last%20commit%20on%20GitHub&style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter)

[![Publish pipeline on GitHub](https://github.com/damian-buho/doomsday-clock-exporter/actions/workflows/published.yaml/badge.svg?style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter/actions) [![Vulnerability audit on GitHub](https://github.com/damian-buho/doomsday-clock-exporter/actions/workflows/audited.yaml/badge.svg?style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter/actions) [![Dependency freshness on GitHub](https://github.com/damian-buho/doomsday-clock-exporter/actions/workflows/check-outdated.yaml/badge.svg?style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter/actions) [![Analysis sweep on GitHub](https://github.com/damian-buho/doomsday-clock-exporter/actions/workflows/analyzed.yaml/badge.svg?style=flat-square)](https://github.com/damian-buho/doomsday-clock-exporter/actions)

[![Publish pipeline on kiota.ch](https://kiota.ch/damian-buho/doomsday-clock-exporter/badges/workflows/published.yaml/badge.svg?style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter/actions) [![Vulnerability audit on kiota.ch](https://kiota.ch/damian-buho/doomsday-clock-exporter/badges/workflows/audited.yaml/badge.svg?style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter/actions) [![Dependency freshness on kiota.ch](https://kiota.ch/damian-buho/doomsday-clock-exporter/badges/workflows/check-outdated.yaml/badge.svg?style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter/actions) [![Analysis sweep on kiota.ch](https://kiota.ch/damian-buho/doomsday-clock-exporter/badges/workflows/analyzed.yaml/badge.svg?style=flat-square)](https://kiota.ch/damian-buho/doomsday-clock-exporter/actions)

## Características

- Recolección con caché y degradación elegante
- Configuración mediante variables de entorno
- Comprobación de estado específica del servicio
- Métricas de autoobservabilidad

También hereda las características de B19 / Ubuntu; consulta [Características](FEATURES.md) para ver la lista completa.

## Inicio rápido

Guarda esto como `compose.yaml`:

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

Después, arráncalo con `docker compose up --detach`.

## Qué entrega este proyecto

- **Servicio** `exporter` — escucha en `8080 (metrics)` — Prometheus metrics endpoint
- **Ejecutable** `doomsday-clock-exporter` — comando `doomsday-clock-exporter`
- **Imagen de contenedor** `ghcr.io/damian-buho/doomsday-clock-exporter:latest`
- **Imagen de contenedor** `damianbuho/doomsday-clock-exporter:latest`

## Instalación

### Imagen de contenedor

Descarga la imagen de contenedor publicada:

#### Descargar de GHCR — linux/amd64, linux/arm64, linux/riscv64

```sh
docker pull ghcr.io/damian-buho/doomsday-clock-exporter:latest
```

#### Descargar de DockerHub — linux/amd64

```sh
docker pull damianbuho/doomsday-clock-exporter:latest
```

Las versiones estables también publican las etiquetas `X.Y.Z`, `X.Y` y `X`: descarga el nivel de precisión que quieras fijar.

Si los registros anteriores no están disponibles, descarga desde el origen:

#### Descargar de Kiota — linux/amd64

```sh
docker pull kiota.ch/damian-buho/doomsday-clock-exporter:latest
```

### Binario precompilado

Descarga el binario precompilado para tu plataforma desde las versiones de GitHub:

```sh
mkdir -p ~/.local/bin
curl --fail --location --output ~/.local/bin/doomsday-clock-exporter https://github.com/damian-buho/doomsday-clock-exporter/releases/latest/download/doomsday-clock-exporter-linux-$(uname -m)
chmod +x ~/.local/bin/doomsday-clock-exporter
~/.local/bin/doomsday-clock-exporter --help
```

Publicado para: `linux/amd64`, `linux/arm64`, `linux/riscv64`

## Uso

Ejecuta el servicio en segundo plano, publicando sus puertos:

### Desde GHCR

```sh
docker run --detach --publish 8080:8080/tcp ghcr.io/damian-buho/doomsday-clock-exporter:latest
```

### Desde DockerHub

```sh
docker run --detach --publish 8080:8080/tcp damianbuho/doomsday-clock-exporter:latest
```

Después, comprueba que responde:

```sh
curl http://localhost:8080/metrics
```

## Compilación

Clona el repositorio con sus submódulos:

```sh
git clone --recurse-submodules https://codeberg.org/damian-buho/doomsday-clock-exporter doomsday-clock-exporter && cd doomsday-clock-exporter
```

Construye la imagen de contenedor en local:

```sh
make container-build
```

- [Referencia del Makefile](../how-to/MAKEFILE.md)

Ejecuta `make` sin argumentos para el destino predeterminado; ejecuta `make help` para listar todos los destinos.

Para el bucle de desarrollo local, `make dev-container` levanta el dev-container.

Puntos de entrada de la canalización:

- `make analyzed` — Ejecuta el análisis pesado (pruebas de mutación, benchmarks)
- `make audited` — Vuelve a escanear las dependencias fijadas y los artefactos publicados en busca de vulnerabilidades nuevas
- `make check-outdated` — Informa de cada dependencia fijada que va por detrás de su versión upstream
- `make ready-to-publish` — Ejecuta localmente el pipeline pseudo-CI — compila, prueba y escanea, sin publicar

## Políticas

- [Cómo contribuir](CONTRIBUTING.md)
- [Política de seguridad](SECURITY.md)
- [Cómo obtener ayuda](SUPPORT.md)
- [Código de conducta](CODE_OF_CONDUCT.md)
- [Política sobre IA y LLM](AI_POLICY.md)

## Enlaces

- [Especificación de Projectfile](https://projectfile.org)

## Licencia

Este proyecto se publica bajo la licencia MIT — consulta el archivo [LICENSE](LICENSE) para más detalles.

<!-- textlint-enable -->

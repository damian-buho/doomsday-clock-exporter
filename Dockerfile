# SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>
#
# SPDX-License-Identifier: MIT

ARG B19_UBUNTU_BASE_IMAGE=registry.invalid/b19/ubuntu/resolute:latest
ARG B19_GO_BASE_IMAGE=registry.invalid/b19/go:latest
ARG B19_UBUNTU_SERIES=resolute

FROM ${B19_GO_BASE_IMAGE} AS o9s-doomsday-clock-exporter-builder

ARG B19_VERBOSITY
ARG LANG
ARG M6E_AI=N
ARG M6E_APT_CACHE_HOST
ARG M6E_APT_CACHE_PORT
ARG M6E_BUILD_DEBUG
ARG M6E_NEAR_CACHE_HOST
ARG M6E_PROJECT
ARG M6E_NAMESPACE
ARG M6E_VERSION
ARG TARGETARCH

COPY --chown=${B19_UID}:${B19_GID} .container/compile-go/           /
COPY --chown=${B19_UID}:${B19_GID} main.go go.mod go.sum            ${B19_HOME}/

USER 0

WORKDIR ${B19_HOME}

RUN --mount=type=bind,from=fetch,source=.,target=/fetch                                           \
    --mount=type=cache,target=${B19_DOWNLOAD_PATH},sharing=shared                                 \
    --mount=type=cache,target=${GOCACHE},sharing=locked                                           \
    --mount=type=cache,target=${GOMODCACHE},sharing=locked                                        \
    --mount=type=cache,id=apt-cache-${B19_UBUNTU_SERIES},target=/var/cache/apt,sharing=shared     \
    --mount=type=cache,id=apt-lists-${B19_UBUNTU_SERIES},target=/var/lib/apt,sharing=shared       \
    --mount=type=tmpfs,target=${B19_TEMP_PATH}                                                    \
    build-stage compile-go

# hadolint ignore=DL3066 # B19_UID comes from the root
USER ${B19_UID}

FROM ${B19_UBUNTU_BASE_IMAGE} AS o9s-doomsday-clock-exporter

ARG B19_VERBOSITY
ARG LANG
ARG M6E_AI=N
ARG M6E_APT_CACHE_HOST
ARG M6E_APT_CACHE_PORT
ARG M6E_BUILD_DEBUG
ARG M6E_NAMESPACE
ARG M6E_NEAR_CACHE_HOST
ARG M6E_PROJECT
ARG M6E_VERSION
ARG TARGETARCH

ENV M6E_VERSION=${M6E_VERSION}                            \
    O9S_DOOMSDAY_CLOCK_EXPORTER_CACHE_TTL=86400           \
    O9S_DOOMSDAY_CLOCK_EXPORTER_SCRAPE_INTERVAL=3600      \
    O9S_DOOMSDAY_CLOCK_EXPORTER_FETCH_TIMEOUT=30          \
    O9S_DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT=8080            \
    O9S_DOOMSDAY_CLOCK_EXPORTER_SCRAPE_URL=https://thebulletin.org/wp-json/wp/v2/pages/10305

COPY --chown=${B19_UID}:${B19_GID} .container/base/ /
COPY --from=o9s-doomsday-clock-exporter-builder /export /

USER 0

WORKDIR ${B19_HOME}

RUN --mount=type=bind,from=fetch,source=.,target=/fetch                                           \
    --mount=type=cache,target=${B19_DOWNLOAD_PATH},sharing=shared                                 \
    --mount=type=cache,id=apt-cache-${B19_UBUNTU_SERIES},target=/var/cache/apt,sharing=shared     \
    --mount=type=cache,id=apt-lists-${B19_UBUNTU_SERIES},target=/var/lib/apt,sharing=shared       \
    --mount=type=tmpfs,target=${B19_TEMP_PATH}                                                    \
    build-stage base

# hadolint ignore=DL3066 # B19_UID comes from the root
USER ${B19_UID}

COPY --chown=${B19_UID}:${B19_GID} .container/user/ /

RUN --mount=type=bind,from=fetch,source=.,target=/fetch                                             \
    --mount=type=cache,target=${B19_DOWNLOAD_PATH},sharing=shared,uid=${B19_UID},gid=${B19_GID}     \
    --mount=type=tmpfs,target=${B19_TEMP_PATH}                                                      \
    build-stage user

# ENTRYPOINT ["entrypoint.d"] is inherited
# HEALTHCHECK CMD ["healthcheck.d"] is inherited
# Don't use CMD ["sleep", "infinity"] here

# Enable Metrics Scraping by Prometheus Docker Discovery
LABEL prometheus.enabled=true
LABEL prometheus.port=${O9S_DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT}

# Enable Traefik Docker Discovery
LABEL traefik.enable=true
LABEL traefik.http.routers.doomsday-clock-exporter.rule="Host(`doomsday-clock-exporter.docker.localhost`)"
LABEL traefik.http.routers.doomsday-clock-exporter.entrypoints=web,websecure
LABEL traefik.http.routers.doomsday-clock-exporter.middlewares=redirect-to-https@file
LABEL traefik.http.services.doomsday-clock-exporter.loadbalancer.server.port=${O9S_DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT}

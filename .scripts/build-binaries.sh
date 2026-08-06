#!/bin/sh

# SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>
#
# SPDX-License-Identifier: MIT

set -eu

# build-binaries.sh — cross-compile doomsday-clock-exporter for one GOOS/GOARCH,
# inject the release version, and write the stripped binary to
# dist/doomsday-clock-exporter-<goos>-<goarch>. Called once per matrix axis by the
# build-binaries CI tool, which supplies the Go SDK (b19/go image, or host go on a
# prefer-local run) — the script never picks the runtime itself.

# Version resolution, portable across planes: the forge runner exports the tag as
# $GITHUB_REF_NAME (forwarded via the build-binaries tool's env: list); the make
# plane has neither an arg nor that env, so fall back to a git-derived / dev version.
# Keeps a local build honestly stamped, mirroring m6e-version.sh (tag → short sha → dev).
version="${1:-${GITHUB_REF_NAME:-}}"
case "${version}" in
	'' | *['{}']*) version="$(git describe --tags --always --dirty 2>/dev/null || echo dev)" ;;
esac

# GOHOSTOS/GOHOSTARCH is the real host even under a cross-compile (GOOS/GOARCH set
# only the TARGET). Default the target to the host when the matrix bound no cell —
# the make-plane host build; the forge matrix sets GOOS/GOARCH per cell.
hostos="$(go env GOHOSTOS)"
hostarch="$(go env GOHOSTARCH)"
goos="${GOOS:-${hostos}}"
goarch="${GOARCH:-${hostarch}}"
out="dist/doomsday-clock-exporter-${goos}-${goarch}"

log() { printf '[build-binaries] %s\n' "$*" >&2; }
log "building doomsday-clock-exporter ${version} for ${goos}/${goarch}"

mkdir -p dist
go build -ldflags="-s -w -X main.version=${version}"      \
         -o "${out}" .

# Stable unsuffixed host-native copy (dist/doomsday-clock-exporter) that gsa
# references by a fixed path — dropped by exactly the host-native cell.
if [ "${goos}/${goarch}" = "${hostos}/${hostarch}" ]; then
	log "host-native cell — copying ${out} -> dist/doomsday-clock-exporter"
	cp "${out}" dist/doomsday-clock-exporter
fi

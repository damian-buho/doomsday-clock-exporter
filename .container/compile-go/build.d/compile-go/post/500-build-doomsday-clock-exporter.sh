#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>
#
# SPDX-License-Identifier: MIT

  b19-run "DOOMSDAY" "$(_ "Download modules")" --     \
    go mod download

  b19-run "DOOMSDAY" "$(_ "Build")" --      \
    go build -ldflags="-s -w -X main.version=${M6E_VERSION:-dev}" -o doomsday-clock-exporter .

  b19-strip "DOOMSDAY" doomsday-clock-exporter

  b19-run "DOOMSDAY" "$(_ "Make directory in export")" --     \
    mkdir -p /export/usr/local/bin

  b19-run "DOOMSDAY" "$(_p "Copy %s to %s" "doomsday-clock-exporter" "/export/usr/local/bin")" --     \
    cp doomsday-clock-exporter /export/usr/local/bin

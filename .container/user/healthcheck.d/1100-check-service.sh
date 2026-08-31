#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>
#
# SPDX-License-Identifier: MIT

set -o pipefail

HTTP_PORT="${O9S_DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT}"

if ! curl -sf "http://localhost:${HTTP_PORT}/health" > /dev/null; then
  b19-log bad "HEALTH.D" "$(_p "Health check failed on port %s" "${HTTP_PORT}")"
  exit 1
fi

b19-log good "HEALTH.D" "$(_p "Service is responsive on port %s" "${HTTP_PORT}")"
exit 0

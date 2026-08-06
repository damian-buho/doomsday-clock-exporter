#!/usr/bin/env bash

# SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>
#
# SPDX-License-Identifier: MIT

  if [ "${ENTRYPOINT_COMMAND_EXECUTED:-N}" == "N" ]; then
    b19-log info "DOOMSDAY_CLOCK_EXPORTER" "$(_ "Starting")"

    b19-exec -- doomsday-clock-exporter
  fi

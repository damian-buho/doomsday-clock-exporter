<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

# damian-buho/doomsday-clock-exporter

Docker image built on [b19/go](../../b19/go/AGENTS.md) and [b19/Ubuntu](../../b19/ubuntu/AGENTS.md)

Custom Prometheus exporter for the Bulletin of Atomic Scientists’ Doomsday Clock value.

## Key facts

- Builder: `b19/go` (source code lives in the project directory — `main.go`, `go.mod`, `go.sum`)
- Final Base: `b19/ubuntu/resolute`
- Arch: amd64 only
- **No upstream version pin** — local source, versioned via the project’s own Git

## ENV

- `O9S_DOOMSDAY_CLOCK_EXPORTER_HTTP_PORT=8080`
- `O9S_DOOMSDAY_CLOCK_EXPORTER_CACHE_TTL=86400` (24h)
- `O9S_DOOMSDAY_CLOCK_EXPORTER_SCRAPE_INTERVAL=3600` (1h)
- `O9S_DOOMSDAY_CLOCK_EXPORTER_FETCH_TIMEOUT=30` (seconds)
- `O9S_DOOMSDAY_CLOCK_EXPORTER_SCRAPE_URL=https://thebulletin.org/wp-json/wp/v2/pages/10305`

## Note

Unique in o9s: the entire Go application is authored in-house. `main.go` is in the project root.

## Documentation

[Project goals and objectives](@docs/goal.md)
[Fitness criteria and acceptance](@docs/fit.md)
[Completed features and milestones](@docs/done.md)
[Known limitations and caveats](@docs/caveats.md)
[Future development plans](@docs/roadmap.md)
[Available make targets](@docs/MAKEFILE.md)

<!--
SPDX-FileCopyrightText: 2026 Damián Búho <damian.buho@proton.me>

SPDX-License-Identifier: MIT
-->

# In-house Go source build

- The entire Go application source (`main.go`, `go.mod`, `go.sum`) is authored in-house and lives in the project directory.
- No upstream version pin; the project is versioned via its own Git history.
- Binary is compiled at build time from the local source, with the project version stamped via ldflags.

# AGENTS.md

## Cursor Cloud specific instructions

This is `sqltoken` — a pure Go library (no runnable services). It tokenizes SQL strings into typed token arrays with support for MySQL/MariaDB, PostgreSQL/CockroachDB, Oracle, and SQL Server dialects.

**Key commands** (all run from repo root):

| Task | Command |
|------|---------|
| Test | `go test ./...` |
| Lint | `golangci-lint run ./...` |
| Build check | `go build ./...` |

**Notes:**

- `golangci-lint` must be on `PATH`. It is installed to `$(go env GOPATH)/bin`; ensure that directory is on your `PATH` (e.g. `export PATH="$PATH:$(go env GOPATH)/bin"`).
- No services, databases, or Docker are needed — all tests are self-contained unit tests.
- The `go.mod` declares `go 1.20`; CI tests against Go 1.18–1.22.

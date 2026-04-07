# CLAUDE.md

## Project overview

gokart is a collection of reusable Go packages for building services. The module path is `github.com/khalilonline/gokart`.

## Packages

- **pkg/logger** — High-performance structured JSON logger with bitmask-based level gating, field masking, emitters, and context propagation.
- **pkg/testflags** — Test filtering via `TEST_TYPE` env var. Supports `unit`, `integration`, `performance`, and `security` types.
- **pkg/utils** — Generic utilities for safe pointer dereferencing, type assertion, and slice conversion.

## Development

### Code style

- Go 1.26 with generics and range-over-int.
- No external test frameworks in pkg/logger — use stdlib `testing` only.
- Field constructors in the logger package are value types designed to stay on the stack.
- Logger uses `//go:noinline` on the hot path `log()` method intentionally — do not remove.

### Build & test commands

```bash
# Build
go build ./...

# Run unit tests
mise run unit-test
# or directly:
TEST_TYPE=unit go test ./...

# Run benchmarks
mise run bench-test
# or directly:
TEST_TYPE=performance go test -bench=. -benchmem ./...

# Lint and format (all: go, markdown, yaml)
mise run lint-and-format

# Lint and format (go only)
mise run lint-and-format-go

# Format markdown / yaml individually
mise run format-markdown
mise run format-yaml

# Tidy modules
go mod tidy

# Install git pre-commit hooks
mise run hooks-install

# Run pre-commit hooks manually
mise run pre-commit
```

### Test conventions

- All tests must be gated with `testflags.UnitTest(t)`, `testflags.IntegrationTest(t)`, `testflags.PerformanceTest(b)`, or `testflags.SecurityTest(t)` as the first line.
- Tests only run when `TEST_TYPE` matches. Without it set, all tests skip.
- Benchmark tests use `testflags.PerformanceTest(b)` (accepts `testing.TB`).

### Commits and PRs

- Use conventional commits standard when writing commit messages
- PR body should comply with [template](.github/PULL_REQUEST_TEMPLATE). However, if the issue number cannot be inferred from the branch name, then omit that section

# gokart

A collection of reusable Go packages for building services.

## Packages

### logger

High-performance, composable structured JSON logger with low-allocation field encoding, bitmask-based level gating, and buffer pooling.

```go
l := logger.New(os.Stdout, logger.ALL,
    logger.WithTimestamp(time.RFC3339),
    logger.WithCaller(),
)

l.Info("request handled",
    logger.Str("method", "GET"),
    logger.Int("status", 200),
    logger.Dur("latency", latency),
)
```

Features:

- **Levels** — `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL` with bitmask composition
- **Typed fields** — `Str`, `Int`, `Int64`, `Float64`, `Bool`, `Err`, `Time`, `Dur`, `Bytes`, `Val`, `Any`
- **Context propagation** — `NewCtx` / `FromCtx` for storing loggers in `context.Context`
- **Child loggers** — `With()` creates child loggers with pre-serialized context fields
- **Hooks** — Transform fields before encoding (e.g. masking)
- **Masking** — Built-in `Masker` hook for redacting sensitive keys and patterns (bearer tokens, API keys, PII)
- **Emitters** — Callbacks for structured log data (e.g. OTel bridge)
- **Level parsing** — `ParseLevel`, `LevelsAbove`, `UnmarshalText` for config integration

#### Benchmarks

Compared against [zerolog](https://github.com/rs/zerolog) and [zap](https://go.uber.org/zap) on Apple M3 Pro:

| Benchmark       | gokart                | zerolog               | zap                   |
| --------------- | --------------------- | --------------------- | --------------------- |
| Disabled level  | 15.02 ns/op, 1 alloc  | 2.98 ns/op, 0 allocs  | 20.22 ns/op, 1 alloc  |
| Simple message  | 20.57 ns/op, 0 allocs | 42.74 ns/op, 0 allocs | 132.1 ns/op, 0 allocs |
| With fields (5) | 148.7 ns/op, 1 alloc  | 128.2 ns/op, 0 allocs | 307.3 ns/op, 1 alloc  |
| With context    | 65.26 ns/op, 1 alloc  | 70.22 ns/op, 0 allocs | 202.4 ns/op, 1 alloc  |
| Parallel        | 110.7 ns/op, 1 alloc  | 17.42 ns/op, 0 allocs | 73.04 ns/op, 1 alloc  |

### protonum

Translates generated protobuf enums to the names a service stores, and back. `New` panics unless every defined value but the zero value has a unique, non-empty name, and on a name for the zero value or an undefined number. Build each translator in a package-level variable, so a value added to the proto without a name fails at load, in every test, instead of in one request.

```go
type ReportReason string // a domain package's own name type

var Reasons = protonum.New(map[eventsv1.ReportReason]ReportReason{
    eventsv1.ReportReason_REPORT_REASON_SPAM:   "spam",
    eventsv1.ReportReason_REPORT_REASON_UNSAFE: "unsafe",
})

name, ok := Reasons.Name(eventsv1.ReportReason_REPORT_REASON_SPAM) // "spam", true
reason, ok := Reasons.Enum("unsafe")                               // REPORT_REASON_UNSAFE, true
names, err := Reasons.Names(req.GetReasons())                      // fails with ErrUnnamed on an unnamed value
reasons, err := Reasons.Enums(stored)                              // fails with ErrUnknownName on an unknown name
all := Reasons.All()                                               // every name, in enum number order
```

`Name` and `Enum` report `false` for the zero value, an undefined number, or an unknown name. `NameOf` returns the empty name instead, and `KnownNames` drops values with no name. The batch calls wrap `ErrUnnamed` and `ErrUnknownName`, so each service maps them to its own errors with `errors.Is`.

### testflags

Test filtering based on the `TEST_TYPE` environment variable.

Examples:

```go
func TestMyFeature(t *testing.T) {
    testflags.UnitTest(t) // skips unless TEST_TYPE=unit
    // ...
}

func BenchmarkMyFeature(b *testing.B) {
    testflags.PerformanceTest(b) // skips unless TEST_TYPE=performance
    // ...
}
```

Supported types: `unit`, `integration`, `performance`, `security`.

### utils

Generic utility functions.

Examples:

```go
// Safe pointer dereferencing
val := utils.SafeDeref(ptr)              // returns zero value if nil
val := utils.SafeDerefOrDefault(ptr, 42) // returns default if nil
```

## Getting started

```bash
go get github.com/khalilonline/gokart
```

## Development

Requires [mise](https://mise.jdx.dev/) for task management.

```bash
mise install              # install toolchain
mise run unit-test        # run unit tests
mise run bench-test       # run benchmarks
mise run lint-and-format-go # lint and format
```

## License

MIT

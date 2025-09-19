# GitHub Copilot Instructions for experia-v10-exporter

These instructions are auto-generated from the repository contents. They are technology- and pattern-aware and must be followed by Copilot when suggesting or generating code for this repository.

## Project summary
- Language: Go
- Go version: 1.25.1 (from `go.mod`)
- Module path: `github.com/GrammaTonic/experia-v10-exporter`
- Project layout: `cmd/experia-v10-exporter` for the main binary, `internal/collector` for the collector package.
- Key dependencies: `github.com/prometheus/client_golang v1.23.2` used for metrics and HTTP handling.
- Runtime: CLI/http server exporting `/metrics` endpoint for Prometheus scraping.
- Configuration via environment variables: `EXPERIA_V10_*` family.

## Priority guidelines
1. Respect the Go toolchain and the Go version declared in `go.mod` (1.25.1). Do not use language features newer than this version.
2. Use the existing module path `github.com/GrammaTonic/experia-v10-exporter` for import paths within the repository.
3. Follow the current project layout: put main programs under `cmd/<name>/` and internal libraries under `internal/`.
4. Keep public API minimal: types and functions intended for use outside the package are exported (capitalized); internal helpers remain unexported.
5. Reuse existing Prometheus patterns (register collectors via `prometheus.Register`, use `prometheus.NewGauge`, `prometheus.NewCounter`, and `prometheus.MustNewConstMetric` for metrics).
6. Environment variables are the primary configuration mechanism—read via `os.Getenv`.
7. Prefer using `http.Client` with a timeout and cookie jar (already present in `internal/collector`) for HTTP interactions with the router.

## Version & dependency rules
- Always honor versions found in `go.mod` when importing third-party libraries.
- When adding new dependencies, prefer minimal, well-maintained packages and add them to `go.mod` via `go get` or `go mod tidy`.

## Code style & patterns
- Follow existing package organization: `cmd/` for executables, `internal/` for non-public packages.
- Use short, descriptive names for functions and variables, following Go conventions (camelCase for local, MixedCaps for exported).
- Error handling: check errors immediately and return or log with context using `fmt` or `log` consistent with existing files.
- Logging: use `log.Printf` or `log.Fatal` in `cmd` level. Internal packages should avoid calling `log.Fatal`; return errors instead.
- Avoid global mutable state in internal packages. Use structs that encapsulate state (e.g., `Experiav10Collector`).

## Testing recommendations
- No unit tests found. When adding tests:
  - Use the Go `testing` package.
  - Place package tests alongside package source files (`internal/collector/collector_test.go`).
  - Use table-driven tests for parsing and small logic. For HTTP interactions, use `httptest.Server` or inject an `http.Client` with a transport mock.

## File and package guidelines
- New commands: place a new `main` under `cmd/<toolname>/main.go` that imports internal packages by module path.
- New packages: prefer `internal/` for packages not intended for upstream reuse.
- Keep files small and focused. Split large files by responsibility (e.g., `auth.go`, `fetch.go`, `metrics.go`).

## Prometheus / metrics guidelines
- Use `prometheus.NewDesc`, `prometheus.NewGauge`, `prometheus.NewCounter` as appropriate.
- Register collectors in `cmd/` using `prometheus.Register(collector)`.
- Avoid creating expensive operations inside `Describe`; collect should perform network calls and create metrics via `MustNewConstMetric` or Gauges.

## Security and secrets
- Do not commit secrets. The project already uses environment variables for credentials—follow this pattern.
- Avoid printing passwords or sensitive tokens to logs. If logging is necessary, redact sensitive fields.

## Examples (follow these patterns)

- Creating a new collector (short example):

```go
// in internal/collector/mycollector.go
package collector

import (
    "net"
    "time"

    "github.com/prometheus/client_golang/prometheus"
)

type MyCollector struct {
    ip net.IP
    timeout time.Duration
    up prometheus.Gauge
}

func NewMyCollector(ip net.IP, timeout time.Duration) *MyCollector {
    return &MyCollector{
        ip: ip,
        timeout: timeout,
        up: prometheus.NewGauge(prometheus.GaugeOpts{Name: metricPrefix + "my_up"}),
    }
}

func (c *MyCollector) Describe(ch chan<- *prometheus.Desc) {
    c.up.Describe(ch)
}

func (c *MyCollector) Collect(ch chan<- prometheus.Metric) {
    // perform network calls, set metrics
    c.up.Set(1)
    c.up.Collect(ch)
}
```

## What NOT to do
- Do not import and use packages that are not present in `go.mod` without adding them properly.
- Do not change `go.mod` `go` directive without explicit need.
- Do not export internal helpers unless they are intended for external use.

## Where to place this file
Write this file to `.github/copilot/copilot-instructions.md` so Copilot can find it.

---
Generated by copilot-instructions-blueprint-generator based on repository analysis. Only include guidance that can be justified from the repository content.

# GitHub Copilot Instructions for experia-v10-exporter

Concise guidance for AI agents editing this repository. Focus on discoverable, repo-specific patterns, build/test workflows, and integration points.

## Quick facts
- Language: Go (go.mod: go 1.25.1)
- Module: github.com/GrammaTonic/experia-v10-exporter
- Main binary: `cmd/experia-v10-exporter`
- Core package: `internal/collector` (scrapes Experia V10 modem and exposes Prometheus metrics)

## Architecture & why
- The collector runs a periodic scrape of the modem (HTTP JSON RPC) and exposes metrics on `/metrics` for Prometheus.
- `cmd/experia-v10-exporter` wires configuration (via `EXPERIA_V10_*` env vars), creates the collector, and registers it with Prometheus.
- `internal/collector` contains: auth, fetch, JSON parsing, metric construction. Keep networking and parsing logic in small files (`auth.go`, `fetch.go`, `json_prod.go`, `metrics.go`).

## Project-specific patterns
- Tests are isolated from network by injecting `http.Client` or `Transport` mocks (see `internal/testutil` for helpers: `RoundTripperFunc`, `RewriteTransport`, `MakeJSONHandler`). Reuse these helpers rather than duplicating round-tripper stubs.
- Prometheus metrics follow the usual pattern: create Gauges/Descs in package, register in `cmd`, and use `MustNewConstMetric` when exporting scraped values.
- Config is read strictly from environment variables; prefer `os.Getenv` and respect existing names (`EXPERIA_V10_*`).
- Avoid global mutable state in `internal` packages; prefer struct-based state (e.g., `Experiav10Collector`).

## Tests & developer workflows
- Run unit tests locally: `go test ./... -v` (Ginkgo is used in some packages; `ginkgo -v ./...` is also supported).
- Use `httptest.Server` or `internal/testutil` helpers to provide deterministic responses for JSON endpoints. Example: `testutil.MakeJSONHandler(sample)`.
- Behavior tests use Ginkgo/Gomega; keep their side-effects isolated and use the test helper registry patterns already present.

## Debugging tips
- If tests fail with network timeouts, check for accidental real network calls; ensure the test injects a mock `Transport` or `httptest.Server`.
- `cmd` tests may log simulated login warnings — these are expected in test mode; focus on assertions and registry values.

## Files to inspect for common edits
- `cmd/experia-v10-exporter/main.go` — wiring and Prometheus registration
- `internal/collector/{auth.go,fetch.go,json_prod.go,metrics.go}` — main logic split across small files
- `internal/testutil/testutil.go` — shared test helpers (use instead of re-creating helpers)

## Additions & dependencies
- Respect versions in `go.mod`. When adding dependencies, run `go get` and `go mod tidy`.

## Examples (copyable)
- Run tests:
```
go test ./... -v
```
- Run Ginkgo behavior tests:
```
ginkgo -v ./internal/collector
```

## What NOT to do
- Don't perform actual network calls in unit tests. Use injected `http.Client`/`Transport` or `httptest.Server`.
- Don't export internal helpers unless needed; prefer adding shared helpers to `internal/testutil`.

If any section is unclear or you want the shorter or longer variant, tell me which parts to adjust.
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

## Folder structure
- `cmd/experia-v10-exporter/`: main program entrypoint.
- `internal/collector/`: main logic for scraping the Experia V10 router and exposing metrics.
- `internal/testutil/`: test helpers for HTTP mocking and Prometheus metric assertions.
- `.github/`: GitHub workflows and Copilot instructions.
- `go.mod` and `go.sum`: dependency management.
- `README.md`: project overview and usage instructions.
- `docs/`: documentation (if any).
- `examples/`: example configurations or usage (if any).
- `scripts/`: utility scripts (if any).
- `docker/`: Dockerfiles or related files (if any).
- `examples/modem_json/`: example JSON outputs from the modem.
- `test-output/`: example outputs for tests (if any).

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

## Go-specific guidelines
- Respect the `go` version declared in `go.mod` and avoid using language features newer than that version.
- Run `gofmt` (or `gofmt -s`) for formatting and `go vet` as part of local development. Prefer `gofumpt` only if the project opts in.
- Prefer `context.Context` for cancellation and timeouts in network calls and exported APIs; pass context from callers where possible.
- Use small interfaces to enable testing (e.g., define an interface for HTTP clients or metrics sinks and inject implementations in tests).
- Favor explicit error wrapping with `fmt.Errorf("context: %w", err)` to preserve cause chains.
- Keep exported APIs well-documented with Go doc comments (starting with the symbol name) for public types and functions.
- Add unit tests alongside code. For HTTP interactions prefer `httptest.Server` or injecting an `http.Client` with a mocked `Transport` for deterministic tests.
- When adding dependencies, prefer minimal, widely-used libraries and add them using `go get` / `go mod tidy`; avoid unnecessary transitive dependencies.

## Testing recommendations
- The project prefers behavior-driven tests using Ginkgo (v2) and Gomega for assertions when adding new tests.
  - Install the test helpers with:

    ```bash
    go get github.com/onsi/ginkgo/v2
    go get github.com/onsi/gomega
    ```

  - Initialize a test suite in a package directory with:

    ```bash
    ginkgo bootstrap
    ginkgo generate <package-name>
    ```

  - Place tests alongside package source files, e.g. `internal/collector/collector_test.go` and use `_test.go` suffixes.
  - Use Ginkgo's Describe/Context/It blocks and Gomega matchers for clear, readable tests.
  - For HTTP interactions, prefer using `httptest.Server` or inject a custom `http.Client` with a mock `Transport` so tests run deterministically without network access.
  - When testing JSON parsing or small logic, use table-driven tests within `Describe/It` blocks to cover edge cases.
  - Example test outline:

    ```go
    package collector_test

    import (
      . "github.com/onsi/ginkgo/v2"
      . "github.com/onsi/gomega"
      "net/http"
      "net/http/httptest"
      "github.com/GrammaTonic/experia-v10-exporter/internal/collector"
    )

    var _ = Describe("Experia Collector", func() {
      It("parses WAN status correctly", func() {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(http.StatusOK)
          w.Write([]byte(`{"status":true,"data":{"ConnectionState":"Connected"}}`))
        }))
        defer server.Close()

        // inject http.Client with server.URL or set up collector to use the test server
        Expect(true).To(BeTrue())
      })
    })
    ```

  - Run tests with `ginkgo -v ./...` or `go test ./...` (ginkgo also supports `ginkgo -r` for recursive runs).
  - Add CI steps to run `ginkgo -v ./...` or `go test ./...` in GitHub Actions.

- If you prefer classic `testing` package tests for small utilities, it's acceptable, but larger integration/behavior tests should use Ginkgo.
- Follow existing patterns for test naming, organization, and use table-driven styles where appropriate.

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

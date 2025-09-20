# Call and Correlation Map

This document lists the primary functions and methods in `internal/collector` and where they are called across the repository. It's intended to help understand runtime flow and test coverage correlations.

## Summary

- Package: `internal/collector`
- Main types: `Experiav10Collector`, `sessionContext`
- Key methods/functions covered: `NewCollector`, `Login`, `Collect`, `fetchURL`, `authenticate`, `CookiesForHost`, `SessionToken`, `setCookiesFromResponse`, `Describe`

---

## Functions / Methods

### NewCollector
Signature: `func NewCollector(ip net.IP, username, password string, timeout time.Duration, candidates ...string) *Experiav10Collector`
Purpose: construct a configured `Experiav10Collector` (HTTP client + metrics). Normalizes optional netdev candidates.
Call sites:
- `cmd/print-cookiejar/main.go:20` — example CLI uses NewCollector to create a collector and inspect cookies.
- `cmd/experia-v10-exporter/main.go:31` — main binary constructs the collector before Login.
- Many unit and integration tests under `internal/collector/*_test.go` create collectors (see tests: `metrics_test.go`, `mib_parse_test.go`, `getnetdevstats_test.go`, `mib_output_json_test.go`, ginkgo tests, etc.). Examples:
  - `internal/collector/metrics_test.go:21`
  - `internal/collector/mib_parse_test.go:20`
  - `internal/collector/getnetdevstats_test.go:20`
  - `internal/collector/mib_output_json_test.go:25`

### Login
Signature: `func (c *Experiav10Collector) Login() error`
Purpose: authenticate to the device and store session token and cookies on the collector.
Call sites:
- `cmd/print-cookiejar/main.go:21` — used by the small CLI to obtain cookies before printing.
- `cmd/experia-v10-exporter/main.go:36` — main binary calls Login during setup and exits on failure.
- Internally referenced in documentation/comments inside `collector.go` as the expected startup call.

### Collect
Signature: `func (c *Experiav10Collector) Collect(ch chan<- prometheus.Metric)`
Purpose: core scrape path — ensures a session, posts multiple NeMo/sah service calls, parses responses, and emits Prometheus metrics.
Call sites:
- Prometheus/registry integration in tests and helpers: `internal/testutil/testutil.go:64`, `internal/collector/testhelpers_test.go:57` and multiple ginkgo tests call `c.Collect(ch)` directly to exercise metric emission.
- `internal/collector/collector_ginkgo_test.go:58` and other tests call Collect to exercise scrape behavior.

Key correlations inside `Collect`:
- Calls `c.authenticate()` when session token is empty (fallback instead of Login). (`internal/collector/collector.go:108,127`)
- Uses `postFetch` helper (closure inside Collect) which calls `c.fetchURL(...)` (`internal/collector/collector.go:188`)
- Emits metrics defined in `metrics.go` (e.g., `ifupTime`, `netdev*` descriptors) via `prometheus.MustNewConstMetric`.

### fetchURL
Signature: `func (c *Experiav10Collector) fetchURL(ctx context.Context, method, url string, headers map[string]string, body []byte) ([]byte, error)`
Purpose: wrapper for performing HTTP requests using the collector's HTTP client, applying headers and returning body or an error for non-2xx responses.
Call sites:
- `internal/collector/collector.go:188` — used by `postFetch` within `Collect` to perform POST calls.
- Unit tests call it directly to exercise error handling and header propagation: `internal/collector/collector_extra_ginkgo_test.go:57`, `internal/collector/collector_cover_ginkgo_test.go:43`, `internal/collector/metrics_test.go:78`.

### authenticate
Signature: `func (c *Experiav10Collector) authenticate() (sessionContext, error)`
Purpose: performs the login flow using `sah.Device.Information.createContext` and records contextID; stores cookies via `setCookiesFromResponse` and optionally performs a follow-up GET to capture additional cookies.
Call sites:
- `internal/collector/collector.go:108` — called by `Login()` during startup.
- `internal/collector/collector.go:127` — called by `Collect()` as fallback if no session present.
- Many tests call `authenticate()` directly to assert behavior in error and success modes: `internal/collector/metrics_test.go` and other test files (`collector_extra_ginkgo_test.go`, `collector_cover_ginkgo_test.go`).
Internal correlations:
- Calls `jsonMarshal` (test override aware) for payload serialization.
- Calls `newRequest` helper (test override points) to build HTTP requests.
- Calls `setCookiesFromResponse` to store cookies into the collector's Jar.
- Calls `c.client.Do(req)` and `c.client.Do(getReq)` to execute HTTP requests (which tests often override via custom RoundTripper).

### CookiesForHost
Signature: `func (c *Experiav10Collector) CookiesForHost(hostURL string) []*http.Cookie`
Purpose: convenience accessor to return cookies stored in the collector's Jar for a given host URL. Returns nil on parse errors or when no jar present.
Call sites:
- `cmd/print-cookiejar/main.go:39` — the CLI uses this to print cookies for an input URL.

### SessionToken
Signature: `func (c *Experiav10Collector) SessionToken() string`
Purpose: thread-safe accessor for the stored session token (contextID).
Call sites:
- `cmd/print-cookiejar/main.go:26` — the CLI prints token to the console.

### setCookiesFromResponse
Signature: `func setCookiesFromResponse(jar http.CookieJar, resp *http.Response, reqURL *url.URL, fallbackURL string)`
Purpose: robust helper for setting cookies into a Jar. Handles missing resp.Request URLs and falls back to parsing raw Set-Cookie headers when net/http's cookie parsing misses non-standard cookie names.
Call sites:
- `internal/collector/auth.go:53` — used to store cookies after authentication response.
- `internal/collector/auth.go:96` — used to store cookies after the follow-up GET.
- Tests exercise the helper directly: `internal/collector/metrics_test.go:272,282`.

### Describe
Signature: `func (c *Experiav10Collector) Describe(ch chan<- *prometheus.Desc)`
Purpose: describe metrics (prometheus.Descriptors) supplied by this collector.
Call sites:
- Registered implicitly when the collector is registered with Prometheus in `cmd/experia-v10-exporter`. Tests exercise Describe through collectors created in tests: `internal/collector/metrics_test.go`.

---

## Notes and next steps

- This file focuses on `internal/collector`. If you want the same mapping across the entire repo (other packages, cmd/), I can extend the scan.
- I intentionally grouped test call-sites separately — tests exercise most internal functions directly, making it easy to see coverage.

If you'd like, I can:
- Add cross-references to specific test assertions for each call site.
- Generate a UML-style call graph (SVG) from these mappings.

---

Generated on: 2025-09-20

## postFetch payloads

The `postFetch` helper closure inside `Collect` is used to perform POST requests to the router API. Below are all payloads passed to `postFetch` in `internal/collector/collector.go` and where they are used.

- `{"service":"NMC","method":"getWANStatus","parameters":{}}` — used to fetch WAN status and emit `ifupTime` / internet connection metrics. (call: `internal/collector/collector.go:198`, assigned to `wanStatusResp`).
- `fmt.Sprintf(`{"service":"NeMo.Intf.%s","method":"getMIBs","parameters":{}}`, cand)` — per-interface MIB fetch used to extract netdev static info and flags. (call: `internal/collector/collector.go:303`, assigned to `resp`).
- `fmt.Sprintf(`{"service":"NeMo.Intf.%s","method":"getNetDevStats","parameters":{}}`, cand)` — per-interface statistics fetch used to emit netdev_* counters (Rx/Tx bytes/packets/errors). (call: `internal/collector/collector.go:527`, assigned to `statsResp`).

Notes:
- `postFetch` sets headers including `Authorization: X-Sah <token>` and `x-context: <token>` and uses `c.fetchURL` to perform the request. When `postFetch` returns an empty string the code emits zero-valued metrics so metric families remain present.
- If you'd like I can extract these payloads programmatically and produce a small JSON/YAML manifest for tooling or tests.

### New service modules added

To support modularization per your request, small service modules were added under `internal/collector/services`:

- `services/nmc` — contains `RequestBody()` for `getWANStatus` and a minimal `ParseWANStatus` helper. File: `internal/collector/services/nmc/getwanstatus.go`.
- `services/nemo` — contains `RequestBody(candidate)` and `RequestBodyStats(candidate)` for `getMIBs` and `getNetDevStats` respectively. Files: `internal/collector/services/nemo/*.go`.

The collector was updated to call these helpers instead of hard-coded JSON strings.

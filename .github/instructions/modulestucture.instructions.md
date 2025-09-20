---
description: 'Instructions for organizing code into modules for better maintainability and scalability'
applyTo: "**"
---

# Go Module Structure Instructions
## Overview
This document provides guidelines for organizing Go code into modules to enhance maintainability, scalability, and clarity. Proper module structure helps in managing dependencies, versioning, and collaboration among developers.

## internal layout

The `internal/` tree contains packages that are implementation details of this repository. Keep packages small and focused. The layout below reflects how the collector code is split by responsibility. For each directory we add a one-line example of what's included to help contributors find the right place for new code.

- **internal/collector** — core collector logic for talking to the Experia V10 and exposing Prometheus metrics.
    - Example: `collector.go` implements the Prometheus Collector interface and coordinates API calls.

- **internal/collector/services** — service-specific API clients and callers.
    - `nmc/` — NMC service helpers
        - Example: `getwanstatus.go` implements the `getWANStatus` call and returns a parsed model.
    - `nemo/` — NeMo service helpers
        - Example: `getmibs.go` (calls `getMIBs`) and `getnetdevstats.go` (calls `getNetDevStats`).

- **internal/collector/metrics** — Prometheus metric definitions and helpers.
    - `nmc/` — metrics related to NMC responses
        - Example: `metrics_getwanstatus.go` registers and updates WAN-related metrics.
    - `nemo/` — metrics related to NeMo responses
        - Example: `metrics_getmibs.go`, `metrics_getnetdevstats.go`.
    - `helpers.go` — shared helpers for metric creation and label handling.

- **internal/collector/parser** — response parsers that convert raw API payloads into typed models.
    - `nemo/`
        - Example: `parse_mibs.go` and `parse_netdevstats.go` transform HTTP responses into Go structs.
    - `nmc/`
        - Example: `parse_wanstatus.go` parses the `getWANStatus` response.

- **internal/collector/models** — shared data models used across services, parsers, and metrics.
    - Example: `models.go` contains structs that represent API responses and normalized internal types.

- **internal/collector/connectivity** — HTTP client, authentication and fetch helpers.
    - Example: `client.go` configures the HTTP client, `auth.go` holds login/session logic, and `fetch.go` exposes helpers to perform authenticated requests.



---
description: 'Instructions for organizing code into modules for better maintainability and scalability'
applyTo: "**"
---

# Go Module Structure Instructions
## Overview
This document provides guidelines for organizing Go code into modules to enhance maintainability, scalability, and clarity. Proper module structure helps in managing dependencies, versioning, and collaboration among developers.
## Layout
- **Root Directory**: The root of your project should contain the `go.mod` file
    - This file defines the module path and its dependencies.
- **cmd/**: This directory contains the main applications for your project.
    - Each application should have its own subdirectory (e.g., `cmd/app1`, `cmd/app2`).
    - Each subdirectory should contain a `main.go` file.
- **internal/**: This directory contains packages that are only intended for use within your project.
    - Similar to `pkg/`, organize related functionality into subdirectories (e.g., `internal/collector`, `internal/utils`).

## internal layout
- **internal/collector**: Contains the main logic for collecting data from the Experia V10 device.
    - `collector.go`: Main collector implementation.
- **internal/collector/services**: Contains service-specific modules for handling different API calls.
    - `nmc/`: Module for NMC service-related functionality.
        - `getwanstatus.go`: Handles the `getWANStatus` API call.
    - `nemo/`: Module for NeMo service-related functionality.
        - `getmibs.go`: Handles the `getMIBs` API call.
        - `getnetdevstats.go`: Handles the `getNetDevStats` API call.
- **internal/collector/metrics**: Contains metric definitions and helpers.
    -  `nmc/`: Module for NMC service-related functionality.
        - `metrics_getwanstatus.go`: Defines metrics related to WAN status.
    - `nemo/`: Module for NeMo service-related functionality.
        - `metrics_getmibs.go`: Defines metrics related to MIBs.
        - `metrics_getnetdevstats.go`: Defines metrics related to network device statistics.
    - `helpers.go`: Common metric helper functions.
- **internal/collector/parser**: Contains parsers for API responses.
    - `nemo/`: Module for NeMo service-related functionality.
        - `parse_mibs.go`: Parses the response from `getMIBs`.
        - `parse_netdevstats.go`: Parses the response from `getNetDevStats`.
    - `nmc/`: Module for NMC service-related functionality.
        - `parse_wanstatus.go`: Parses the response from `getWANStatus`.
- **internal/collector/models**: Contains data models and types used across the collector.
    - `models.go`: Structs and types for API responses and internal data representation.
- **internal/collector/connectivity**: Contains code related to HTTP client and authentication.
    - `client.go`: HTTP client setup and request functions.
    - `auth.go`: Authentication logic for the Experia V10 device.
    - `fetch.go`: Functions for fetching data from the device.



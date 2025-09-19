# experia-v10-exporter
A [prometheus](https://prometheus.io) exporter for getting some metrics of an Experia Box v10 (H369A)

## Features
- Scrapes internet connection status from the Experia Box V10 router
- Exports metrics in Prometheus format
- Supports authentication with the router's API
- Provides metrics on connection state, IP, MAC, link type, and protocol

## Quick Start

### Using Docker (Recommended)
1. Clone the repository:
   ```bash
   git clone https://github.com/GrammaTonic/experia-v10-exporter.git
   cd experia-v10-exporter
   ```

2. Set environment variables:
   ```bash
   export EXPERIA_V10_ROUTER_USERNAME=Admin
   export EXPERIA_V10_ROUTER_PASSWORD=your_password
   export EXPERIA_V10_ROUTER_IP=192.168.2.254
   export EXPERIA_V10_LISTEN_ADDR=localhost:9684
   export EXPERIA_V10_TIMEOUT=10
   ```

3. Run with Docker Compose:
   ```bash
   docker compose up -d
   ```

### Building from Source
1. Ensure you have Go 1.19+ installed
2. Clone and build:
   ```bash
   git clone https://github.com/GrammaTonic/experia-v10-exporter.git
   cd experia-v10-exporter
   go mod tidy
   go build ./cmd/experia-v10-exporter
   ```

3. Run the binary:
   ```bash
   ./experia-v10-exporter
   ```

## Configuration
The exporter is configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `EXPERIA_V10_LISTEN_ADDR` | `localhost:9684` | Address and port to listen on |
| `EXPERIA_V10_TIMEOUT` | `10` | Timeout in seconds for API requests |
| `EXPERIA_V10_ROUTER_IP` | `192.168.2.254` | IP address of the Experia Box router |
| `EXPERIA_V10_ROUTER_USERNAME` | Required | Router admin username |
| `EXPERIA_V10_ROUTER_PASSWORD` | Required | Router admin password |

## Metrics
The exporter provides the following metrics:

- `experia_v10_up`: Whether the exporter is able to scrape the router (1 if up, 0 if down)
- `experia_v10_auth_errors_total`: Number of authentication errors
- `experia_v10_scrape_errors_total`: Number of scraping errors
- `experia_v10_permission_errors_total`: Number of permission denied errors from the router API
- `experia_v10_internet_connection`: Internet connection status with labels:
  - `link_type`: Type of link (e.g., Ethernet, WiFi)
  - `protocol`: Connection protocol
  - `connection_state`: Current state (e.g., Connected, Disconnected)
  - `ip`: IP address
  - `mac`: MAC address

## Prometheus Configuration
Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'experia-v10'
    static_configs:
      - targets: ['localhost:9684']
    scrape_interval: 30s
```

## Troubleshooting
- Ensure the router IP and credentials are correct
- Check that the router's web interface is accessible
- Verify firewall settings allow connections to the router
- Use `curl http://localhost:9684/metrics` to test the exporter

## Development
The project uses a standard Go project structure:
- `cmd/`: Main application entry point
- `internal/collector/`: Core collector logic, split into multiple files for modularity
- `pkg/`: (Future) Shared packages

To contribute:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and build
5. Submit a pull request

## Netdev metrics (added)

The collector now dynamically discovers network interfaces from the router MIBs and exposes these metrics:

- `experia_v10_netdev_up{ifname="..."}` — 1 if interface is up, 0 otherwise
- `experia_v10_netdev_mtu{ifname="..."}` — MTU value
- `experia_v10_netdev_tx_queue_len{ifname="..."}` — transmit queue length
- `experia_v10_netdev_speed_mbps{ifname="..."}` — link speed in Mbps (if available)
- `experia_v10_netdev_last_change_seconds{ifname="..."}` — last change time (seconds)
- `experia_v10_netdev_info{ifname="...",alias="...",flags="...",lladdr="...",type="..."}` — presence/info metric (value 1)

Notes:
- `ifname` is normalized to uppercase when emitted to avoid duplicate series and make queries simpler.

PromQL examples:

- Count interfaces that are up:

   `count(experia_v10_netdev_up == 1)`

- MTU for ETH0:

   `experia_v10_netdev_mtu{ifname="ETH0"}`

- Interfaces with MTU less than 1500:

   `experia_v10_netdev_mtu < 1500`


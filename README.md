# experia-v10-exporter
A [prometheus](https://prometheus.io) exporter for getting some metrics of an Experia Box v10 (H369A)

To run the exporter as a Docker container, provide the `EXPERIA_V10_ROUTER_USERNAME` and `EXPERIA_V10_ROUTER_PASSWORD` environment variables and run

```
docker compose up -d
```

## Usage
```plain
$ experia-v10-exporter
```

The following environment variables are required:
```
EXPERIA_V10_LISTEN_ADDR=localhost:9684
EXPERIA_V10_TIMEOUT=10
EXPERIA_V10_ROUTER_IP=192.168.2.254
EXPERIA_V10_ROUTER_USERNAME=Admin
EXPERIA_V10_ROUTER_PASSWORD="PASSWORD"
```

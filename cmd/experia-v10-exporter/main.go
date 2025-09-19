package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/GrammaTonic/experia-v10-exporter/internal/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Setup prepares the collector and HTTP handlers and returns the listen address and the created collector.
// It does not start the HTTP server, allowing tests to call Setup without blocking.
func Setup() (string, *collector.Experiav10Collector, error) {
	listenAddr := os.Getenv("EXPERIA_V10_LISTEN_ADDR")
	timeout, err := time.ParseDuration(os.Getenv("EXPERIA_V10_TIMEOUT"))
	if err != nil {
		return "", nil, fmt.Errorf("EXPERIA_V10_TIMEOUT invalid: %w", err)
	}
	ip := net.ParseIP(os.Getenv("EXPERIA_V10_ROUTER_IP"))
	if ip == nil {
		return "", nil, fmt.Errorf("EXPERIA_V10_ROUTER_IP invalid")
	}
	username := os.Getenv("EXPERIA_V10_ROUTER_USERNAME")
	password := os.Getenv("EXPERIA_V10_ROUTER_PASSWORD")

	col := collector.NewCollector(ip, username, password, timeout)
	if err := prometheus.Register(col); err != nil {
		return "", nil, fmt.Errorf("Failed to register collector: %w", err)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", http.RedirectHandler("/metrics", http.StatusFound))

	return listenAddr, col, nil
}

func main() {
	if err := runMain(); err != nil {
		exitOnError(err)
		return
	}
}

// runMain contains the testable main logic and returns an error instead of exiting.
func runMain() error {
	listenAddr, _, err := Setup()
	if err != nil {
		return err
	}
	log.Printf("Listen on %s...", listenAddr)
	if err := listenAndServe(listenAddr, nil); err != nil {
		return err
	}
	return nil
}

// listenAndServe allows tests to override the real http.ListenAndServe.
var listenAndServe = http.ListenAndServe

// exitOnError is called when main needs to exit due to an error. Tests may override it to avoid exiting the process.
var exitOnError = func(err error) {
	if err != nil {
		// When running tests we can set EXPERIA_V10_TEST_MODE=1 to avoid exiting the process
		if os.Getenv("EXPERIA_V10_TEST_MODE") == "1" {
			log.Printf("exitOnError (test mode): %v", err)
			return
		}
		log.Fatal(err)
	}
}

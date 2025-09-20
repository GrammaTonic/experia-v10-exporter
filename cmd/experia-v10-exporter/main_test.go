package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/GrammaTonic/experia-v10-exporter/internal/collector"
	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
	"github.com/prometheus/client_golang/prometheus"
)

func TestSetupRegistersCollector(t *testing.T) {
	// Set minimal env vars
	os.Setenv("EXPERIA_V10_LISTEN_ADDR", ":0")
	os.Setenv("EXPERIA_V10_TIMEOUT", "1s")
	os.Setenv("EXPERIA_V10_ROUTER_IP", "192.0.2.5")
	os.Setenv("EXPERIA_V10_ROUTER_USERNAME", "user")
	os.Setenv("EXPERIA_V10_ROUTER_PASSWORD", "pass")
	defer func() {
		os.Unsetenv("EXPERIA_V10_LISTEN_ADDR")
		os.Unsetenv("EXPERIA_V10_TIMEOUT")
		os.Unsetenv("EXPERIA_V10_ROUTER_IP")
		os.Unsetenv("EXPERIA_V10_ROUTER_USERNAME")
		os.Unsetenv("EXPERIA_V10_ROUTER_PASSWORD")
	}()

	listen, col, err := Setup()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	if listen == "" {
		t.Fatalf("expected non-empty listen address")
	}

	// Unregister collector
	prometheus.Unregister(col)
}

func TestSetupDuplicateRegister(t *testing.T) {
	// Set minimal env vars
	os.Setenv("EXPERIA_V10_LISTEN_ADDR", ":0")
	os.Setenv("EXPERIA_V10_TIMEOUT", "1s")
	os.Setenv("EXPERIA_V10_ROUTER_IP", "192.0.2.8")
	os.Setenv("EXPERIA_V10_ROUTER_USERNAME", "user")
	os.Setenv("EXPERIA_V10_ROUTER_PASSWORD", "pass")
	defer func() {
		os.Unsetenv("EXPERIA_V10_LISTEN_ADDR")
		os.Unsetenv("EXPERIA_V10_TIMEOUT")
		os.Unsetenv("EXPERIA_V10_ROUTER_IP")
		os.Unsetenv("EXPERIA_V10_ROUTER_USERNAME")
		os.Unsetenv("EXPERIA_V10_ROUTER_PASSWORD")
	}()

	// Ensure fresh mux to avoid handler conflicts
	http.DefaultServeMux = http.NewServeMux()

	// First call should register successfully
	_, col, err := Setup()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Second call should fail due to duplicate registration
	http.DefaultServeMux = http.NewServeMux()
	_, _, err = Setup()
	if err == nil {
		// cleanup
		prometheus.Unregister(col)
		t.Fatalf("expected Setup to fail on duplicate register")
	}

	// cleanup
	prometheus.Unregister(col)
}

func TestMainListenAddrEmpty(t *testing.T) {
	// empty listen address case
	os.Setenv("EXPERIA_V10_LISTEN_ADDR", "")
	os.Setenv("EXPERIA_V10_TIMEOUT", "1s")
	os.Setenv("EXPERIA_V10_ROUTER_IP", "192.0.2.9")
	os.Setenv("EXPERIA_V10_ROUTER_USERNAME", "user")
	os.Setenv("EXPERIA_V10_ROUTER_PASSWORD", "pass")
	defer func() {
		os.Unsetenv("EXPERIA_V10_LISTEN_ADDR")
		os.Unsetenv("EXPERIA_V10_TIMEOUT")
		os.Unsetenv("EXPERIA_V10_ROUTER_IP")
		os.Unsetenv("EXPERIA_V10_ROUTER_USERNAME")
		os.Unsetenv("EXPERIA_V10_ROUTER_PASSWORD")
	}()

	// reset mux
	http.DefaultServeMux = http.NewServeMux()

	var gotAddr string
	origListen := listenAndServe
	listenAndServe = func(addr string, handler http.Handler) error {
		gotAddr = addr
		return nil
	}
	defer func() { listenAndServe = origListen }()

	main()
	if gotAddr != "" {
		t.Fatalf("expected listen addr to be empty, got %q", gotAddr)
	}

	// cleanup: unregister collector created by Setup
	col := collector.NewCollector(net.ParseIP("192.0.2.9"), os.Getenv("EXPERIA_V10_ROUTER_USERNAME"), os.Getenv("EXPERIA_V10_ROUTER_PASSWORD"), 1*time.Second)
	prometheus.Unregister(col)
}

func TestMainHandlerNil(t *testing.T) {
	// non-empty listen addr, ensure handler arg can be nil
	os.Setenv("EXPERIA_V10_LISTEN_ADDR", ":0")
	os.Setenv("EXPERIA_V10_TIMEOUT", "1s")
	os.Setenv("EXPERIA_V10_ROUTER_IP", "192.0.2.10")
	defer func() {
		os.Unsetenv("EXPERIA_V10_LISTEN_ADDR")
		os.Unsetenv("EXPERIA_V10_TIMEOUT")
		os.Unsetenv("EXPERIA_V10_ROUTER_IP")
	}()

	http.DefaultServeMux = http.NewServeMux()

	var gotHandlerNotNil bool
	origListen := listenAndServe
	listenAndServe = func(addr string, handler http.Handler) error {
		if handler != nil {
			gotHandlerNotNil = true
		}
		return nil
	}
	defer func() { listenAndServe = origListen }()

	main()

	if gotHandlerNotNil {
		t.Fatalf("expected handler passed to listenAndServe to be nil")
	}

	// cleanup: unregister collector created by Setup
	col := collector.NewCollector(net.ParseIP("192.0.2.10"), os.Getenv("EXPERIA_V10_ROUTER_USERNAME"), os.Getenv("EXPERIA_V10_ROUTER_PASSWORD"), 1*time.Second)
	prometheus.Unregister(col)
}

func TestExitOnErrorTestMode(t *testing.T) {
	// set test mode so exitOnError doesn't call log.Fatal
	os.Setenv("EXPERIA_V10_TEST_MODE", "1")
	defer os.Unsetenv("EXPERIA_V10_TEST_MODE")

	called := false
	// call exitOnError with a non-nil error; in test mode it should return without exiting
	exitOnError(fmt.Errorf("test error"))
	called = true
	if !called {
		t.Fatalf("exitOnError did not return in test mode")
	}
}

func TestRunMainSuccess(t *testing.T) {
	os.Setenv("EXPERIA_V10_LISTEN_ADDR", ":0")
	os.Setenv("EXPERIA_V10_TIMEOUT", "1s")
	os.Setenv("EXPERIA_V10_ROUTER_IP", "192.0.2.11")
	os.Setenv("EXPERIA_V10_ROUTER_USERNAME", "user")
	os.Setenv("EXPERIA_V10_ROUTER_PASSWORD", "pass")
	defer func() {
		os.Unsetenv("EXPERIA_V10_LISTEN_ADDR")
		os.Unsetenv("EXPERIA_V10_TIMEOUT")
		os.Unsetenv("EXPERIA_V10_ROUTER_IP")
		os.Unsetenv("EXPERIA_V10_ROUTER_USERNAME")
		os.Unsetenv("EXPERIA_V10_ROUTER_PASSWORD")
	}()

	http.DefaultServeMux = http.NewServeMux()
	origListen := listenAndServe
	listenAndServe = func(addr string, handler http.Handler) error { return nil }
	defer func() { listenAndServe = origListen }()

	if err := runMain(); err != nil {
		t.Fatalf("runMain failed: %v", err)
	}

	// cleanup
	col := collector.NewCollector(net.ParseIP("192.0.2.11"), os.Getenv("EXPERIA_V10_ROUTER_USERNAME"), os.Getenv("EXPERIA_V10_ROUTER_PASSWORD"), 1*time.Second)
	prometheus.Unregister(col)
}

func TestRunMainListenFail(t *testing.T) {
	os.Setenv("EXPERIA_V10_LISTEN_ADDR", ":0")
	os.Setenv("EXPERIA_V10_TIMEOUT", "1s")
	os.Setenv("EXPERIA_V10_ROUTER_IP", "192.0.2.12")
	defer func() {
		os.Unsetenv("EXPERIA_V10_LISTEN_ADDR")
		os.Unsetenv("EXPERIA_V10_TIMEOUT")
		os.Unsetenv("EXPERIA_V10_ROUTER_IP")
	}()

	http.DefaultServeMux = http.NewServeMux()
	origListen := listenAndServe
	listenAndServe = func(addr string, handler http.Handler) error { return &testutil.SimpleErr{S: "listen fail"} }
	defer func() { listenAndServe = origListen }()

	if err := runMain(); err == nil {
		t.Fatalf("expected runMain to return error when listen fails")
	}

	// cleanup
	col := collector.NewCollector(net.ParseIP("192.0.2.12"), os.Getenv("EXPERIA_V10_ROUTER_USERNAME"), os.Getenv("EXPERIA_V10_ROUTER_PASSWORD"), 1*time.Second)
	prometheus.Unregister(col)
}

func TestMainEntry(t *testing.T) {
	// set env for Setup
	os.Setenv("EXPERIA_V10_LISTEN_ADDR", ":0")
	os.Setenv("EXPERIA_V10_TIMEOUT", "1s")
	os.Setenv("EXPERIA_V10_ROUTER_IP", "192.0.2.6")
	os.Setenv("EXPERIA_V10_ROUTER_USERNAME", "user")
	os.Setenv("EXPERIA_V10_ROUTER_PASSWORD", "pass")
	defer func() {
		os.Unsetenv("EXPERIA_V10_LISTEN_ADDR")
		os.Unsetenv("EXPERIA_V10_TIMEOUT")
		os.Unsetenv("EXPERIA_V10_ROUTER_IP")
		os.Unsetenv("EXPERIA_V10_ROUTER_USERNAME")
		os.Unsetenv("EXPERIA_V10_ROUTER_PASSWORD")
	}()

	// Reset default mux to avoid duplicate handler registration from previous tests
	http.DefaultServeMux = http.NewServeMux()

	// Override listenAndServe to avoid binding ports
	orig := listenAndServe
	defer func() { listenAndServe = orig }()
	listenAndServe = func(addr string, handler http.Handler) error { return nil }

	// call main (it will call Setup and then our stubbed listenAndServe)
	main()
}

func TestSetupErrorPaths(t *testing.T) {
	// Invalid timeout
	os.Setenv("EXPERIA_V10_TIMEOUT", "notaduration")
	os.Setenv("EXPERIA_V10_ROUTER_IP", "192.0.2.6")
	defer func() { os.Unsetenv("EXPERIA_V10_TIMEOUT") }()
	// call Setup directly which should return an error
	_, _, err := Setup()
	if err == nil {
		t.Fatalf("expected Setup to return error for invalid timeout")
	}

	// Reset and test invalid IP
	os.Setenv("EXPERIA_V10_TIMEOUT", "1s")
	os.Setenv("EXPERIA_V10_ROUTER_IP", "not-an-ip")
	defer func() {
		os.Unsetenv("EXPERIA_V10_TIMEOUT")
		os.Unsetenv("EXPERIA_V10_ROUTER_IP")
	}()
	_, _, err = Setup()
	if err == nil {
		t.Fatalf("expected Setup to return error for invalid IP")
	}
}

func TestMainExitPaths(t *testing.T) {
	// Test Setup error path via main()
	os.Setenv("EXPERIA_V10_TIMEOUT", "bad")
	os.Setenv("EXPERIA_V10_ROUTER_IP", "192.0.2.6")
	defer func() {
		os.Unsetenv("EXPERIA_V10_TIMEOUT")
		os.Unsetenv("EXPERIA_V10_ROUTER_IP")
	}()

	called := false
	origExit := exitOnError
	exitOnError = func(err error) { called = true }
	defer func() { exitOnError = origExit }()

	main()
	if !called {
		t.Fatalf("expected exitOnError to be called when Setup fails in main")
	}

	// Test listenAndServe error path
	// Prepare valid env
	os.Setenv("EXPERIA_V10_TIMEOUT", "1s")
	os.Setenv("EXPERIA_V10_ROUTER_IP", "192.0.2.7")
	defer func() {
		os.Unsetenv("EXPERIA_V10_TIMEOUT")
		os.Unsetenv("EXPERIA_V10_ROUTER_IP")
	}()

	// reset mux
	http.DefaultServeMux = http.NewServeMux()

	origListen := listenAndServe
	listenAndServe = func(addr string, handler http.Handler) error { return &testutil.SimpleErr{S: "listen fail"} }
	defer func() { listenAndServe = origListen }()

	called = false
	origExit = exitOnError
	exitOnError = func(err error) { called = true }
	defer func() { exitOnError = origExit }()

	main()
	if !called {
		t.Fatalf("expected exitOnError to be called when listenAndServe fails in main")
	}

	// cleanup: unregister collector by creating a similar one and unregistering
	// use same env values used by Setup
	col := func() *collector.Experiav10Collector {
		ip := net.ParseIP("192.0.2.7")
		return collector.NewCollector(ip, os.Getenv("EXPERIA_V10_ROUTER_USERNAME"), os.Getenv("EXPERIA_V10_ROUTER_PASSWORD"), 1*time.Second)
	}()
	prometheus.Unregister(col)
}

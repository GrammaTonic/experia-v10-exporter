package main

import (
	"os"
	"testing"
)

func TestSetup_WithEnv(t *testing.T) {
	os.Setenv("EXPERIA_V10_TIMEOUT", "1s")
	os.Setenv("EXPERIA_V10_ROUTER_IP", "127.0.0.1")
	os.Setenv("EXPERIA_V10_ROUTER_USERNAME", "u")
	os.Setenv("EXPERIA_V10_ROUTER_PASSWORD", "p")
	defer func() {
		_ = os.Unsetenv("EXPERIA_V10_TIMEOUT")
		_ = os.Unsetenv("EXPERIA_V10_ROUTER_IP")
		_ = os.Unsetenv("EXPERIA_V10_ROUTER_USERNAME")
		_ = os.Unsetenv("EXPERIA_V10_ROUTER_PASSWORD")
	}()

	_, col, err := Setup()
	if err != nil {
		t.Fatalf("Setup returned error: %v", err)
	}
	if col == nil {
		t.Fatalf("expected collector instance")
	}
	// unregister to avoid global state in other tests
	_ = collector
}

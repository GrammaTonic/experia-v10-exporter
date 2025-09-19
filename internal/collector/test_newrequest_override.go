//go:build test
// +build test

package collector

import "net/http"

// test build override: default to http.NewRequest but allows tests to replace it
var newRequest = http.NewRequest

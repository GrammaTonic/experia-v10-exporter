//go:build !test
// +build !test

package collector

import "net/http"

// newRequest is an indirection to allow tests to simulate http.NewRequest failures or malformed requests.
var newRequest = http.NewRequest

//go:build test
// +build test

package collector

import (
	"context"
	"net"
	"testing"
)

// TestAuthenticateNewRequestError forces an invalid apiUrl to trigger http.NewRequest error in authenticate
func TestAuthenticateNewRequestError(t *testing.T) {
	old := apiUrl
	defer func() { apiUrl = old }()
	apiUrl = "http://\x00%s/invalid"

	c := NewCollector(net.ParseIP("192.0.2.30"), "u", "p", 1)
	if _, err := c.authenticate(); err == nil {
		t.Fatalf("expected authenticate to fail when apiUrl is invalid")
	}
}

// TestFetchURLInvalidURL forces fetchURL to fail on NewRequest by passing an invalid URL
func TestFetchURLInvalidURL(t *testing.T) {
	c := NewCollector(nil, "", "", 1)
	if _, err := c.fetchURL(context.Background(), "GET", "http://\x00/", nil, nil); err == nil {
		t.Fatalf("expected fetchURL to fail for invalid URL")
	}
}

func TestAuthenticateJSONMarshalError(t *testing.T) {
	old := jsonMarshal
	defer func() { jsonMarshal = old }()
	jsonMarshal = func(v any) ([]byte, error) { return nil, &simpleErr{"marshal fail"} }

	c := NewCollector(net.ParseIP("192.0.2.40"), "u", "p", 1)
	if _, err := c.authenticate(); err == nil {
		t.Fatalf("expected authenticate to fail when json.Marshal errors")
	}
}

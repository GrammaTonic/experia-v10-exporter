package collector

import (
    "io"
    "net/http"
    "strings"
)

// roundTripperFunc is a small helper to stub HTTP responses in tests.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// makeResp creates an *http.Response with the provided body and status 200.
func makeResp(body string) *http.Response {
    return &http.Response{
        StatusCode: 200,
        Body:       io.NopCloser(strings.NewReader(body)),
        Header:     http.Header{"Content-Type": {"application/json"}},
    }
}

// sampleMibJSON is a compact sample used by multiple tests. Keep it here so
// tests reuse the same fixture.
const sampleMibJSON = `{"status":true,"data":{"status":{"base":{"ETH2":{"Name":"ETH2","Enable":true,"Status":true,"Flags":"enabled netdev vlan","LLAddress":"88:D2:74:AB:05:D0","TxQueueLen":0,"MTU":1500,"NetDevState":"up","CurrentBitRate":1000,"LastChangeTime":75},"eth3":{"Name":"eth3","Enable":true,"Status":true,"Flags":"up broadcast multicast","LLAddress":"88:D2:74:AB:05:D0","TxQueueLen":1000,"MTU":1500,"NetDevState":"up","CurrentBitRate":1000,"LastChangeTime":73}},"alias":{"ETH2":{"Alias":"cpe-ETH2"},"eth3":{"Alias":"cpe-eth3"}}}}}`

// makeJSONHandler returns an http.Handler that responds with the provided
// body for any request. Useful for creating httptest servers in tests.
func makeJSONHandler(body string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        _, _ = w.Write([]byte(body))
    })
}

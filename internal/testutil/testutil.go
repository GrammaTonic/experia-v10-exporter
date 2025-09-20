package testutil

import (
	"io"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// RoundTripperFunc is a small helper to stub HTTP responses in tests.
type RoundTripperFunc func(*http.Request) (*http.Response, error)

// roundTripperFunc is an exported alias matching previous tests' expectations.
type roundTripperFunc = RoundTripperFunc

func (f RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// MakeRoundTripper converts a simple func into an http.RoundTripper
func MakeRoundTripper(f func(*http.Request) (*http.Response, error)) http.RoundTripper {
	return RoundTripperFunc(f)
}

// MakeResp creates an *http.Response with the provided body and status 200.
func MakeResp(body string) *http.Response {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": {"application/json"}},
	}
}

// SampleMibJSON is a compact sample used by multiple tests.
const SampleMibJSON = `{"status":true,"data":{"status":{"base":{"ETH2":{"Name":"ETH2","Enable":true,"Status":true,"Flags":"enabled netdev vlan","LLAddress":"88:D2:74:AB:05:D0","TxQueueLen":0,"MTU":1500,"NetDevState":"up","CurrentBitRate":1000,"LastChangeTime":75},"eth3":{"Name":"eth3","Enable":true,"Status":true,"Flags":"up broadcast multicast","LLAddress":"88:D2:74:AB:05:D0","TxQueueLen":1000,"MTU":1500,"NetDevState":"up","CurrentBitRate":1000,"LastChangeTime":73}},"alias":{"ETH2":{"Alias":"cpe-ETH2"},"eth3":{"Alias":"cpe-eth3"}}}}}`

// MakeJSONHandler returns an http.Handler that responds with the provided
// body for any request. Useful for creating httptest servers in tests.
func MakeJSONHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	})
}

// RewriteTransport rewrites all requests to the baseURL while preserving the method and headers/body.
func RewriteTransport(baseURL string) http.RoundTripper {
	return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		// Build new request to test server
		newReq, err := http.NewRequest(req.Method, baseURL, req.Body)
		if err != nil {
			return nil, err
		}
		newReq.Header = req.Header.Clone()
		// Use default transport
		return http.DefaultTransport.RoundTrip(newReq)
	})
}

// ReadCounterValue reads the current value of a prometheus.Counter (via its Desc/Collect path)
func ReadCounterValue(c prometheus.Counter) float64 {
	ch := make(chan prometheus.Metric, 1)
	go func() {
		c.Collect(ch)
		close(ch)
	}()
	for m := range ch {
		pm := &dto.Metric{}
		if err := m.Write(pm); err == nil {
			if pm.Counter != nil {
				return pm.Counter.GetValue()
			}
		}
	}
	return 0
}

// FmtError returns a test error value.
func FmtError(s string) error { return &SimpleErr{s} }

// ErrReadCloser is an io.ReadCloser that returns an error on Read.
type ErrReadCloser struct{}

// SimpleErr is a small error type used by tests to simulate failures.
type SimpleErr struct{ S string }

func (e *SimpleErr) Error() string { return e.S }

func (e *ErrReadCloser) Read(b []byte) (int, error) { return 0, &SimpleErr{"read error"} }
func (e *ErrReadCloser) Close() error               { return nil }

// SampleStatsFmt is a template for a getNetDevStats response where numeric
// fields are provided so tests can assert emitted metric values.
const SampleStatsFmt = `{"status":true,"data":{` +
	`"RxPackets":%d,"TxPackets":%d,"RxBytes":%d,"TxBytes":%d,` +
	`"RxErrors":%d,"TxErrors":%d,"RxDropped":%d,"TxDropped":%d,` +
	`"Multicast":%d,"Collisions":%d,` +
	`"RxLengthErrors":%d,"RxOverErrors":%d,"RxCrcErrors":%d,` +
	`"RxFrameErrors":%d,"RxFifoErrors":%d,"RxMissedErrors":%d,` +
	`"TxAbortedErrors":%d,"TxCarrierErrors":%d,"TxFifoErrors":%d,` +
	`"TxHeartbeatErrors":%d,"TxWindowErrors":%d}}`

package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
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

// rewriteTransport rewrites all requests to the baseURL while preserving the method and headers/body.
func rewriteTransport(baseURL string) http.RoundTripper {
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
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

// readCounterValue reads the current value of a prometheus.Counter (via its Desc/Collect path)
func readCounterValue(c prometheus.Counter) float64 {
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

func fmtError(s string) error { return &simpleErr{s} }

type errReadCloser struct{}

// simpleErr is a small error type used by tests to simulate failures.
type simpleErr struct{ s string }

func (e *simpleErr) Error() string { return e.s }

func (e *errReadCloser) Read(b []byte) (int, error) { return 0, &simpleErr{"read error"} }
func (e *errReadCloser) Close() error               { return nil }

// sampleStatsFmt is a template for a getNetDevStats response where numeric
// fields are provided so tests can assert emitted metric values.
const sampleStatsFmt = `{"status":true,"data":{` +
	`"RxPackets":%d,"TxPackets":%d,"RxBytes":%d,"TxBytes":%d,` +
	`"RxErrors":%d,"TxErrors":%d,"RxDropped":%d,"TxDropped":%d,` +
	`"Multicast":%d,"Collisions":%d,` +
	`"RxLengthErrors":%d,"RxOverErrors":%d,"RxCrcErrors":%d,` +
	`"RxFrameErrors":%d,"RxFifoErrors":%d,"RxMissedErrors":%d,` +
	`"TxAbortedErrors":%d,"TxCarrierErrors":%d,"TxFifoErrors":%d,` +
	`"TxHeartbeatErrors":%d,"TxWindowErrors":%d}}`

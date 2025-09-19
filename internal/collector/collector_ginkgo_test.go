package collector

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var _ = Describe("Experia Collector", func() {
	It("authenticates and collects WAN status metric", func() {
		// Test server to handle createContext and getWANStatus
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var resp interface{}
			// Simple switch by request body content
			if r.Header.Get("Authorization") == "X-Sah-Login" {
				// createContext
				resp = map[string]interface{}{"data": map[string]string{"contextID": "token-123"}}
			} else {
				// getWANStatus
				resp = map[string]interface{}{
					"status": true,
					"data": map[string]interface{}{
						"LinkType":        "ETH",
						"LinkState":       "Up",
						"MACAddress":      "aa:bb:cc:dd:ee:ff",
						"Protocol":        "DHCP",
						"ConnectionState": "Connected",
						"IPAddress":       "1.2.3.4",
					},
				}
			}
			b, _ := json.Marshal(resp)
			w.WriteHeader(http.StatusOK)
			w.Write(b)
		}))
		defer ts.Close()

		// Create collector and inject client that rewrites requests to our test server.
		ip := net.ParseIP("192.0.2.1")
		c := NewCollector(ip, "user", "pass", 2*time.Second)

		// Replace client's Transport to point to test server by modifying requests' URL on the fly.
		c.client.Transport = rewriteTransport(ts.URL)

		// Collect metrics into buffer channel
		ch := make(chan prometheus.Metric, 10)
		go func() {
			c.Collect(ch)
			close(ch)
		}()

		// Drain metrics and find internet_connection metric
		var found bool
		for m := range ch {
			desc := m.Desc().String()
			if strings.Contains(desc, "internet_connection") || desc == ifupTime.String() {
				// convert to dto and inspect value
				pm := &dto.Metric{}
				Expect(m.Write(pm)).To(Succeed())
				Expect(pm.GetGauge()).ToNot(BeNil())
				Expect(pm.GetGauge().GetValue()).To(Equal(1.0))
				found = true
			}
		}
		Expect(found).To(BeTrue())
	})

	It("handles multiple failure scenarios (table-driven)", func() {
		tests := []struct {
			name           string
			serverHandler  http.HandlerFunc
			setupTransport func(c *Experiav10Collector, tsURL string)
			validate       func(c *Experiav10Collector, ch chan prometheus.Metric)
		}{
			{
				name: "auth failure exports up=0 and auth error counter",
				serverHandler: func(w http.ResponseWriter, r *http.Request) {
					resp := map[string]interface{}{"data": map[string]string{"contextID": ""}}
					b, _ := json.Marshal(resp)
					w.WriteHeader(http.StatusOK)
					w.Write(b)
				},
				setupTransport: func(c *Experiav10Collector, tsURL string) { c.client.Transport = rewriteTransport(tsURL) },
				validate: func(c *Experiav10Collector, ch chan prometheus.Metric) {
					// Expect to receive up metric (gauge=0) and auth error counter metric
					var sawUp bool
					var sawAuthErr bool
					for m := range ch {
						pm := &dto.Metric{}
						Expect(m.Write(pm)).To(Succeed())
						if pm.Gauge != nil {
							Expect(pm.Gauge.GetValue()).To(Equal(0.0))
							sawUp = true
						}
						if pm.Counter != nil {
							if pm.Counter.GetValue() >= 1.0 {
								sawAuthErr = true
							}
						}
					}
					Expect(sawUp).To(BeTrue())
					Expect(sawAuthErr).To(BeTrue())
				},
			},
			{
				name: "permission denied increments permissionErrors",
				serverHandler: func(w http.ResponseWriter, r *http.Request) {
					var resp interface{}
					if r.Header.Get("Authorization") == "X-Sah-Login" {
						resp = map[string]interface{}{"data": map[string]string{"contextID": "tok-perm"}}
					} else {
						resp = map[string]interface{}{
							"status": false,
							"errors": []map[string]interface{}{{"error": 1, "description": "Permission denied", "info": ""}},
						}
					}
					b, _ := json.Marshal(resp)
					w.WriteHeader(http.StatusOK)
					w.Write(b)
				},
				setupTransport: func(c *Experiav10Collector, tsURL string) { c.client.Transport = rewriteTransport(tsURL) },
				validate: func(c *Experiav10Collector, ch chan prometheus.Metric) {
					// we already incremented permissionErrors via Collect; verify counter increased by 1
					// note: permissionErrors is a package-level counter
					// drain channel first
					for range ch {
					}
					// read via helper
					after := readCounterValue(permissionErrors)
					Expect(after).To(BeNumerically(
						">=", 1.0))
				},
			},
			{
				name: "scrape error increments scrapeErrorsMetric",
				serverHandler: func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("Authorization") == "X-Sah-Login" {
						resp := map[string]interface{}{"data": map[string]string{"contextID": "tok-scrape"}}
						b, _ := json.Marshal(resp)
						w.WriteHeader(http.StatusOK)
						w.Write(b)
						return
					}
					resp := map[string]interface{}{"status": true, "data": map[string]interface{}{"ConnectionState": "Disconnected"}}
					b, _ := json.Marshal(resp)
					w.WriteHeader(http.StatusOK)
					w.Write(b)
				},
				setupTransport: func(c *Experiav10Collector, tsURL string) {
					// allow auth but simulate network error for subsequent fetches
					c.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
						if req.Header.Get("Authorization") == "X-Sah-Login" {
							newReq, _ := http.NewRequest(req.Method, tsURL, req.Body)
							newReq.Header = req.Header.Clone()
							return http.DefaultTransport.RoundTrip(newReq)
						}
						return nil, fmtError("simulated network error")
					})
				},
				validate: func(c *Experiav10Collector, ch chan prometheus.Metric) {
					for range ch {
					}
					after := readCounterValue(c.scrapeErrorsMetric)
					Expect(after).To(BeNumerically(
						">=", 1.0))
				},
			},
		}

		for _, tc := range tests {
			By(tc.name)
			ts := httptest.NewServer(tc.serverHandler)
			func() {
				defer ts.Close()
				ip := net.ParseIP("192.0.2.2")
				c := NewCollector(ip, "user", "pass", 2*time.Second)
				if tc.setupTransport != nil {
					tc.setupTransport(c, ts.URL)
				} else {
					c.client.Transport = rewriteTransport(ts.URL)
				}

				ch := make(chan prometheus.Metric, 10)
				go func() {
					c.Collect(ch)
					close(ch)
				}()
				tc.validate(c, ch)
			}()
		}
	})
})

// contains performs a substring check
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (stringIndex(s, sub) >= 0)))
}

// stringIndex is a simple implementation of strings.Index to avoid importing strings
func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
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

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

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

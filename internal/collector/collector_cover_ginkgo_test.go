package collector

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("collector cover additional", func() {
	It("authenticate returns error when contextID empty", func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{"contextID":""}}`))
		}))
		defer ts.Close()

		c := NewCollector(net.ParseIP("192.0.2.30"), "u", "p", 1*time.Second)
		c.client.Transport = rewriteTransport(ts.URL)
		_, err := c.authenticate()
		Expect(err).To(HaveOccurred())
	})

	It("fetchURL handles header and no-header variants", func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Variant") == "1" {
				w.Write([]byte("v1"))
				return
			}
			w.Write([]byte("base"))
		}))
		defer ts.Close()

		c := NewCollector(net.ParseIP("192.0.2.31"), "u", "p", 1*time.Second)
		c.client.Transport = rewriteTransport(ts.URL)

		b, err := c.fetchURL(context.Background(), "GET", "http://example/", nil, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(b)).To(Equal("base"))

		b, err = c.fetchURL(context.Background(), "GET", "http://example/", map[string]string{"X-Variant": "1"}, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(b)).To(Equal("v1"))
	})
})

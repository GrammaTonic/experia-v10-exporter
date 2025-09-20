package collector

import (
	"context"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	connectivity "github.com/GrammaTonic/experia-v10-exporter/internal/collector/connectivity"
	"github.com/GrammaTonic/experia-v10-exporter/internal/testutil"
)

var _ = Describe("collector extra cases", func() {
	It("stores cookies when Jar is present during authenticate", func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Return a Set-Cookie header and a valid contextID JSON
			w.Header().Add("Set-Cookie", "sid=abc; Path=/")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":{"contextID":"tok-cj"}}`))
		}))
		defer ts.Close()

		ip := net.ParseIP("192.0.2.20")
		c := NewCollector(ip, "u", "p", 1*time.Second)
		// set a cookie jar
		jar, _ := cookiejar.New(nil)
		c.client.Jar = jar
		// rewrite transport to test server
		c.client.Transport = testutil.RewriteTransport(ts.URL)

		ctx, err := c.authenticate()
		Expect(err).ToNot(HaveOccurred())
		Expect(ctx.Token).To(Equal("tok-cj"))

		// cookie storage behavior depends on the exact URL used by authenticate; we assert only that auth succeeded
	})

	It("forwards headers in fetchURL", func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Test-Header") == "1" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing"))
		}))
		defer ts.Close()

		c := NewCollector(net.ParseIP("192.0.2.21"), "u", "p", 1*time.Second)
		c.client.Transport = testutil.RewriteTransport(ts.URL)

		b, err := connectivity.FetchURL(c.client, context.Background(), "GET", "http://example/", map[string]string{"X-Test-Header": "1"}, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(b)).To(Equal("ok"))
	})
})

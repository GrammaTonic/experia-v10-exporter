// Fetch eth0 interface stats and log the response
package collector

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const metricPrefix = "experia_v10_"

// apiUrl is provided via build-tag files (production in apiurl_prod.go and test override when running
// tests with -tags test).

type sessionContext struct {
	Token string
}

type Experiav10Collector struct {
	ip                 net.IP
	username           string
	password           string
	client             *http.Client
	upMetric           prometheus.Gauge
	authErrorsMetric   prometheus.Counter
	scrapeErrorsMetric prometheus.Counter
}

func NewCollector(ip net.IP, username, password string, timeout time.Duration) *Experiav10Collector {
	cookieJar, _ := cookiejar.New(nil)

	return &Experiav10Collector{
		ip:       ip,
		username: username,
		password: password,
		client: &http.Client{
			Timeout: timeout,
			Jar:     cookieJar,
		},
		upMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: metricPrefix + "up",
			Help: "Shows if the Experia Box V10 is deemed up by the collector.",
		}),
		authErrorsMetric: prometheus.NewCounter(prometheus.CounterOpts{
			Name: metricPrefix + "auth_errors_total",
			Help: "Counts number of authentication errors encountered by the collector.",
		}),
		scrapeErrorsMetric: prometheus.NewCounter(prometheus.CounterOpts{
			Name: metricPrefix + "scrape_errors_total",
			Help: "Counts the number of scrape errors by this collector.",
		}),
	}
}

func (c *Experiav10Collector) Collect(ch chan<- prometheus.Metric) {
	ctx, err := c.authenticate()
	if err != nil {
		c.authErrorsMetric.Inc()
		c.upMetric.Set(0)
		c.upMetric.Collect(ch)
		c.authErrorsMetric.Collect(ch)
		c.scrapeErrorsMetric.Collect(ch)
		return
	}

	// Helper for POST fetches, returns response as string
	postFetch := func(body string) string {
		url := fmt.Sprintf(apiUrl, c.ip.String())
		headers := map[string]string{
			"accept":          "*/*",
			"accept-language": "en-US,en;q=0.7",
			"content-type":    "application/x-sah-ws-4-call+json",
			"sec-gpc":         "1",
			"Authorization":   "X-Sah " + ctx.Token,
			"x-context":       ctx.Token,
			// Add browser-matching headers for POSTs
			"Origin":  "http://192.168.2.254",
			"Referer": "http://192.168.2.254/",
		}
		resp, err := c.fetchURL("POST", url, headers, []byte(body))
		if err != nil {
			fmt.Printf("Failed to fetch %s: %v\n", body, err)
			c.scrapeErrorsMetric.Inc()
			return ""
		}
		return string(resp)
	}

	// Fetch getWANStatus and export metrics
	wanStatusResp := postFetch(`{"service":"NMC","method":"getWANStatus","parameters":{}}`)
	fmt.Printf("DEBUG: getWANStatus response: %s\n", wanStatusResp)
	var wanStatus struct {
		Status bool `json:"status"`
		Data   struct {
			LinkType        string `json:"LinkType"`
			LinkState       string `json:"LinkState"`
			MACAddress      string `json:"MACAddress"`
			Protocol        string `json:"Protocol"`
			ConnectionState string `json:"ConnectionState"`
			IPAddress       string `json:"IPAddress"`
		} `json:"data"`
		Errors []struct {
			Error       int    `json:"error"`
			Description string `json:"description"`
			Info        string `json:"info"`
		} `json:"errors"`
	}
	if wanStatusResp != "" {
		if err := json.Unmarshal([]byte(wanStatusResp), &wanStatus); err == nil {
			if len(wanStatus.Errors) > 0 {
				for _, e := range wanStatus.Errors {
					if e.Description == "Permission denied" {
						permissionErrors.Inc()
					}
				}
			} else if wanStatus.Status {
				// Export as gauge: 1 if connected, 0 otherwise
				val := 0.0
				if wanStatus.Data.ConnectionState == "Connected" {
					val = 1.0
				}
				ch <- prometheus.MustNewConstMetric(
					ifupTime,
					prometheus.GaugeValue,
					val,
					wanStatus.Data.LinkType,
					wanStatus.Data.Protocol,
					wanStatus.Data.ConnectionState,
					wanStatus.Data.IPAddress,
					wanStatus.Data.MACAddress,
				)
			}
		}
	}
}

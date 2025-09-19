// Fetch eth0 interface stats and log the response
package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"

	"strings"
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
		// Even when authentication fails, emit a placeholder internet_connection
		// metric so that registry.Gather() returns the metric family. This keeps
		// the behavior consistent for tests and scrapers that expect the family
		// to always be present.
		ch <- prometheus.MustNewConstMetric(
			ifupTime,
			prometheus.GaugeValue,
			0.0,
			"", "", "Unknown", "", "",
		)
		return
	}

	// whether we're running E2E/debug mode
	e2e := os.Getenv("EXPERIA_E2E") == "1"
	debugLog := func(format string, args ...interface{}) {
		if e2e {
			log.Printf(format, args...)
		}
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
			log.Printf("ERROR: failed to fetch %s: %v", body, err)
			c.scrapeErrorsMetric.Inc()
			return ""
		}
		return string(resp)
	}

	// Fetch getWANStatus and export metrics
	wanStatusResp := postFetch(`{"service":"NMC","method":"getWANStatus","parameters":{}}`)
	debugLog("DEBUG: getWANStatus response length=%d", len(wanStatusResp))
	// When running an E2E invocation, print the raw WAN JSON so we can
	// diagnose firmware variations in the JSON schema. debugLog already
	// checks EXPERIA_E2E, so call it directly.
	debugLog("DEBUG RAW getWANStatus: %s", wanStatusResp)
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
	emittedWAN := false
	if wanStatusResp != "" {
		if err := json.Unmarshal([]byte(wanStatusResp), &wanStatus); err == nil {
			if len(wanStatus.Errors) > 0 {
				for _, e := range wanStatus.Errors {
					if e.Description == "Permission denied" {
						permissionErrors.Inc()
					}
				}
			}
			// Emit the metric even if Status=false so the family is always present.
			val := 0.0
			if wanStatus.Status && wanStatus.Data.ConnectionState == "Connected" {
				val = 1.0
			}
			connState := wanStatus.Data.ConnectionState
			if connState == "" {
				connState = "Unknown"
			}
			ch <- prometheus.MustNewConstMetric(
				ifupTime,
				prometheus.GaugeValue,
				val,
				wanStatus.Data.LinkType,
				wanStatus.Data.Protocol,
				connState,
				wanStatus.Data.IPAddress,
				wanStatus.Data.MACAddress,
			)
			emittedWAN = true
		}
	}

	// If we didn't get a usable WAN status, emit an explicit placeholder so the
	// metric family is always present for Gather() consumers (tests, scrapers).
	if !emittedWAN {
		ch <- prometheus.MustNewConstMetric(
			ifupTime,
			prometheus.GaugeValue,
			0.0,
			"", "", "Unknown", "", "",
		)
	}

	// Fetch MIBs per-interface using NeMo.Intf.<IF> services. We request using
	// uppercase interface names in the service (NeMo.Intf.ETH0) but we expose
	// metrics with lowercase ifname labels (eth0) so that consumers see the
	// expected Linux-like names.

	// Candidate interfaces (uppercase for service name); allow override via
	// EXPERIA_EXPECT_NETDEV_IFACES (comma-separated, will be upper-cased here).
	candidates := []string{"ETH0", "ETH1", "ETH2", "ETH3"}
	if env := os.Getenv("EXPERIA_EXPECT_NETDEV_IFACES"); env != "" {
		candidates = nil
		for _, p := range strings.Split(env, ",") {
			pp := strings.TrimSpace(p)
			if pp != "" {
				candidates = append(candidates, strings.ToUpper(pp))
			}
		}
	}

	// sanitizeKey trims surrounding whitespace, quotes and backslashes and
	// lowercases the key so we can use a single normalized form everywhere.
	sanitizeKey := func(s string) string {
		if s == "" {
			return ""
		}
		// Trim spaces then remove surrounding quote and backslash characters.
		t := strings.TrimSpace(s)
		// Remove surrounding quote or backslash characters (both ends).
		trimChar := func(r byte) bool {
			return r == '"' || r == '\\' || r == '\''
		}
		for len(t) > 0 && trimChar(t[0]) {
			t = t[1:]
		}
		for len(t) > 0 && trimChar(t[len(t)-1]) {
			t = t[:len(t)-1]
		}
		return strings.ToLower(t)
	}

	// Process each candidate immediately and emit metrics per-interface.
	for _, cand := range candidates {
		body := fmt.Sprintf(`{"service":"NeMo.Intf.%s","method":"getMIBs","parameters":{}}`, cand)
		resp := postFetch(body)
		debugLog("DEBUG: getMIBs service=%s response length=%d", cand, len(resp))
		// debugLog is gated by EXPERIA_E2E; print raw JSON when enabled.
		debugLog("DEBUG RAW getMIBs service=%s: %s", cand, resp)
		if resp == "" {
			// still emit a zeroed metric so that collectors see the family when
			// the device doesn't return useful data for a candidate.
			canon := sanitizeKey(cand)
			ch <- prometheus.MustNewConstMetric(netdevUp, prometheus.GaugeValue, 0.0, canon)
			ch <- prometheus.MustNewConstMetric(netdevMtu, prometheus.GaugeValue, 0.0, canon)
			ch <- prometheus.MustNewConstMetric(netdevTxQueueLen, prometheus.GaugeValue, 0.0, canon)
			ch <- prometheus.MustNewConstMetric(netdevSpeedMbps, prometheus.GaugeValue, 0.0, canon)
			ch <- prometheus.MustNewConstMetric(netdevLastChange, prometheus.GaugeValue, 0.0, canon)
			ch <- prometheus.MustNewConstMetric(netdevInfo, prometheus.GaugeValue, 1.0, canon, "", "", "", "")
			continue
		}
		var g map[string]any
		if err := json.Unmarshal([]byte(resp), &g); err != nil {
			continue
		}
		// findStatus scans a decoded JSON map for a child that contains 'base' or 'netdev'.
		var findStatus func(m map[string]any) map[string]any
		findStatus = func(m map[string]any) map[string]any {
			if m == nil {
				return nil
			}
			if _, ok := m["base"]; ok {
				return m
			}
			if _, ok := m["netdev"]; ok {
				return m
			}
			for _, v := range m {
				if child, ok := v.(map[string]any); ok {
					if found := findStatus(child); found != nil {
						return found
					}
				}
			}
			return nil
		}

		var s map[string]any
		if ss := findStatus(g); ss != nil {
			s = ss
		} else if data, ok := g["data"].(map[string]any); ok {
			if ss := findStatus(data); ss != nil {
				s = ss
			}
		}
		if s == nil {
			if top, ok := g["status"].(map[string]any); ok {
				if ss := findStatus(top); ss != nil {
					s = ss
				} else {
					s = top
				}
			}
		}
		if s == nil {
			continue
		}

		// Merge base and netdev entries into a single inner map for this
		// candidate. We try to find a matching inner map by name and fall back
		// to using the candidate name as the canonical label.
		var innerMap map[string]any
		if b, ok := s["base"].(map[string]any); ok {
			// Try finding an inner map that matches candidate
			for _, v := range b {
				if im, ok := v.(map[string]any); ok {
					if n, ok := im["Name"].(string); ok && sanitizeKey(n) == sanitizeKey(cand) {
						innerMap = im
						break
					}
					if n, ok := im["NetDevName"].(string); ok && sanitizeKey(n) == sanitizeKey(cand) {
						innerMap = im
						break
					}
				}
			}
			// if not found, and there is only a single entry, pick it
			if innerMap == nil && len(b) == 1 {
				for _, v := range b {
					if im, ok := v.(map[string]any); ok {
						innerMap = im
						break
					}
				}
			}
		}
		if innerMap == nil {
			if nd, ok := s["netdev"].(map[string]any); ok {
				for _, v := range nd {
					if im, ok := v.(map[string]any); ok {
						if n, ok := im["NetDevIfName"].(string); ok && sanitizeKey(n) == sanitizeKey(cand) {
							innerMap = im
							break
						}
						if n, ok := im["NetDevName"].(string); ok && sanitizeKey(n) == sanitizeKey(cand) {
							innerMap = im
							break
						}
					}
				}
				if innerMap == nil && len(nd) == 1 {
					for _, v := range nd {
						if im, ok := v.(map[string]any); ok {
							innerMap = im
							break
						}
					}
				}
			}
		}

		// Build a normalized key/value map from innerMap for easy lookups.
		norm := map[string]any{}
		for k, v := range innerMap {
			norm[strings.ToLower(k)] = v
		}

		// If there is a netdev section, try to find a matching netdev entry
		// that corresponds to this candidate/innerMap and merge its fields
		// (prefer netdev values for MTU/Tx, etc.). This handles firmware
		// that splits base and netdev data across sections.
		if nd, ok := s["netdev"].(map[string]any); ok {
			// build a set of possible names to match against
			names := map[string]struct{}{}
			names[sanitizeKey(cand)] = struct{}{}
			if innerMap != nil {
				if n, ok := innerMap["Name"].(string); ok && n != "" {
					names[sanitizeKey(n)] = struct{}{}
				}
				if n, ok := innerMap["NetDevName"].(string); ok && n != "" {
					names[sanitizeKey(n)] = struct{}{}
				}
				if n, ok := innerMap["NetDevIfName"].(string); ok && n != "" {
					names[sanitizeKey(n)] = struct{}{}
				}
			}

			// find a matching netdev entry
			for _, v := range nd {
				if im, ok := v.(map[string]any); ok {
					var candidate string
					if n, ok := im["NetDevIfName"].(string); ok && n != "" {
						candidate = sanitizeKey(n)
					} else if n, ok := im["NetDevName"].(string); ok && n != "" {
						candidate = sanitizeKey(n)
					} else if n, ok := im["Name"].(string); ok && n != "" {
						candidate = sanitizeKey(n)
					}
					if candidate == "" {
						continue
					}
					if _, found := names[candidate]; found {
						// merge netdev fields, prefer netdev values by overriding
						for k, vv := range im {
							norm[strings.ToLower(k)] = vv
						}
						// also ensure innerMap is set so later code can inspect it
						if innerMap == nil {
							innerMap = im
						}
						break
					}
				}
			}
		}

		// Helper extractors
		getBool := func(k string) (bool, bool) {
			if v, ok := norm[strings.ToLower(k)]; ok {
				if b, ok2 := v.(bool); ok2 {
					return b, true
				}
			}
			return false, false
		}
		getString := func(k string) (string, bool) {
			if v, ok := norm[strings.ToLower(k)]; ok {
				if s, ok2 := v.(string); ok2 {
					return s, true
				}
			}
			return "", false
		}
		getFloat := func(k string) (float64, bool) {
			if v, ok := norm[strings.ToLower(k)]; ok {
				switch vv := v.(type) {
				case float64:
					return vv, true
				case int:
					return float64(vv), true
				}
			}
			return 0, false
		}

		// Extract metrics
		state := 0.0
		if b, ok := getBool("Status"); ok && b {
			state = 1.0
		}
		if sStr, ok := getString("NetDevState"); ok && strings.ToLower(sStr) == "up" {
			state = 1.0
		}
		flags := ""
		if sStr, ok := getString("NetDevFlags"); ok {
			flags = sStr
		} else if sStr, ok := getString("Flags"); ok {
			flags = sStr
		}
		mtu := 0.0
		if v, ok := getFloat("MTU"); ok {
			mtu = v
		}
		tx := 0.0
		if v, ok := getFloat("TxQueueLen"); ok {
			tx = v
		}
		speed := 0.0
		if v, ok := getFloat("CurrentBitRate"); ok {
			speed = v
		} else if v, ok := getFloat("CurrentBitRateMbps"); ok {
			speed = v
		}
		lct := 0.0
		if v, ok := getFloat("LastChangeTime"); ok {
			lct = v
		}
		lladdr := ""
		if sStr, ok := getString("LLAddress"); ok {
			lladdr = sStr
		}
		dtype := ""
		if sStr, ok := getString("NetDevType"); ok {
			dtype = sStr
		}

		// alias: try to read top-level alias if present
		alias := ""
		if am, ok := s["alias"].(map[string]any); ok {
			if aStr, ok := am["Alias"].(string); ok {
				alias = aStr
			}
		}

		// Determine canonical interface name to use as label
		canon := sanitizeKey(cand)
		if innerMap != nil {
			if n, ok := innerMap["NetDevIfName"].(string); ok && n != "" {
				canon = sanitizeKey(n)
			} else if n, ok := innerMap["Name"].(string); ok && n != "" {
				canon = sanitizeKey(n)
			} else if n, ok := innerMap["NetDevName"].(string); ok && n != "" {
				canon = sanitizeKey(n)
			}
		}

		debugLog("EMIT netdev ifname=%s mtu=%v tx=%v up=%v", canon, mtu, tx, state)
		ch <- prometheus.MustNewConstMetric(netdevUp, prometheus.GaugeValue, state, canon)
		ch <- prometheus.MustNewConstMetric(netdevMtu, prometheus.GaugeValue, mtu, canon)
		ch <- prometheus.MustNewConstMetric(netdevTxQueueLen, prometheus.GaugeValue, tx, canon)
		ch <- prometheus.MustNewConstMetric(netdevSpeedMbps, prometheus.GaugeValue, speed, canon)
		ch <- prometheus.MustNewConstMetric(netdevLastChange, prometheus.GaugeValue, lct, canon)
		ch <- prometheus.MustNewConstMetric(netdevInfo, prometheus.GaugeValue, 1.0, canon, alias, flags, lladdr, dtype)
	}

}

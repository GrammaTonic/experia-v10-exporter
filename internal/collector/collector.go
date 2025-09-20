// Fetch eth0 interface stats and log the response
package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"

	connectivity "github.com/GrammaTonic/experia-v10-exporter/internal/collector/connectivity"

	"strconv"
	"strings"
	"time"

	metrics "github.com/GrammaTonic/experia-v10-exporter/internal/collector/metrics"
	parsernemo "github.com/GrammaTonic/experia-v10-exporter/internal/collector/parser/nemo"
	nemo "github.com/GrammaTonic/experia-v10-exporter/internal/collector/services/nemo"
	nmc "github.com/GrammaTonic/experia-v10-exporter/internal/collector/services/nmc"
	"github.com/prometheus/client_golang/prometheus"
)

// metricPrefix moved to internal/collector/metrics as MetricPrefix

// defaultNetdevCandidates lists the uppercase interface identifiers used when
// constructing NeMo service calls (NeMo.Intf.<IF>). These are requested as
// uppercase (for example "ETH0") but exported Prometheus labels are stable
// lowercase names derived from the candidate index (eth1, eth2...).
//
// The default set may be overridden at runtime using the
// EXPERIA_EXPECT_NETDEV_IFACES environment variable. That variable should be a
// comma-separated list of interface identifiers (for example: "eth0,eth1");
// entries will be upper-cased before use in service payloads.
var defaultNetdevCandidates = []string{"ETH0", "ETH1", "ETH2", "ETH3"}

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
	// netdevCandidates, when non-empty, overrides the package default list of
	// interface candidates used to construct NeMo service calls. Values should
	// be provided as uppercase identifiers (e.g. "ETH0"). The
	// EXPERIA_EXPECT_NETDEV_IFACES environment variable still takes
	// precedence at runtime for overriding candidates.
	netdevCandidates []string
	// session holds the active authentication context (token). It's set by
	// Login() at startup and refreshed on-demand. Protect with a RWMutex.
	session   sessionContext
	sessionMu sync.RWMutex
}

func NewCollector(ip net.IP, username, password string, timeout time.Duration, candidates ...string) *Experiav10Collector {
	c := &Experiav10Collector{
		ip:       ip,
		username: username,
		password: password,
		client:   connectivity.NewHTTPClient(timeout),
		upMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: metrics.MetricPrefix + "up",
			Help: "Shows if the Experia Box V10 is deemed up by the collector.",
		}),
		authErrorsMetric: prometheus.NewCounter(prometheus.CounterOpts{
			Name: metrics.MetricPrefix + "auth_errors_total",
			Help: "Counts number of authentication errors encountered by the collector.",
		}),
		scrapeErrorsMetric: prometheus.NewCounter(prometheus.CounterOpts{
			Name: metrics.MetricPrefix + "scrape_errors_total",
			Help: "Counts the number of scrape errors by this collector.",
		}),
	}

	// If explicit candidates passed, normalize and store them on the collector.
	if len(candidates) > 0 {
		list := make([]string, 0, len(candidates))
		for _, p := range candidates {
			if p == "" {
				continue
			}
			list = append(list, strings.ToUpper(strings.TrimSpace(p)))
		}
		c.netdevCandidates = list
	}

	return c
}

// Login performs authentication and stores the session token on the collector.
// This is intended to be called once at startup so subsequent scrapes use the
// established session (cookies + token headers). It returns an error if
// authentication fails.
func (c *Experiav10Collector) Login() error {
	apiURL := fmt.Sprintf(apiUrl, c.ip.String())
	token, err := connectivity.Authenticate(c.client, apiURL, c.username, c.password, newRequest, jsonMarshal)
	if err != nil {
		return err
	}
	c.sessionMu.Lock()
	c.session = sessionContext{Token: token}
	c.sessionMu.Unlock()
	return nil
}

func (c *Experiav10Collector) Collect(ch chan<- prometheus.Metric) {
	// Use the pre-established session if available. If there is no session
	// (empty token) attempt to authenticate on-demand; this provides a
	// fallback for tests or runs where Login() was not invoked.
	c.sessionMu.RLock()
	sess := c.session
	c.sessionMu.RUnlock()
	if sess.Token == "" {
		// Try to establish a session for this scrape
		apiURL := fmt.Sprintf(apiUrl, c.ip.String())
		token, err := connectivity.Authenticate(c.client, apiURL, c.username, c.password, newRequest, jsonMarshal)
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
				metrics.IfupTime,
				prometheus.GaugeValue,
				0.0,
				"", "", "Unknown", "", "",
			)
			return
		}
		c.sessionMu.Lock()
		c.session = sessionContext{Token: token}
		c.sessionMu.Unlock()
		sess = sessionContext{Token: token}
	}

	// whether we're running E2E/debug mode
	e2e := os.Getenv("EXPERIA_E2E") == "1"
	debugLog := func(format string, args ...interface{}) {
		if e2e {
			log.Printf(format, args...)
		}
	}

	// Helper for POST fetches, returns response as string. Always re-read the
	// collector's stored session under the mutex so that Login() performed at
	// startup is respected by subsequent scrapes.
	postFetch := func(body string) string {
		url := fmt.Sprintf(apiUrl, c.ip.String())

		// Read the (possibly updated) session token under a read lock.
		c.sessionMu.RLock()
		token := c.session.Token
		c.sessionMu.RUnlock()

		headers := map[string]string{
			"accept":          "*/*",
			"accept-language": "en-US,en;q=0.7",
			"content-type":    "application/x-sah-ws-4-call+json",
			"sec-gpc":         "1",
			"Authorization":   "X-Sah " + token,
			"x-context":       token,
			// Add browser-matching headers for POSTs
			"Origin":  "http://192.168.2.254",
			"Referer": "http://192.168.2.254/",
		}

		// Per-request context with the client's configured timeout so a
		// single slow request does not block indefinitely.
		reqCtx, cancel := context.WithTimeout(context.Background(), c.client.Timeout)
		defer cancel()

		resp, err := connectivity.FetchURL(c.client, reqCtx, "POST", url, headers, []byte(body))
		if err != nil {
			log.Printf("ERROR: failed to fetch %s: %v", body, err)
			c.scrapeErrorsMetric.Inc()
			return ""
		}
		return string(resp)
	}

	// Fetch getWANStatus and export metrics
	wanStatusResp := postFetch(nmc.RequestBody())
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
						metrics.PermissionErrors.Inc()
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
				metrics.IfupTime,
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
			metrics.IfupTime,
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
	// EXPERIA_EXPECT_NETDEV_IFACES (comma-separated). By default we use the
	// collector's configured netdevCandidates (if provided) otherwise the
	// package-level defaultNetdevCandidates.
	var candidates []string
	if len(c.netdevCandidates) > 0 {
		candidates = make([]string, len(c.netdevCandidates))
		copy(candidates, c.netdevCandidates)
	} else {
		candidates = make([]string, len(defaultNetdevCandidates))
		copy(candidates, defaultNetdevCandidates)
	}
	if env := os.Getenv("EXPERIA_EXPECT_NETDEV_IFACES"); env != "" {
		candidates = nil
		for _, p := range strings.Split(env, ",") {
			pp := strings.TrimSpace(p)
			if pp != "" {
				candidates = append(candidates, strings.ToUpper(pp))
			}
		}
	}

	// ...existing code...

	// Process each candidate immediately and emit metrics per-interface.
	// We no longer try to match names returned by the device. Instead we
	// perform the requests for "ETH0","ETH1","ETH2","ETH3" and
	// unconditionally expose metrics with labels "eth1","eth2","eth3",
	// "eth4" based on the candidate index (1-based). This removes any
	// name-mapping logic and keeps label naming stable across firmware.
	for idx, cand := range candidates {
		// canonical label: eth1..ethN (1-based)
		labelName := fmt.Sprintf("eth%d", idx+1)
		body := nemo.RequestBody(cand)
		resp := postFetch(body)
		debugLog("DEBUG: getMIBs service=%s response length=%d", cand, len(resp))
		// debugLog is gated by EXPERIA_E2E; print raw JSON when enabled.
		debugLog("DEBUG RAW getMIBs service=%s: %s", cand, resp)
		if resp == "" {
			// still emit a zeroed metric so that collectors see the family when
			// the device doesn't return useful data for a candidate. Use the
			// stable labelName (eth1..ethN) instead of device-provided names.
			ch <- prometheus.MustNewConstMetric(metrics.NetdevUp, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevMtu, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevTxQueueLen, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevSpeedMbps, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevLastChange, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevInfo, prometheus.GaugeValue, 1.0, labelName, "", "", "", "")
			continue
		}
		// Parse and normalize MIB response using shared parser logic
		norm, s, err := parsernemo.ParseMIBs([]byte(resp), cand)
		if err != nil {
			continue
		}
		if norm == nil {
			// no usable data for this candidate, emit zeroed metrics
			ch <- prometheus.MustNewConstMetric(metrics.NetdevUp, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevMtu, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevTxQueueLen, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevSpeedMbps, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevLastChange, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevInfo, prometheus.GaugeValue, 1.0, labelName, "", "", "", "")
			continue
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

		// alias: try to read top-level alias if present (use status map returned by parser)
		alias := ""
		if s != nil {
			if am, ok := s["alias"].(map[string]any); ok {
				if aStr, ok := am["Alias"].(string); ok {
					alias = aStr
				}
			}
		}

		// Emit metrics using the stable labelName (eth1..ethN).
		debugLog("EMIT netdev ifname=%s mtu=%v tx=%v up=%v", labelName, mtu, tx, state)
		ch <- prometheus.MustNewConstMetric(metrics.NetdevUp, prometheus.GaugeValue, state, labelName)
		ch <- prometheus.MustNewConstMetric(metrics.NetdevMtu, prometheus.GaugeValue, mtu, labelName)
		ch <- prometheus.MustNewConstMetric(metrics.NetdevTxQueueLen, prometheus.GaugeValue, tx, labelName)
		ch <- prometheus.MustNewConstMetric(metrics.NetdevSpeedMbps, prometheus.GaugeValue, speed, labelName)
		ch <- prometheus.MustNewConstMetric(metrics.NetdevLastChange, prometheus.GaugeValue, lct, labelName)
		ch <- prometheus.MustNewConstMetric(metrics.NetdevInfo, prometheus.GaugeValue, 1.0, labelName, alias, flags, lladdr, dtype)

		// Also fetch per-interface statistics via getNetDevStats and export them.
		// This mirrors the device API call:
		// {"service":"NeMo.Intf.ETH0","method":"getNetDevStats","parameters":{}}
		statsBody := nemo.RequestBodyStats(cand)
		statsResp := postFetch(statsBody)
		debugLog("DEBUG: getNetDevStats service=%s response length=%d", cand, len(statsResp))
		debugLog("DEBUG RAW getNetDevStats service=%s: %s", cand, statsResp)
		if statsResp == "" {
			// Emit zeroed stats so metric families are present even when the
			// device doesn't return data for this candidate.
			ch <- prometheus.MustNewConstMetric(metrics.NetdevRxPackets, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevTxPackets, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevRxBytes, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevTxBytes, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevRxErrors, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevTxErrors, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevRxDropped, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevTxDropped, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevMulticast, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevCollisions, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevRxLengthErrors, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevRxOverErrors, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevRxCrcErrors, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevRxFrameErrors, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevRxFifoErrors, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevRxMissedErrors, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevTxAbortedErrors, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevTxCarrierErrors, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevTxFifoErrors, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevTxHeartbeatErrors, prometheus.GaugeValue, 0.0, labelName)
			ch <- prometheus.MustNewConstMetric(metrics.NetdevTxWindowErrors, prometheus.GaugeValue, 0.0, labelName)
			continue
		}
		if data, err := parsernemo.ParseNetDevStats([]byte(statsResp)); err == nil {
			getNum := func(k string) (float64, bool) {
				if data == nil {
					return 0, false
				}
				if v, ok := data[k]; ok {
					switch vv := v.(type) {
					case float64:
						return vv, true
					case int:
						return float64(vv), true
					case string:
						if vv == "" {
							return 0, false
						}
						if f, err := strconv.ParseFloat(vv, 64); err == nil {
							return f, true
						}
					}
				}
				// try lower-cased keys as some firmwares use different casing
				if v, ok := data[strings.ToLower(k)]; ok {
					switch vv := v.(type) {
					case float64:
						return vv, true
					case int:
						return float64(vv), true
					case string:
						if vv == "" {
							return 0, false
						}
						if f, err := strconv.ParseFloat(vv, 64); err == nil {
							return f, true
						}
					}
				}
				return 0, false
			}

			// Emit metrics if present, otherwise zero values to keep families present.
			if v, ok := getNum("RxPackets"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxPackets, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxPackets, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("TxPackets"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxPackets, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxPackets, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("RxBytes"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxBytes, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxBytes, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("TxBytes"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxBytes, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxBytes, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("RxErrors"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxErrors, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxErrors, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("TxErrors"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxErrors, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxErrors, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("RxDropped"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxDropped, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxDropped, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("TxDropped"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxDropped, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxDropped, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("Multicast"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevMulticast, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevMulticast, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("Collisions"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevCollisions, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevCollisions, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("RxLengthErrors"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxLengthErrors, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxLengthErrors, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("RxOverErrors"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxOverErrors, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxOverErrors, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("RxCrcErrors"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxCrcErrors, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxCrcErrors, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("RxFrameErrors"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxFrameErrors, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxFrameErrors, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("RxFifoErrors"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxFifoErrors, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxFifoErrors, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("RxMissedErrors"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxMissedErrors, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevRxMissedErrors, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("TxAbortedErrors"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxAbortedErrors, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxAbortedErrors, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("TxCarrierErrors"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxCarrierErrors, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxCarrierErrors, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("TxFifoErrors"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxFifoErrors, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxFifoErrors, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("TxHeartbeatErrors"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxHeartbeatErrors, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxHeartbeatErrors, prometheus.GaugeValue, 0.0, labelName)
			}
			if v, ok := getNum("TxWindowErrors"); ok {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxWindowErrors, prometheus.GaugeValue, v, labelName)
			} else {
				ch <- prometheus.MustNewConstMetric(metrics.NetdevTxWindowErrors, prometheus.GaugeValue, 0.0, labelName)
			}
		}
	}

}

// CookiesForHost returns the cookies stored in the client's jar for the
// provided host URL (for example "http://192.168.2.254"). It returns an
// empty slice if there are no cookies or the URL cannot be parsed.
func (c *Experiav10Collector) CookiesForHost(hostURL string) []*http.Cookie {
	if c == nil || c.client == nil || c.client.Jar == nil {
		return nil
	}
	u, err := url.Parse(hostURL)
	if err != nil {
		return nil
	}
	return c.client.Jar.Cookies(u)
}

// SessionToken returns the currently stored session token (contextID). It
// acquires the read lock while reading the value.
func (c *Experiav10Collector) SessionToken() string {
	c.sessionMu.RLock()
	defer c.sessionMu.RUnlock()
	return c.session.Token
}

// authenticate is retained for test compatibility: many tests call the
// collector's authenticate method directly. It wraps the exported
// connectivity.Authenticate and returns the sessionContext on success.
func (c *Experiav10Collector) authenticate() (sessionContext, error) {
	apiURL := fmt.Sprintf(apiUrl, c.ip.String())
	token, err := connectivity.Authenticate(c.client, apiURL, c.username, c.password, newRequest, jsonMarshal)
	if err != nil {
		return sessionContext{}, err
	}
	// store token on the collector
	c.sessionMu.Lock()
	c.session = sessionContext{Token: token}
	c.sessionMu.Unlock()
	return sessionContext{Token: token}, nil
}

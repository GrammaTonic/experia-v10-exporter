package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricPrefix = "experia_v10_"
)

const (
	tokenUrl  = "http://%s/function_module/login_module/login_page/logintoken_lua.lua"
	loginUrl  = "http://%s"
	logoutUrl = "http://%s"

	lanPageUrl    = "http://%s/getpage.lua?pid=123&nextpage=Localnet_LAN_LocalnetStatus_t.lp&Menu3Location=0&_=1611056303063"
	lanMetricsUrl = "http://%s/common_page/lanStatus_lua.lua"

	ethernetPageUrl    = "http://%s/getpage.lua?pid=123&nextpage=Internet_InternetStatusforRoute_t.lp&Menu3Location=0&_=1583884785730"
	ethernetMetricsUrl = "http://%s/common_page/internet_eth_interface_lua.lua"
)

var (
	ethernetDesc = prometheus.NewDesc(
		metricPrefix+"ethernet_interface",
		"All wan related metadata.",
		[]string{"value"}, nil)

	ifInOctets = prometheus.NewDesc(
		metricPrefix+"interface_received_bytes_total",
		"The total number of bytes received on the interface",
		[]string{"id", "alias", "mac", "status"}, nil)
	ifOutOctets = prometheus.NewDesc(
		metricPrefix+"interface_sent_bytes_total",
		"The total number of bytes transmitted out of the interface",
		[]string{"id", "alias", "mac", "status"}, nil)
)

type experiav10Collector struct {
	ip                 net.IP
	username           string
	password           string
	client             *http.Client
	upMetric           prometheus.Gauge
	authErrorsMetric   prometheus.Counter
	scrapeErrorsMetric prometheus.Counter
}

func newCollector(ip net.IP, username, password string, timeout time.Duration) *experiav10Collector {
	cookieJar, _ := cookiejar.New(nil)

	return &experiav10Collector{
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

func (c *experiav10Collector) Describe(ch chan<- *prometheus.Desc) {
	c.upMetric.Describe(ch)
	c.authErrorsMetric.Describe(ch)
	c.scrapeErrorsMetric.Describe(ch)
	ch <- ifInOctets
	ch <- ifOutOctets
}

func (c *experiav10Collector) Collect(ch chan<- prometheus.Metric) {
	if err := c.login(ch); err != nil {
		log.Printf("Error during authentication: %s", err)

		c.authErrorsMetric.Inc()
		c.upMetric.Set(0)
	} else {
		defer c.logout(ch)

		if err := c.scrape(ch); err != nil {
			log.Printf("Error during scrape: %s", err)

			c.scrapeErrorsMetric.Inc()
			c.upMetric.Set(0)
		} else {
			c.upMetric.Set(1)
		}
	}

	c.upMetric.Collect(ch)
	c.authErrorsMetric.Collect(ch)
	c.scrapeErrorsMetric.Collect(ch)
}

func (c *experiav10Collector) login(ch chan<- prometheus.Metric) error {
	loginPageRequest, err := c.client.Get(fmt.Sprintf(loginUrl, c.ip.String()))
	if err != nil {
		return err
	}
	loginPageRequest.Body.Close()

	tokenRequest, err := c.client.Get(fmt.Sprintf(tokenUrl, c.ip.String()))
	if err != nil {
		return err
	}
	defer tokenRequest.Body.Close()

	var tokenResponse struct {
		Token int `xml:",chardata"`
	}
	if err := xml.NewDecoder(tokenRequest.Body).Decode(&tokenResponse); err != nil {
		return fmt.Errorf("failed to parse login token: %w", err)
	}

	loginParams := url.Values{}
	passwordHash := sha256.Sum256([]byte(c.password + strconv.Itoa(tokenResponse.Token)))

	loginParams.Set("Username", c.username)
	loginParams.Set("Password", hex.EncodeToString(passwordHash[:]))
	loginParams.Set("action", "login")

	loginRequest, err := c.client.PostForm(fmt.Sprintf(loginUrl, c.ip.String()), loginParams)
	if err != nil {
		return err
	}
	defer loginRequest.Body.Close()

	body, err := io.ReadAll(loginRequest.Body)
	if err != nil {
		return err
	}

	if strings.Contains(string(body), "loginWrapper") {
		return errors.New("unable to login")
	}

	return nil
}

func (c *experiav10Collector) logout(ch chan<- prometheus.Metric) error {
	logoutParams := url.Values{}
	logoutParams.Set("IF_LogOff", "1")
	logoutParams.Set("IF_LanguageSwitch", "")
	logoutParams.Set("IF_ModeSwitch", "")

	logoutRequest, err := c.client.PostForm(fmt.Sprintf(logoutUrl, c.ip.String()), logoutParams)
	if err != nil {
		return err
	}
	logoutRequest.Body.Close()

	c.client.Jar, _ = cookiejar.New(nil)

	return nil
}

func (c *experiav10Collector) scrape(ch chan<- prometheus.Metric) error {
	ethernetPageRequest, err := c.client.Get(fmt.Sprintf(ethernetPageUrl, c.ip.String()))
	if err != nil {
		return err
	}
	ethernetPageRequest.Body.Close()

	ethernetMetricsRequest, err := c.client.Get(fmt.Sprintf(ethernetMetricsUrl, c.ip.String()))
	if err != nil {
		return err
	}
	defer ethernetMetricsRequest.Body.Close()

	var ethernetMetrics struct {
		Names  []string `xml:"OBJ_ETH_ID>Instance>ParaName"`
		Values []string `xml:"OBJ_ETH_ID>Instance>ParaValue"`
	}

	if err := xml.NewDecoder(ethernetMetricsRequest.Body).Decode(&ethernetMetrics); err != nil {
		return err
	}

	fmt.Print(ethernetMetrics)

	for i := 0; i < len(ethernetMetrics.Names); i += 9 {
		ifName := ethernetMetrics.Values[i]
		// ifAlias := ethernetMetrics.Values[i+1]
		ifAlias := "WAN"
		ifMAC := ethernetMetrics.Values[i+5]
		ifStatus := ethernetMetrics.Values[i+3]

		inBytes, err := strconv.ParseFloat(ethernetMetrics.Values[i+2], 64)
		if err != nil {
			continue
		}
		outBytes, err := strconv.ParseFloat(ethernetMetrics.Values[i+7], 64)
		if err != nil {
			continue
		}

		if inBytes > 0 {
			ch <- prometheus.MustNewConstMetric(ifInOctets, prometheus.CounterValue, inBytes, ifName, ifAlias, ifMAC, ifStatus)
		}
		if outBytes > 0 {
			ch <- prometheus.MustNewConstMetric(ifOutOctets, prometheus.CounterValue, outBytes, ifName, ifAlias, ifMAC, ifStatus)
		}
	}
	// Lan Metrics

	lanPageRequest, err := c.client.Get(fmt.Sprintf(lanPageUrl, c.ip.String()))
	if err != nil {
		return err
	}
	lanPageRequest.Body.Close()

	lanMetricsRequest, err := c.client.Get(fmt.Sprintf(lanMetricsUrl, c.ip.String()))
	if err != nil {
		return err
	}
	defer lanMetricsRequest.Body.Close()

	var lanMetrics struct {
		Names  []string `xml:"OBJ_ETH_ID>Instance>ParaName"`
		Values []string `xml:"OBJ_ETH_ID>Instance>ParaValue"`
	}

	if err := xml.NewDecoder(lanMetricsRequest.Body).Decode(&lanMetrics); err != nil {
		return err
	}

	fmt.Print(lanMetrics)

	for i := 0; i < len(lanMetrics.Names); i += 6 {
		ifName := lanMetrics.Values[i]
		ifAlias := lanMetrics.Values[i+1]
		ifMAC := lanMetrics.Values[i+4]
		ifStatus := lanMetrics.Values[i+3]

		inBytes, err := strconv.ParseFloat(lanMetrics.Values[i+2], 64)
		if err != nil {
			continue
		}
		outBytes, err := strconv.ParseFloat(lanMetrics.Values[i+5], 64)
		if err != nil {
			continue
		}

		if inBytes > 0 {
			ch <- prometheus.MustNewConstMetric(ifInOctets, prometheus.CounterValue, inBytes, ifName, ifAlias, ifMAC, ifStatus)
		}
		if outBytes > 0 {
			ch <- prometheus.MustNewConstMetric(ifOutOctets, prometheus.CounterValue, outBytes, ifName, ifAlias, ifMAC, ifStatus)
		}
	}

	return nil
}

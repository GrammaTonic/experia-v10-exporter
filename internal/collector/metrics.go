package collector

import (
"github.com/prometheus/client_golang/prometheus"
)

var (
ifupTime = prometheus.NewDesc(
metricPrefix+"internet_connection",
"The internet connection status",
[]string{"link_type", "protocol", "connection_state", "ip", "mac"}, nil)
permissionErrors = prometheus.NewCounter(prometheus.CounterOpts{
Name: metricPrefix + "permission_errors_total",
Help: "Counts the number of permission denied errors from the modem API.",
})
)

func (c *Experiav10Collector) Describe(ch chan<- *prometheus.Desc) {
	c.upMetric.Describe(ch)
	c.authErrorsMetric.Describe(ch)
	c.scrapeErrorsMetric.Describe(ch)
	ch <- ifupTime
	ch <- permissionErrors.Desc()
}

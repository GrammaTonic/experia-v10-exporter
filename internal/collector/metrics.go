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
	// Netdev / ETH metrics
	netdevUp = prometheus.NewDesc(
		metricPrefix+"netdev_up",
		"1 if the network device is up",
		[]string{"ifname"}, nil)
	netdevMtu = prometheus.NewDesc(
		metricPrefix+"netdev_mtu",
		"MTU of the network device",
		[]string{"ifname"}, nil)
	netdevTxQueueLen = prometheus.NewDesc(
		metricPrefix+"netdev_tx_queue_len",
		"Tx queue length of the network device",
		[]string{"ifname"}, nil)
	netdevSpeedMbps = prometheus.NewDesc(
		metricPrefix+"netdev_speed_mbps",
		"Current bit rate of the device in Mbps",
		[]string{"ifname"}, nil)
	netdevLastChange = prometheus.NewDesc(
		metricPrefix+"netdev_last_change_seconds",
		"LastChange time reported by the device (seconds)",
		[]string{"ifname"}, nil)
	netdevInfo = prometheus.NewDesc(
		metricPrefix+"netdev_info",
		"Static info about the netdev (value is always 1), labels: alias, flags, lladdr, type",
		[]string{"ifname", "alias", "flags", "lladdr", "type"}, nil)

	// Per-interface network statistics (from getNetDevStats)
	netdevRxPackets = prometheus.NewDesc(
		metricPrefix+"netdev_rx_packets_total",
		"Number of received packets",
		[]string{"ifname"}, nil)
	netdevTxPackets = prometheus.NewDesc(
		metricPrefix+"netdev_tx_packets_total",
		"Number of transmitted packets",
		[]string{"ifname"}, nil)
	netdevRxBytes = prometheus.NewDesc(
		metricPrefix+"netdev_rx_bytes_total",
		"Number of received bytes",
		[]string{"ifname"}, nil)
	netdevTxBytes = prometheus.NewDesc(
		metricPrefix+"netdev_tx_bytes_total",
		"Number of transmitted bytes",
		[]string{"ifname"}, nil)
	netdevRxErrors = prometheus.NewDesc(
		metricPrefix+"netdev_rx_errors_total",
		"Number of receive errors",
		[]string{"ifname"}, nil)
	netdevTxErrors = prometheus.NewDesc(
		metricPrefix+"netdev_tx_errors_total",
		"Number of transmit errors",
		[]string{"ifname"}, nil)
	netdevRxDropped = prometheus.NewDesc(
		metricPrefix+"netdev_rx_dropped_total",
		"Number of received dropped packets",
		[]string{"ifname"}, nil)
	netdevTxDropped = prometheus.NewDesc(
		metricPrefix+"netdev_tx_dropped_total",
		"Number of transmitted dropped packets",
		[]string{"ifname"}, nil)
	netdevMulticast = prometheus.NewDesc(
		metricPrefix+"netdev_multicast_total",
		"Number of multicast packets",
		[]string{"ifname"}, nil)
	netdevCollisions = prometheus.NewDesc(
		metricPrefix+"netdev_collisions_total",
		"Number of collisions",
		[]string{"ifname"}, nil)
	// Misc error counters present in some firmware responses
	netdevRxLengthErrors    = prometheus.NewDesc(metricPrefix+"netdev_rx_length_errors_total", "Rx length errors", []string{"ifname"}, nil)
	netdevRxOverErrors      = prometheus.NewDesc(metricPrefix+"netdev_rx_over_errors_total", "Rx over errors", []string{"ifname"}, nil)
	netdevRxCrcErrors       = prometheus.NewDesc(metricPrefix+"netdev_rx_crc_errors_total", "Rx CRC errors", []string{"ifname"}, nil)
	netdevRxFrameErrors     = prometheus.NewDesc(metricPrefix+"netdev_rx_frame_errors_total", "Rx frame errors", []string{"ifname"}, nil)
	netdevRxFifoErrors      = prometheus.NewDesc(metricPrefix+"netdev_rx_fifo_errors_total", "Rx FIFO errors", []string{"ifname"}, nil)
	netdevRxMissedErrors    = prometheus.NewDesc(metricPrefix+"netdev_rx_missed_errors_total", "Rx missed errors", []string{"ifname"}, nil)
	netdevTxAbortedErrors   = prometheus.NewDesc(metricPrefix+"netdev_tx_aborted_errors_total", "Tx aborted errors", []string{"ifname"}, nil)
	netdevTxCarrierErrors   = prometheus.NewDesc(metricPrefix+"netdev_tx_carrier_errors_total", "Tx carrier errors", []string{"ifname"}, nil)
	netdevTxFifoErrors      = prometheus.NewDesc(metricPrefix+"netdev_tx_fifo_errors_total", "Tx FIFO errors", []string{"ifname"}, nil)
	netdevTxHeartbeatErrors = prometheus.NewDesc(metricPrefix+"netdev_tx_heartbeat_errors_total", "Tx heartbeat errors", []string{"ifname"}, nil)
	netdevTxWindowErrors    = prometheus.NewDesc(metricPrefix+"netdev_tx_window_errors_total", "Tx window errors", []string{"ifname"}, nil)
)

func (c *Experiav10Collector) Describe(ch chan<- *prometheus.Desc) {
	c.upMetric.Describe(ch)
	c.authErrorsMetric.Describe(ch)
	c.scrapeErrorsMetric.Describe(ch)
	ch <- ifupTime
	permissionErrors.Describe(ch)
	// describe netdev metrics
	ch <- netdevUp
	ch <- netdevMtu
	ch <- netdevTxQueueLen
	ch <- netdevSpeedMbps
	ch <- netdevLastChange
	ch <- netdevInfo
	// describe netdev stats
	ch <- netdevRxPackets
	ch <- netdevTxPackets
	ch <- netdevRxBytes
	ch <- netdevTxBytes
	ch <- netdevRxErrors
	ch <- netdevTxErrors
	ch <- netdevRxDropped
	ch <- netdevTxDropped
	ch <- netdevMulticast
	ch <- netdevCollisions
	ch <- netdevRxLengthErrors
	ch <- netdevRxOverErrors
	ch <- netdevRxCrcErrors
	ch <- netdevRxFrameErrors
	ch <- netdevRxFifoErrors
	ch <- netdevRxMissedErrors
	ch <- netdevTxAbortedErrors
	ch <- netdevTxCarrierErrors
	ch <- netdevTxFifoErrors
	ch <- netdevTxHeartbeatErrors
	ch <- netdevTxWindowErrors
	ch <- permissionErrors.Desc()
}

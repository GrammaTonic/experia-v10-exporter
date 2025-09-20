package metrics

import "github.com/prometheus/client_golang/prometheus"

const MetricPrefix = "experia_v10_"

var (
	IfupTime = prometheus.NewDesc(
		MetricPrefix+"internet_connection",
		"The internet connection status",
		[]string{"link_type", "protocol", "connection_state", "ip", "mac"}, nil)

	PermissionErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: MetricPrefix + "permission_errors_total",
		Help: "Counts the number of permission denied errors from the modem API.",
	})

	// Netdev / ETH metrics
	NetdevUp = prometheus.NewDesc(
		MetricPrefix+"netdev_up",
		"1 if the network device is up",
		[]string{"ifname"}, nil)
	NetdevMtu = prometheus.NewDesc(
		MetricPrefix+"netdev_mtu",
		"MTU of the network device",
		[]string{"ifname"}, nil)
	NetdevTxQueueLen = prometheus.NewDesc(
		MetricPrefix+"netdev_tx_queue_len",
		"Tx queue length of the network device",
		[]string{"ifname"}, nil)
	NetdevSpeedMbps = prometheus.NewDesc(
		MetricPrefix+"netdev_speed_mbps",
		"Current bit rate of the device in Mbps",
		[]string{"ifname"}, nil)
	NetdevLastChange = prometheus.NewDesc(
		MetricPrefix+"netdev_last_change_seconds",
		"LastChange time reported by the device (seconds)",
		[]string{"ifname"}, nil)
	NetdevInfo = prometheus.NewDesc(
		MetricPrefix+"netdev_info",
		"Static info about the netdev (value is always 1), labels: alias, flags, lladdr, type",
		[]string{"ifname", "alias", "flags", "lladdr", "type"}, nil)

	// Per-interface network statistics (from getNetDevStats)
	NetdevRxPackets = prometheus.NewDesc(
		MetricPrefix+"netdev_rx_packets_total",
		"Number of received packets",
		[]string{"ifname"}, nil)
	NetdevTxPackets = prometheus.NewDesc(
		MetricPrefix+"netdev_tx_packets_total",
		"Number of transmitted packets",
		[]string{"ifname"}, nil)
	NetdevRxBytes = prometheus.NewDesc(
		MetricPrefix+"netdev_rx_bytes_total",
		"Number of received bytes",
		[]string{"ifname"}, nil)
	NetdevTxBytes = prometheus.NewDesc(
		MetricPrefix+"netdev_tx_bytes_total",
		"Number of transmitted bytes",
		[]string{"ifname"}, nil)
	NetdevRxErrors = prometheus.NewDesc(
		MetricPrefix+"netdev_rx_errors_total",
		"Number of receive errors",
		[]string{"ifname"}, nil)
	NetdevTxErrors = prometheus.NewDesc(
		MetricPrefix+"netdev_tx_errors_total",
		"Number of transmit errors",
		[]string{"ifname"}, nil)
	NetdevRxDropped = prometheus.NewDesc(
		MetricPrefix+"netdev_rx_dropped_total",
		"Number of received dropped packets",
		[]string{"ifname"}, nil)
	NetdevTxDropped = prometheus.NewDesc(
		MetricPrefix+"netdev_tx_dropped_total",
		"Number of transmitted dropped packets",
		[]string{"ifname"}, nil)
	NetdevMulticast = prometheus.NewDesc(
		MetricPrefix+"netdev_multicast_total",
		"Number of multicast packets",
		[]string{"ifname"}, nil)
	NetdevCollisions = prometheus.NewDesc(
		MetricPrefix+"netdev_collisions_total",
		"Number of collisions",
		[]string{"ifname"}, nil)
	// Misc error counters present in some firmware responses
	NetdevRxLengthErrors    = prometheus.NewDesc(MetricPrefix+"netdev_rx_length_errors_total", "Rx length errors", []string{"ifname"}, nil)
	NetdevRxOverErrors      = prometheus.NewDesc(MetricPrefix+"netdev_rx_over_errors_total", "Rx over errors", []string{"ifname"}, nil)
	NetdevRxCrcErrors       = prometheus.NewDesc(MetricPrefix+"netdev_rx_crc_errors_total", "Rx CRC errors", []string{"ifname"}, nil)
	NetdevRxFrameErrors     = prometheus.NewDesc(MetricPrefix+"netdev_rx_frame_errors_total", "Rx frame errors", []string{"ifname"}, nil)
	NetdevRxFifoErrors      = prometheus.NewDesc(MetricPrefix+"netdev_rx_fifo_errors_total", "Rx FIFO errors", []string{"ifname"}, nil)
	NetdevRxMissedErrors    = prometheus.NewDesc(MetricPrefix+"netdev_rx_missed_errors_total", "Rx missed errors", []string{"ifname"}, nil)
	NetdevTxAbortedErrors   = prometheus.NewDesc(MetricPrefix+"netdev_tx_aborted_errors_total", "Tx aborted errors", []string{"ifname"}, nil)
	NetdevTxCarrierErrors   = prometheus.NewDesc(MetricPrefix+"netdev_tx_carrier_errors_total", "Tx carrier errors", []string{"ifname"}, nil)
	NetdevTxFifoErrors      = prometheus.NewDesc(MetricPrefix+"netdev_tx_fifo_errors_total", "Tx FIFO errors", []string{"ifname"}, nil)
	NetdevTxHeartbeatErrors = prometheus.NewDesc(MetricPrefix+"netdev_tx_heartbeat_errors_total", "Tx heartbeat errors", []string{"ifname"}, nil)
	NetdevTxWindowErrors    = prometheus.NewDesc(MetricPrefix+"netdev_tx_window_errors_total", "Tx window errors", []string{"ifname"}, nil)
)

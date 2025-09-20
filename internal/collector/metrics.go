package collector

import (
	metrics "github.com/GrammaTonic/experia-v10-exporter/internal/collector/metrics"
	nemo "github.com/GrammaTonic/experia-v10-exporter/internal/collector/metrics/nemo"
	nmc "github.com/GrammaTonic/experia-v10-exporter/internal/collector/metrics/nmc"
	"github.com/prometheus/client_golang/prometheus"
)

func (c *Experiav10Collector) Describe(ch chan<- *prometheus.Desc) {
	c.upMetric.Describe(ch)
	c.authErrorsMetric.Describe(ch)
	c.scrapeErrorsMetric.Describe(ch)
	// describe WAN / general metrics
	ch <- metrics.IfupTime
	nmc.PermissionErrors.Describe(ch)
	// describe netdev metrics
	ch <- nemo.NetdevUp
	ch <- nemo.NetdevMtu
	ch <- nemo.NetdevTxQueueLen
	ch <- nemo.NetdevSpeedMbps
	ch <- nemo.NetdevLastChange
	ch <- nemo.NetdevInfo
	// describe netdev stats
	ch <- nemo.NetdevRxPackets
	ch <- nemo.NetdevTxPackets
	ch <- nemo.NetdevRxBytes
	ch <- nemo.NetdevTxBytes
	ch <- nemo.NetdevRxErrors
	ch <- nemo.NetdevTxErrors
	ch <- nemo.NetdevRxDropped
	ch <- nemo.NetdevTxDropped
	ch <- nemo.NetdevMulticast
	ch <- nemo.NetdevCollisions
	ch <- nemo.NetdevRxLengthErrors
	ch <- nemo.NetdevRxOverErrors
	ch <- nemo.NetdevRxCrcErrors
	ch <- nemo.NetdevRxFrameErrors
	ch <- nemo.NetdevRxFifoErrors
	ch <- nemo.NetdevRxMissedErrors
	ch <- nemo.NetdevTxAbortedErrors
	ch <- nemo.NetdevTxCarrierErrors
	ch <- nemo.NetdevTxFifoErrors
	ch <- nemo.NetdevTxHeartbeatErrors
	ch <- nemo.NetdevTxWindowErrors
	metrics.PermissionErrors.Describe(ch)
}

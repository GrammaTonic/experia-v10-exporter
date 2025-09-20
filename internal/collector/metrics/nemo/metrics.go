package nemo

import base "github.com/GrammaTonic/experia-v10-exporter/internal/collector/metrics"

var (
	NetdevUp         = base.NetdevUp
	NetdevMtu        = base.NetdevMtu
	NetdevTxQueueLen = base.NetdevTxQueueLen
	NetdevSpeedMbps  = base.NetdevSpeedMbps
	NetdevLastChange = base.NetdevLastChange
	NetdevInfo       = base.NetdevInfo

	NetdevPortCurrentBitrate      = base.NetdevPortCurrentBitrate
	NetdevPortMaxBitRateSupported = base.NetdevPortMaxBitRateSupported
	NetdevPortMaxBitRateEnabled   = base.NetdevPortMaxBitRateEnabled
	NetdevPortDuplexEnabled       = base.NetdevPortDuplexEnabled
	NetdevPortSetPortInfo         = base.NetdevPortSetPortInfo

	NetdevRxPackets  = base.NetdevRxPackets
	NetdevTxPackets  = base.NetdevTxPackets
	NetdevRxBytes    = base.NetdevRxBytes
	NetdevTxBytes    = base.NetdevTxBytes
	NetdevRxErrors   = base.NetdevRxErrors
	NetdevTxErrors   = base.NetdevTxErrors
	NetdevRxDropped  = base.NetdevRxDropped
	NetdevTxDropped  = base.NetdevTxDropped
	NetdevMulticast  = base.NetdevMulticast
	NetdevCollisions = base.NetdevCollisions

	NetdevRxLengthErrors    = base.NetdevRxLengthErrors
	NetdevRxOverErrors      = base.NetdevRxOverErrors
	NetdevRxCrcErrors       = base.NetdevRxCrcErrors
	NetdevRxFrameErrors     = base.NetdevRxFrameErrors
	NetdevRxFifoErrors      = base.NetdevRxFifoErrors
	NetdevRxMissedErrors    = base.NetdevRxMissedErrors
	NetdevTxAbortedErrors   = base.NetdevTxAbortedErrors
	NetdevTxCarrierErrors   = base.NetdevTxCarrierErrors
	NetdevTxFifoErrors      = base.NetdevTxFifoErrors
	NetdevTxHeartbeatErrors = base.NetdevTxHeartbeatErrors
	NetdevTxWindowErrors    = base.NetdevTxWindowErrors
)

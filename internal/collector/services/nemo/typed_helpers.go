package nemo

import (
	"strconv"
	"strings"
)

// MIBInfo is a typed representation of key values extracted from a getMIBs
// response for a single interface candidate.
type MIBInfo struct {
	Candidate           string
	Alias               string
	Flags               string
	LLAddress           string
	MTU                 float64
	TxQueueLen          float64
	NetDevState         string
	CurrentBitRate      float64
	LastChangeTime      float64
	MaxBitRateSupported float64
	MaxBitRateEnabled   float64
	CurrentDuplexMode   string
	DuplexModeEnabled   bool
}

// NetDevStats holds numeric per-interface counters returned by getNetDevStats.
type NetDevStats struct {
	RxPackets         float64
	TxPackets         float64
	RxBytes           float64
	TxBytes           float64
	RxErrors          float64
	TxErrors          float64
	RxDropped         float64
	TxDropped         float64
	Multicast         float64
	Collisions        float64
	RxLengthErrors    float64
	RxOverErrors      float64
	RxCrcErrors       float64
	RxFrameErrors     float64
	RxFifoErrors      float64
	RxMissedErrors    float64
	TxAbortedErrors   float64
	TxCarrierErrors   float64
	TxFifoErrors      float64
	TxHeartbeatErrors float64
	TxWindowErrors    float64
}

// PortParams contains port parameters commonly pulled from a getMIBs response
// for WAN/ethernet port pages (MaxBitRate*, Duplex, LLIntf mapping etc.).
type PortParams struct {
	LLIntf              string
	CurrentBitRate      float64
	MaxBitRateSupported float64
	MaxBitRateEnabled   float64
	CurrentDuplexMode   string
	DuplexModeEnabled   bool
	SetPort             string
}

// helper: convert an any -> float64 when possible
func anyToFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case string:
		if t == "" {
			return 0, false
		}
		if f, err := strconv.ParseFloat(t, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// helper: read string from normalized map (keys are lowercased by ParseMIBs)
func readString(norm map[string]any, key string) (string, bool) {
	if norm == nil {
		return "", false
	}
	if v, ok := norm[strings.ToLower(key)]; ok {
		if s, ok2 := v.(string); ok2 {
			return s, true
		}
	}
	if v, ok := norm[key]; ok {
		if s, ok2 := v.(string); ok2 {
			return s, true
		}
	}
	return "", false
}

// helper: read float from normalized map
func readFloat(norm map[string]any, key string) (float64, bool) {
	if norm == nil {
		return 0, false
	}
	if v, ok := norm[strings.ToLower(key)]; ok {
		if f, ok2 := anyToFloat(v); ok2 {
			return f, true
		}
	}
	if v, ok := norm[key]; ok {
		if f, ok2 := anyToFloat(v); ok2 {
			return f, true
		}
	}
	return 0, false
}

// helper: read bool from normalized map
func readBool(norm map[string]any, key string) (bool, bool) {
	if norm == nil {
		return false, false
	}
	if v, ok := norm[strings.ToLower(key)]; ok {
		if b, ok2 := v.(bool); ok2 {
			return b, true
		}
		if s, ok2 := v.(string); ok2 {
			// common string representations
			if s == "true" || s == "1" {
				return true, true
			}
			if s == "false" || s == "0" {
				return false, true
			}
		}
	}
	if v, ok := norm[key]; ok {
		if b, ok2 := v.(bool); ok2 {
			return b, true
		}
	}
	return false, false
}

// GetMIBsTyped returns a typed MIBInfo and the raw status map (if available).
// It uses the existing ParseMIBs parser to normalize the input then maps
// commonly used fields into the typed struct.
func GetMIBsTyped(data []byte, candidate string) (MIBInfo, map[string]any, error) {
	var mi MIBInfo
	norm, s, err := ParseMIBs(data, candidate)
	if err != nil || norm == nil {
		return mi, s, err
	}
	mi.Candidate = candidate
	if v, ok := readString(norm, "alias"); ok {
		mi.Alias = v
	}
	if v, ok := readString(norm, "flags"); ok {
		mi.Flags = v
	}
	if v, ok := readString(norm, "lladdress"); ok {
		mi.LLAddress = v
	}
	if v, ok := readFloat(norm, "mtu"); ok {
		mi.MTU = v
	}
	if v, ok := readFloat(norm, "txqueuelen"); ok {
		mi.TxQueueLen = v
	}
	if v, ok := readString(norm, "netdevstate"); ok {
		mi.NetDevState = v
	}
	if v, ok := readFloat(norm, "currentbitrate"); ok {
		mi.CurrentBitRate = v
	}
	if v, ok := readFloat(norm, "lastchangetime"); ok {
		mi.LastChangeTime = v
	}
	if v, ok := readFloat(norm, "maxbitratesupported"); ok {
		mi.MaxBitRateSupported = v
	}
	if v, ok := readFloat(norm, "maxbitrateenabled"); ok {
		mi.MaxBitRateEnabled = v
	}
	if v, ok := readString(norm, "currentduplexmode"); ok {
		mi.CurrentDuplexMode = v
	}
	if b, ok := readBool(norm, "duplexmodeenabled"); ok {
		mi.DuplexModeEnabled = b
	}
	return mi, s, nil
}

// GetNetDevStatsTyped parses a getNetDevStats response and returns a typed
// NetDevStats struct populated with recognized counters.
func GetNetDevStatsTyped(data []byte) (NetDevStats, error) {
	var ns NetDevStats
	m, err := ParseNetDevStats(data)
	if err != nil || m == nil {
		return ns, err
	}
	// helper to read numeric fields by name (case-insensitive)
	get := func(name string) float64 {
		if v, ok := m[name]; ok {
			if f, ok2 := anyToFloat(v); ok2 {
				return f
			}
		}
		if v, ok := m[strings.ToLower(name)]; ok {
			if f, ok2 := anyToFloat(v); ok2 {
				return f
			}
		}
		return 0
	}
	ns.RxPackets = get("RxPackets")
	ns.TxPackets = get("TxPackets")
	ns.RxBytes = get("RxBytes")
	ns.TxBytes = get("TxBytes")
	ns.RxErrors = get("RxErrors")
	ns.TxErrors = get("TxErrors")
	ns.RxDropped = get("RxDropped")
	ns.TxDropped = get("TxDropped")
	ns.Multicast = get("Multicast")
	ns.Collisions = get("Collisions")
	ns.RxLengthErrors = get("RxLengthErrors")
	ns.RxOverErrors = get("RxOverErrors")
	ns.RxCrcErrors = get("RxCrcErrors")
	ns.RxFrameErrors = get("RxFrameErrors")
	ns.RxFifoErrors = get("RxFifoErrors")
	ns.RxMissedErrors = get("RxMissedErrors")
	ns.TxAbortedErrors = get("TxAbortedErrors")
	ns.TxCarrierErrors = get("TxCarrierErrors")
	ns.TxFifoErrors = get("TxFifoErrors")
	ns.TxHeartbeatErrors = get("TxHeartbeatErrors")
	ns.TxWindowErrors = get("TxWindowErrors")
	return ns, nil
}

// GetPortParamsFromMIBs extracts PortParams from a getMIBs response. It
// returns best-effort values (zero values when fields are absent).
func GetPortParamsFromMIBs(data []byte, candidate string) (PortParams, error) {
	var pp PortParams
	norm, s, err := ParseMIBs(data, candidate)
	if err != nil || norm == nil {
		return pp, err
	}
	if v, ok := readString(norm, "llintf"); ok {
		pp.LLIntf = v
	}
	if v, ok := readFloat(norm, "currentbitrate"); ok {
		pp.CurrentBitRate = v
	}
	if v, ok := readFloat(norm, "maxbitratesupported"); ok {
		pp.MaxBitRateSupported = v
	}
	if v, ok := readFloat(norm, "maxbitrateenabled"); ok {
		pp.MaxBitRateEnabled = v
	}
	if v, ok := readString(norm, "currentduplexmode"); ok {
		pp.CurrentDuplexMode = v
	}
	if b, ok := readBool(norm, "duplexmodeenabled"); ok {
		pp.DuplexModeEnabled = b
	}
	// Attempt to extract SetPort from the raw status map when available. The
	// UI code often inspects status.base.<CANDIDATE>.LLIntf and takes the
	// first key to determine the device mapping; replicate that behavior.
	if s != nil {
		if baseRaw, ok := s["base"].(map[string]any); ok {
			// try candidate key
			if candRaw, ok := baseRaw[candidate].(map[string]any); ok {
				if llv, ok := candRaw["LLIntf"]; ok {
					switch tv := llv.(type) {
					case map[string]any:
						for k := range tv {
							pp.SetPort = k
							break
						}
					case string:
						pp.SetPort = tv
					}
				}
			}
		}
	}
	return pp, nil
}

package collector

import (
	"strconv"
	"strings"
)

type MIBMetrics struct {
	Up         float64
	Mtu        float64
	TxQueueLen float64
	SpeedMbps  float64
	LastChange float64
	Alias      string
	Flags      string
	Lladdr     string
	Type       string
}

type NetDevStatsMetrics struct {
	Values map[string]float64
}

func BuildMIBMetrics(norm map[string]any, status map[string]any) MIBMetrics {
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
			case string:
				if f, err := strconv.ParseFloat(vv, 64); err == nil {
					return f, true
				}
			}
		}
		return 0, false
	}

	m := MIBMetrics{}
	if b, ok := getBool("Status"); ok && b {
		m.Up = 1.0
	}
	if sStr, ok := getString("NetDevState"); ok && strings.ToLower(sStr) == "up" {
		m.Up = 1.0
	}
	if v, ok := getFloat("MTU"); ok {
		m.Mtu = v
	}
	if v, ok := getFloat("TxQueueLen"); ok {
		m.TxQueueLen = v
	}
	if v, ok := getFloat("CurrentBitRate"); ok {
		m.SpeedMbps = v
	} else if v, ok := getFloat("CurrentBitRateMbps"); ok {
		m.SpeedMbps = v
	}
	if v, ok := getFloat("LastChangeTime"); ok {
		m.LastChange = v
	}
	if sStr, ok := getString("LLAddress"); ok {
		m.Lladdr = sStr
	}
	if sStr, ok := getString("NetDevType"); ok {
		m.Type = sStr
	}
	if sStr, ok := getString("NetDevFlags"); ok {
		m.Flags = sStr
	} else if sStr, ok := getString("Flags"); ok {
		m.Flags = sStr
	}

	if status != nil {
		if am, ok := status["alias"].(map[string]any); ok {
			if aStr, ok := am["Alias"].(string); ok {
				m.Alias = aStr
			}
		}
	}

	return m
}

func BuildNetDevStatsMetrics(data map[string]any) NetDevStatsMetrics {
	res := NetDevStatsMetrics{Values: map[string]float64{}}
	if data == nil {
		return res
	}
	for k, v := range data {
		switch vv := v.(type) {
		case float64:
			res.Values[k] = vv
		case int:
			res.Values[k] = float64(vv)
		case string:
			if f, err := strconv.ParseFloat(vv, 64); err == nil {
				res.Values[k] = f
			}
		}
	}
	return res
}

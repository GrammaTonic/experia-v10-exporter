package nemo

import (
	"encoding/json"
)

// ParseNetDevStats returns a generic map of values parsed from getNetDevStats.
func ParseNetDevStats(data []byte) (map[string]any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var sg map[string]any
	if err := json.Unmarshal(data, &sg); err != nil {
		return nil, err
	}
	// Prefer data map when present
	var dataMap map[string]any
	if d, ok := sg["data"].(map[string]any); ok {
		dataMap = d
	} else if top, ok := sg["status"].(map[string]any); ok {
		dataMap = top
	} else {
		dataMap = sg
	}
	return dataMap, nil
}

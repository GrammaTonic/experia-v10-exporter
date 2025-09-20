package nemo

import (
	"encoding/json"
	"strings"
)

// findStatus scans a decoded JSON map for a child that contains 'base' or 'netdev'.
func findStatus(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	if _, ok := m["base"]; ok {
		return m
	}
	if _, ok := m["netdev"]; ok {
		return m
	}
	for _, v := range m {
		if child, ok := v.(map[string]any); ok {
			if found := findStatus(child); found != nil {
				return found
			}
		}
	}
	return nil
}

// ParseMIBs decodes a getMIBs response into a normalized map[string]any for
// callers to extract fields from. It mirrors the earlier collector logic but
// returns the 'norm' map for further processing.
func ParseMIBs(data []byte, candidate string) (map[string]any, map[string]any, error) {
	if len(data) == 0 {
		return nil, nil, nil
	}
	var g map[string]any
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, nil, err
	}

	var s map[string]any
	if ss := findStatus(g); ss != nil {
		s = ss
	} else if dataMap, ok := g["data"].(map[string]any); ok {
		if ss := findStatus(dataMap); ss != nil {
			s = ss
		}
	}
	if s == nil {
		if top, ok := g["status"].(map[string]any); ok {
			if ss := findStatus(top); ss != nil {
				s = ss
			} else {
				s = top
			}
		}
	}
	if s == nil {
		return nil, nil, nil
	}

	// Merge base and netdev
	var innerMap map[string]any
	if b, ok := s["base"].(map[string]any); ok {
		if im, ok := b[candidate].(map[string]any); ok {
			innerMap = im
		} else if im, ok := b[strings.ToLower(candidate)].(map[string]any); ok {
			innerMap = im
		} else {
			for _, v := range b {
				if im, ok := v.(map[string]any); ok {
					innerMap = im
					break
				}
			}
		}
	}
	if innerMap == nil {
		if nd, ok := s["netdev"].(map[string]any); ok {
			if im, ok := nd[candidate].(map[string]any); ok {
				innerMap = im
			} else if im, ok := nd[strings.ToLower(candidate)].(map[string]any); ok {
				innerMap = im
			} else {
				for _, v := range nd {
					if im, ok := v.(map[string]any); ok {
						innerMap = im
						break
					}
				}
			}
		}
	}

	var ndFirst map[string]any
	if nd, ok := s["netdev"].(map[string]any); ok {
		if im, ok := nd[candidate].(map[string]any); ok {
			ndFirst = im
		} else if im, ok := nd[strings.ToLower(candidate)].(map[string]any); ok {
			ndFirst = im
		} else {
			for _, v := range nd {
				if im, ok := v.(map[string]any); ok {
					ndFirst = im
					break
				}
			}
		}
	}

	norm := map[string]any{}
	for k, v := range innerMap {
		norm[strings.ToLower(k)] = v
	}
	for k, vv := range ndFirst {
		norm[strings.ToLower(k)] = vv
	}

	return norm, s, nil
}

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

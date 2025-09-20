package nemo

import "fmt"

// RequestBodyWanStats returns the JSON body for a hypothetical getWanStats
// NeMo service. It follows the same pattern as other service helpers.
func RequestBodyWanStats() string {
	return fmt.Sprintf(`{"service":"NeMo.WAN","method":"getWanStats","parameters":{}}`)
}

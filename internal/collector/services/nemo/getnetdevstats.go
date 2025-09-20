package nemo

import "fmt"

// RequestBodyStats returns the JSON body for getNetDevStats for a given candidate.
func RequestBodyStats(candidate string) string {
	return fmt.Sprintf(`{"service":"NeMo.Intf.%s","method":"getNetDevStats","parameters":{}}`, candidate)
}

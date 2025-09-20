package nemo

import "fmt"

// RequestBody returns the JSON body for getMIBs for a given candidate (e.g. "ETH0").
func RequestBody(candidate string) string {
	return fmt.Sprintf(`{"service":"NeMo.Intf.%s","method":"getMIBs","parameters":{}}`, candidate)
}

package nemo

import "fmt"

// RequestBodyWanMibs returns the JSON body for a hypothetical getWanMIBs
// NeMo service. The service name mirrors other NeMo calls and takes no
// parameters.
func RequestBodyWanMibs() string {
	return fmt.Sprintf(`{"service":"NeMo.WAN","method":"getWanMIBs","parameters":{}}`)
}

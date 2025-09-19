//go:build test
// +build test

package collector

import "encoding/json"

// test build overrides for apiUrl and jsonMarshal
var apiUrl = "http://%s/ws/NeMo/Intf/lan:getMIBs"
var jsonMarshal = json.Marshal

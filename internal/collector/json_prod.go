//go:build !test
// +build !test

package collector

import "encoding/json"

// jsonMarshal is a variable indirection to allow test builds to override JSON marshal behavior.
var jsonMarshal = json.Marshal

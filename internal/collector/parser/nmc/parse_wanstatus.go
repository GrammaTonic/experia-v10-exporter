package nmc

import "encoding/json"

// WANStatus is the typed representation of the getWANStatus response.
type WANStatus struct {
	Status bool `json:"status"`
	Data   struct {
		LinkType        string `json:"LinkType"`
		LinkState       string `json:"LinkState"`
		MACAddress      string `json:"MACAddress"`
		Protocol        string `json:"Protocol"`
		ConnectionState string `json:"ConnectionState"`
		IPAddress       string `json:"IPAddress"`
	} `json:"data"`
	Errors []struct {
		Error       int    `json:"error"`
		Description string `json:"description"`
		Info        string `json:"info"`
	} `json:"errors"`
}

// ParseWANStatus unmarshals the raw getWANStatus response into the typed struct.
func ParseWANStatus(data []byte) (WANStatus, error) {
	var s WANStatus
	if len(data) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, err
	}
	return s, nil
}

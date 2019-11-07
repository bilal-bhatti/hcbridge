package ha

// Device ...
type Device struct {
	Identifiers  string `json:"identifiers"`
	Name         string `json:"name"`
	SWVersion    string `json:"sw_version"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
}

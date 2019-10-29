package ha

// SwitchDevice represents home assistant discovery information
type SwitchDevice struct {
	Name              string `json:"name"`
	StateTopic        string `json:"state_topic"`
	CommandTopic      string `json:"command_topic"`
	AvailabilityTopic string `json:"availability_topic"`
	UniqueID          string `json:"unique_id"`
	Device            struct {
		Identifiers  string `json:"identifiers"`
		Name         string `json:"name"`
		SWVersion    string `json:"sw_version"`
		Model        string `json:"model"`
		Manufacturer string `json:"manufacturer"`
	} `json:"device"`
}

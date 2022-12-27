package hcbridge

// SwitchConfig represents home assistant discovery information
type SwitchConfig struct {
	Name              string `json:"name"`
	StateTopic        string `json:"state_topic"`
	CommandTopic      string `json:"command_topic"`
	AvailabilityTopic string `json:"availability_topic"`
	UniqueID          string `json:"unique_id"`
	Device            Device `json:"device"`
}

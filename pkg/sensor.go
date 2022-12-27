package hcbridge

// SensorConfig represents home assistant discovery information
type SensorConfig struct {
	Name              string `json:"name"`
	StateTopic        string `json:"state_topic"`
	AvailabilityTopic string `json:"availability_topic"`
	UniqueID          string `json:"unique_id"`
	UnitOfMeasurement string `json:"unit_of_measurement"`
	ExpireAfter       int32  `json:"expire_after"`
	Device            Device `json:"device"`
}

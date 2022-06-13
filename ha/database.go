package ha

import (
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
)

// Database ...
type Database struct {
	Containers map[string]*Container
}

// Container ...
type Container struct {
	*Device
	Things    map[string]interface{} `json:"things"`
	Accessory func() *accessory.Accessory
}

// Thing ...
type Thing struct {
	Config  interface{}
	Service func() *service.Service
}

// NewDatabase ...
func NewDatabase() *Database {
	return &Database{
		Containers: map[string]*Container{},
	}
}

func (db *Database) add(dd Device, id string, thing interface{}) {
	// get device if known
	container, _ := db.Containers[dd.Identifiers]

	// add it if it's new
	if container == nil {
		container = &Container{Device: &dd, Things: make(map[string]interface{})}
		db.Containers[dd.Identifiers] = container
	}

	container.Things[id] = thing
}

// AddSwitch ...
func (db *Database) AddSwitch(dd SwitchConfig) {
	db.add(dd.Device, dd.UniqueID, dd)
}

// AddSensor ...
func (db *Database) AddSensor(dd SensorConfig) {
	db.add(dd.Device, dd.UniqueID, dd)
}

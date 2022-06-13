package main

import (
	"hcbridge/ha"
	"log"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
)

// Database ...
type Database struct {
	devices map[string]VDevice
}

// VDevice ...
type VDevice interface {
	PutService(dd ha.SensorConfig, s *service.Service)
}

// VThermometer ...
type VThermometer struct {
	Accessory  accessory.Thermometer
	ServiceMap map[string]*service.Service
}

// GetAccessory ...
// func (vt *VThermometer) GetAccessory() *accessory.Accessory {
// 	return vt.Accessory.Accessory
// }

// PutService ...
func (vt VThermometer) PutService(dd ha.SensorConfig, s *service.Service) {
	if _, found := vt.ServiceMap[dd.UniqueID]; found {
		log.Println("Service already known, skipping")
	} else {
		vt.ServiceMap[dd.UniqueID] = s
	}
}

// PutSensorDevice ...
func (d *Database) PutSensorDevice(dd ha.SensorConfig) *accessory.Thermometer {
	svd, ok := d.devices[dd.Device.Identifiers]

	// device already known
	if ok {
		// check for service
		// svd.PutService()
	}

	if t, ok := svd.(VThermometer); ok {
		log.Println("t", t)
		if s, ok := t.ServiceMap[dd.UniqueID]; ok {
			log.Println("s", s)
		}
	}

	// svc, _ := vd.ServiceMap[dd.UniqueID]

	info := accessory.Info{
		Name:             dd.Name,
		Manufacturer:     dd.Device.Manufacturer,
		SerialNumber:     dd.Device.Identifiers,
		Model:            dd.Device.Model,
		FirmwareRevision: dd.Device.SWVersion,
	}

	device := accessory.NewTemperatureSensor(info, 25, 10, 65, .1)

	// acc := Thermometer{}
	// acc.Accessory = New(info, TypeThermostat)
	// 	acc.TempSensor = service.NewTemperatureSensor()
	// acc.TempSensor.CurrentTemperature.SetValue(temp)
	// acc.TempSensor.CurrentTemperature.SetMinValue(min)
	// acc.TempSensor.CurrentTemperature.SetMaxValue(max)
	// acc.TempSensor.CurrentTemperature.SetStepValue(steps)

	// acc.AddService(acc.TempSensor.Service)

	// return &acc

	return device
}

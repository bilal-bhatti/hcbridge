package main

import (
	"bytes"
	"encoding/json"
	"hcbridge/ha"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	pinCode = "15018183"
)

// VBridge ...
type VBridge struct {
	bridge   *accessory.Bridge
	devices  []*accessory.Accessory
	stopper  func()
	starting atomic.Value
}

// NewVBridge ...
func NewVBridge() *VBridge {
	bridge := accessory.NewBridge(accessory.Info{
		Name:             "HiTechBridge",
		Manufacturer:     "HiTechBridge",
		SerialNumber:     "VF8R7GHO",
		Model:            "HiTech",
		FirmwareRevision: "OEI-839",
	})

	vb := &VBridge{
		bridge: bridge,
	}
	vb.starting.Store(false)
	return vb
}

// TODO: Add virtual bridge service to report it's status

// OnSwitch ...
func (b *VBridge) OnSwitch(client mqtt.Client, msg mqtt.Message) {
	var dd ha.SwitchDevice
	err := json.NewDecoder(bytes.NewReader(msg.Payload())).Decode(&dd)
	if err != nil {
		panic(err)
	}

	log.Printf("Adding switch %s", dd.Name)

	info := accessory.Info{
		Name:             dd.Name,
		Manufacturer:     dd.Device.Manufacturer,
		SerialNumber:     dd.Device.Identifiers,
		Model:            dd.Device.Model,
		FirmwareRevision: dd.Device.SWVersion,
	}

	device := accessory.NewSwitch(info)
	device.Switch.On.OnValueRemoteUpdate(func(on bool) {
		log.Printf("Received HomeKit update from %s, publishing to MQTT", info.Name)
		if on == true {
			client.Publish(dd.CommandTopic, 0, false, "ON")
		} else {
			client.Publish(dd.CommandTopic, 0, false, "OFF")
		}
	})

	client.Subscribe(dd.StateTopic, 0, func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Status update received from MQTT for %s", info.Name)
		if string(msg.Payload()) == "ON" {
			device.Switch.On.SetValue(true)
		} else {
			device.Switch.On.SetValue(false)
		}
	})

	b.devices = append(b.devices, device.Accessory)
	b.start()
}

// OnSensor ...
func (b *VBridge) OnSensor(client mqtt.Client, msg mqtt.Message) {
	var dd ha.SensorDevice
	err := json.NewDecoder(bytes.NewReader(msg.Payload())).Decode(&dd)
	if err != nil {
		panic(err)
	}

	log.Printf("Adding sensor %s", dd.Name)

	info := accessory.Info{
		Name:             dd.Name,
		Manufacturer:     dd.Device.Manufacturer,
		SerialNumber:     dd.Device.Identifiers,
		Model:            dd.Device.Model,
		FirmwareRevision: dd.Device.SWVersion,
	}

	device := accessory.NewTemperatureSensor(info, 25, 10, 65, .1)

	client.Subscribe(dd.StateTopic, 0, func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Status update received from MQTT for %s with value %v", info.Name, string(msg.Payload()))
		if temp, err := strconv.ParseFloat(string(msg.Payload()), 64); err == nil {
			device.TempSensor.CurrentTemperature.UpdateValue(temp)
		} else {
			log.Printf("Failed to parse sensor reading to float: %v", msg.Payload())
		}
	})

	b.devices = append(b.devices, device.Accessory)
	b.start()
}

// Stop ...
func (b *VBridge) Stop() {
	if b.stopper != nil {
		b.stopper()
	}
}

func (b *VBridge) start() {
	if b.starting.Load() == true {
		// already in process of starting
		return
	}

	b.starting.Store(true)

	// actually start the bridge
	go func() {
		// stop if running
		b.Stop()

		// TODO: debounce better
		log.Println("Starting in 5 seconds ....")
		time.Sleep(5 * time.Second)

		t, err := hc.NewIPTransport(hc.Config{Pin: pinCode}, b.bridge.Accessory, b.devices...)

		b.bridge.OnIdentify(func() {
			log.Println("Identity confirmed " + b.bridge.Info.Identify.Description)
		})

		if err != nil {
			log.Fatal(err)
		}

		b.stopper = func() {
			log.Println("Stopping underlying bridge")
			<-t.Stop()
		}

		log.Printf("Registering %d devices", len(b.devices))
		t.Start()
		b.starting.Store(false)
	}()
}

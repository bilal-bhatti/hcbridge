package hcbridge

import (
	"bytes"
	"encoding/json"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/bep/debounce"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// const (
// 	pinCode = "18058084"
// )

// Bridge ...
type Bridge struct {
	PinCode   string
	bridge    *accessory.Bridge
	stopper   func()
	starting  atomic.Value
	deviceMap map[string]*accessory.Accessory
	debounce  func(f func())
}

// NewVBridge ...
func NewVBridge(pinCode string) *Bridge {
	bridge := accessory.NewBridge(accessory.Info{
		Name:             "ESPHomeBridge",
		Manufacturer:     "ESP Home Bridge",
		SerialNumber:     "AAAXXXXXX",
		Model:            "ESP_Home_Bridge",
		FirmwareRevision: "OEI-AAA",
	})

	bridge.OnIdentify(func() {
		log.Println("Identity confirmed " + bridge.Info.Identify.Description)
	})

	vb := &Bridge{
		PinCode:   pinCode,
		bridge:    bridge,
		debounce:  debounce.New(1000 * time.Millisecond),
		deviceMap: make(map[string]*accessory.Accessory),
	}
	vb.starting.Store(false)
	return vb
}

// TODO: Add virtual bridge service to report it's status

// OnSwitch ...
func (b *Bridge) OnSwitch(client mqtt.Client, msg mqtt.Message) {
	var dd SwitchDevice
	err := json.NewDecoder(bytes.NewReader(msg.Payload())).Decode(&dd)
	if err != nil {
		panic(err)
	}

	if _, ok := b.deviceMap[dd.UniqueID]; ok {
		return
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
		if on {
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

	b.deviceMap[dd.UniqueID] = device.Accessory
	b.debounce(b.start)
	// b.start()
}

// OnSensor ...
func (b *Bridge) OnSensor(client mqtt.Client, msg mqtt.Message) {
	var dd SensorDevice
	err := json.NewDecoder(bytes.NewReader(msg.Payload())).Decode(&dd)
	if err != nil {
		panic(err)
	}

	if dd.UniqueID == "" {
		log.Println("Missing unique id from device", dd.Name)
		return
	}
	if _, ok := b.deviceMap[dd.UniqueID]; ok {
		return
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

	b.deviceMap[dd.UniqueID] = device.Accessory
	b.debounce(b.start)
	// b.start()
}

// Stop ...
func (b *Bridge) Stop() {
	if b.stopper != nil {
		b.stopper()
	}
}

func (b *Bridge) start() {
	log.Println("Starting transport")
	b.Stop()

	// TODO: debounce better
	log.Println("Starting in 5 seconds with pin", b.PinCode)
	time.Sleep(5 * time.Second)

	var devices []*accessory.Accessory
	for _, v := range b.deviceMap {
		devices = append(devices, v)
	}

	t, err := hc.NewIPTransport(hc.Config{Pin: b.PinCode}, b.bridge.Accessory, devices...)
	if err != nil {
		log.Fatal(err)
	}

	b.stopper = func() {
		log.Println("Stopping underlying bridge")
		<-t.Stop()
	}

	log.Printf("Registering %d devices", len(devices))
	t.Start()
}

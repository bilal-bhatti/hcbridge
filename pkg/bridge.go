package bridge

import (
	"bytes"
	"encoding/json"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/bep/debounce"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Bridge ...
type Bridge struct {
	pinCode   string
	bridge    *accessory.Bridge
	stopper   func()
	deviceMap map[string]*accessory.Accessory
	debounce  func(f func())
}

// NewVBridge ...
func NewBridge(pinCode string) *Bridge {
	bridge := accessory.NewBridge(accessory.Info{
		Name:             "MQTTBridge",
		Manufacturer:     "MQTT Bridge",
		SerialNumber:     "AAAXXXXXX",
		Model:            "MQTT_Bridge",
		FirmwareRevision: "OEI-AAA",
	})

	bridge.OnIdentify(func() {
		log.Println("Identity confirmed " + bridge.Info.Identify.Description)
	})

	vb := &Bridge{
		pinCode:   pinCode,
		bridge:    bridge,
		debounce:  debounce.New(1000 * time.Millisecond),
		deviceMap: make(map[string]*accessory.Accessory),
	}
	return vb
}

// OnSwitch ...
func (b *Bridge) OnSwitch(client mqtt.Client, msg mqtt.Message) {
	var dd SwitchConfig

	err := json.NewDecoder(bytes.NewReader(msg.Payload())).Decode(&dd)
	if err != nil {
		panic(err)
	}

	if dd.UniqueID == "" {
		log.Println("Missing unique id from device", dd.Name)
		return
	}

	if _, ok := b.deviceMap[dd.UniqueID]; ok {
		log.Println("Ignore device", dd.UniqueID)
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
}

// OnSensor ...
func (b *Bridge) OnSensor(client mqtt.Client, msg mqtt.Message) {
	var dd SensorConfig

	err := json.NewDecoder(bytes.NewReader(msg.Payload())).Decode(&dd)
	if err != nil {
		panic(err)
	}

	if dd.UniqueID == "" {
		log.Println("Missing unique id from device", dd.Name)
		return
	}

	if _, ok := b.deviceMap[dd.UniqueID]; ok {
		log.Println("Ignore device", dd.UniqueID)
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
		log.Printf("Status update received from MQTT for %s with value %v", dd.Name, string(msg.Payload()))
		if temp, err := strconv.ParseFloat(string(msg.Payload()), 64); err == nil {
			device.TempSensor.CurrentTemperature.UpdateValue(temp)
		} else {
			log.Printf("Failed to parse sensor reading to float: %v", msg.Payload())
		}
	})

	b.deviceMap[dd.UniqueID] = device.Accessory
	b.debounce(b.start)
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

	log.Println("Starting in 5 seconds with pin", b.pinCode)
	time.Sleep(5 * time.Second)

	device_ids := make([]string, 0, len(b.deviceMap))
	for k := range b.deviceMap {
		device_ids = append(device_ids, k)
	}
	sort.Strings(device_ids) // to provide HomeKit an ordered/consistent list of accessories

	devices := make([]*accessory.Accessory, 0, len(b.deviceMap))
	for _, id := range device_ids {
		devices = append(devices, b.deviceMap[id])
	}

	t, err := hc.NewIPTransport(hc.Config{Pin: b.pinCode}, b.bridge.Accessory, devices...)
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

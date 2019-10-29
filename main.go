package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"hcbridge/ha"
	"log"
	"net/url"
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	mqttClient = "HCBRIDGE-U19HGD8"
)

func main() {
	mqttURI := flag.String("mqtt-uri", "tcp://localhost:1883", "Specify MQTT URI")
	flag.Parse()

	bridge := accessory.NewBridge(accessory.Info{
		Name:             "HC Bridge",
		Manufacturer:     "HC Bridge",
		SerialNumber:     "VF8RAW9DB",
		Model:            "HiTech",
		FirmwareRevision: "OEI-839",
	})

	client := connect(mqttClient, *mqttURI)

	devices := []*accessory.Accessory{}

	registerSwitches := func(client mqtt.Client, msg mqtt.Message) {
		var dd ha.SwitchDevice
		err := json.NewDecoder(bytes.NewReader(msg.Payload())).Decode(&dd)
		if err != nil {
			panic(err)
		}

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

		devices = append(devices, device.Accessory)
	}

	client.Subscribe("homeassistant/switch/#", 0, registerSwitches)

	// Hack to read MQTT discovered devices
	// TODO: use mqtt messages to bootstrap devices
	time.Sleep(5 * time.Second)
	log.Printf("Registering %d devices", len(devices))

	t, err := hc.NewIPTransport(hc.Config{Pin: "35018183"}, bridge.Accessory, devices...)

	bridge.OnIdentify(func() {
		log.Println("Identity confirmed " + bridge.Info.Identify.Description)
	})

	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		client.Disconnect(2)
		<-t.Stop()
	})

	t.Start()
}

func connect(clientID string, mqttURI string) mqtt.Client {
	uri, err := url.Parse(mqttURI)
	if err != nil {
		panic(err)
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))
	opts.SetUsername(uri.User.Username())
	password, _ := uri.User.Password()
	opts.SetPassword(password)
	opts.SetClientID(clientID)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MQTT host ", uri.Host)
	return client
}

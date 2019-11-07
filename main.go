package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/brutella/hc"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/thanhpk/randstr"
)

var mqttClient = "HKBR-" + randstr.String(6)

func main() {
	mqttURI := flag.String("mqtt-uri", "tcp://localhost:1883", "Specify MQTT URI")
	flag.Parse()

	client := connect(mqttClient, *mqttURI)

	done := make(chan bool, 1)

	vb := NewVBridge()
	client.Subscribe("homeassistant/switch/#", 0, vb.OnSwitch)
	client.Subscribe("homeassistant/sensor/#", 0, vb.OnSensor)

	hc.OnTermination(func() {
		client.Disconnect(2)
		vb.Stop()
		close(done)
	})

	// block until done
	<-done
	log.Println("Stopped virtual bridge server")
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

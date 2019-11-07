# HC Home Bridge

A simple HomeKit bridge between ESPHome enabled devices that broadcast HomeAssistant discovery messages and a HomeKit Hub like Apple TV etc.

Supports:
 - Switches (on/off)
 - Temperature Sensors (basic)
 
Code is fairly basic. Mostly a test to see if it would work but seems fairly stable and so will add more accessories over time.

## Build
`go build -o hcbridged`.

## Build for Linux
`env GOOS=linux GOARCH=amd64 go build -o hcbridged`

## Systemd configuration
To use as a systemd service
 - create hcbridge user `useradd -s /usr/sbin/nologin -r -m hcbridge`
 - copy binary `cp hcbridged /opt/bin/hcbridged`
 - `cp systemd/hcbridged /etc/default/hcbridged`
 - update `vi /etc/default/hcbridged` to point to your MQTT host
 - `cp systemd/hcbridged.service /etc/systemd/system/hcbridged.service`
 - run `systemctl enable hcbridged`
 - run `systemctl start hcbridged`
 - run `journalctl -f` to check logs and see it started

# Credits
github.com/brutella/hc
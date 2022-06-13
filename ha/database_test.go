package ha

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseContainerHierarchy(t *testing.T) {
	sw := SwitchConfig{
		Name:     "Test Switch One",
		UniqueID: "test_switch_1",
		Device: Device{
			Identifiers: "1",
		},
	}

	sw2 := SwitchConfig{
		Name:     "Test Switch Two",
		UniqueID: "test_switch_2",
		Device: Device{
			Identifiers: "2",
		},
	}

	sd := SensorConfig{
		Name:     "Test Sensor One",
		UniqueID: "test_sensor_1",
		Device: Device{
			Identifiers: "1",
		},
	}

	sd2 := SensorConfig{
		Name:     "Test Sensor Two",
		UniqueID: "test_sensor_2",
		Device: Device{
			Identifiers: "3",
		},
	}

	sd3 := SensorConfig{
		Name:     "Test Sensor Three",
		UniqueID: "test_sensor_3",
		Device: Device{
			Identifiers: "3",
		},
	}

	sd4 := SensorConfig{
		Name:     "Test Sensor Four",
		UniqueID: "test_sensor_4",
		Device: Device{
			Identifiers: "3",
		},
	}

	db := NewDatabase()

	db.AddSensor(sd2)
	db.AddSwitch(sw)
	db.AddSensor(sd)
	db.AddSensor(sd3)
	db.AddSwitch(sw2)
	db.AddSensor(sd4)

	assert.Equal(t, len(db.Containers), 3, "Should be 3")

	single := db.Containers["2"]
	assert.Equal(t, len(single.Things), 1, "Should be 1")

	multi := db.Containers["1"]
	assert.Equal(t, len(multi.Things), 2, "Should be 2")
	assert.Equal(t, multi.Things["test_switch_1"].(SwitchConfig).Name, "Test Switch One")
	assert.Equal(t, multi.Things["test_sensor_1"].(SensorConfig).Name, "Test Sensor One")

	multisensor := db.Containers["3"]
	assert.Equal(t, len(multisensor.Things), 3, "Should be 3")
	assert.Equal(t, multisensor.Things["test_sensor_2"].(SensorConfig).Name, "Test Sensor Two")
	assert.Equal(t, multisensor.Things["test_sensor_3"].(SensorConfig).Name, "Test Sensor Three")
	assert.Equal(t, multisensor.Things["test_sensor_4"].(SensorConfig).Name, "Test Sensor Four")
}

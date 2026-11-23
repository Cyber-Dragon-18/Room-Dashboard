package devices

import (
	"fmt"
	"os"
)

// A factory takes a device's external ID and builds a Light.
type LightFactory func(externalID string) (Light, error)

// The registry: driver name -> how to build it.
var lightDrivers = map[string]LightFactory{
	"mock": func(id string) (Light, error) {
		return NewMockLight(), nil
	},
	"tuya": func(id string) (Light, error) {
		return NewTuyaLight(os.Getenv("TUYA_ID"), os.Getenv("TUYA_SECRET"), id), nil
	},
}

// NewLight looks up the driver and builds the light.
func NewLight(driver, externalID string) (Light, error) {
	factory, ok := lightDrivers[driver]
	if !ok {
		return nil, fmt.Errorf("unknown light driver %q", driver)
	}
	return factory(externalID)
}

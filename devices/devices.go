package devices

import (
	"strings"
)

type Device struct {
	Name    string
	Payload []byte
}

// NewDevice creates a new Device and determines its type based on the first two characters of its name.
func NewDevice(name string, payload []byte) Device {
	return Device{
		Name:    name,
		Payload: payload,
	}
}

// GetDeviceType determines the device type based on the first two characters of the device name.
func (d *Device) GetDeviceType() string {
	switch strings.ToLower(d.Name[:2]) {
	case "co":
		return "Core"
	case "la":
		return "Label"
	// Add more cases as needed.
	default:
		return "Unknown"
	}
}

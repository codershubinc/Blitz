package utils

import (
	"fmt"
	"regexp"
	"strings"
)

type BluetoothDevice struct {
	Name       string `json:"name"`
	MACAddress string `json:"mac"`
	Battery    int    `json:"battery"` // -1 if not available
	Icon       string `json:"icon"`
	Connected  bool   `json:"connected"`
}

// GetBluetoothDevices returns a list of connected Bluetooth devices with battery info
func GetBluetoothDevices() ([]BluetoothDevice, error) {
	// Get list of connected devices
	output, err := SpawnProcess("bluetoothctl", []string{"devices", "Connected"})
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	devices := []BluetoothDevice{}

	// Parse each device
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Format: "Device MAC_ADDRESS Device_Name"
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		mac := parts[1]
		name := strings.Join(parts[2:], " ")

		// Get device info (including battery)
		device := BluetoothDevice{
			Name:       name,
			MACAddress: mac,
			Battery:    -1, // default: not available
			Icon:       "bluetooth",
			Connected:  true,
		}

		// Get detailed info for this device
		infoOutput, err := SpawnProcess("bluetoothctl", []string{"info", mac})
		if err == nil {
			infoStr := string(infoOutput)

			// Extract battery percentage if available
			batteryRegex := regexp.MustCompile(`Battery Percentage: [^\(]*\((\d+)\)`)
			if matches := batteryRegex.FindStringSubmatch(infoStr); len(matches) > 1 {
				battery := 0
				if _, err := fmt.Sscanf(matches[1], "%d", &battery); err == nil {
					device.Battery = battery
				}
			}

			// Extract icon if available
			iconRegex := regexp.MustCompile(`Icon: (.+)`)
			if matches := iconRegex.FindStringSubmatch(infoStr); len(matches) > 1 {
				device.Icon = strings.TrimSpace(matches[1])
			}
		}

		devices = append(devices, device)
	}

	return devices, nil
}

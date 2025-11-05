package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// BluetoothDevice represents a connected Bluetooth device with its info
type BluetoothDevice struct {
	Name       string `json:"name"`
	MacAddress string `json:"macAddress"`
	Connected  bool   `json:"connected"`
	Battery    int    `json:"battery"` // -1 if not available
	Icon       string `json:"icon"`
}

// GetBluetoothDevices returns all connected Bluetooth devices with battery info
// Uses bluetoothctl (part of BlueZ) which is available on all Linux distros
func GetBluetoothDevices() ([]BluetoothDevice, error) {
	// Step 1: Get list of connected devices
	// Command: bluetoothctl devices Connected
	// Output: "Device AA:BB:CC:DD:EE:FF Sony WH-1000XM4"
	output, err := SpawnProcess("bluetoothctl", []string{"devices", "Connected"})
	if err != nil {
		// bluetoothctl not available or no devices - return empty array (not fatal)
		return []BluetoothDevice{}, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	devices := []BluetoothDevice{}

	// Regex to parse: "Device AA:BB:CC:DD:EE:FF Device Name Here"
	deviceRegex := regexp.MustCompile(`^Device\s+([0-9A-F:]+)\s+(.+)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := deviceRegex.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue // Skip malformed lines
		}

		macAddress := matches[1]
		deviceName := matches[2]

		// Step 2: Get detailed info for this device (battery & icon)
		battery, icon := getDeviceDetails(macAddress)

		devices = append(devices, BluetoothDevice{
			Name:       deviceName,
			MacAddress: macAddress,
			Connected:  true,
			Battery:    battery,
			Icon:       icon,
		})
	}

	return devices, nil
}

// getDeviceDetails fetches battery and icon info for a specific device
// Uses: bluetoothctl info <MAC_ADDRESS>
func getDeviceDetails(macAddress string) (battery int, icon string) {
	// Defaults
	battery = -1 // -1 means battery info not available
	icon = "bluetooth"

	// Get detailed device info
	output, err := SpawnProcess("bluetoothctl", []string{"info", macAddress})
	if err != nil {
		return battery, icon
	}

	// Parse the output line by line
	lines := strings.Split(string(output), "\n")

	// Regex for battery: "	Battery Percentage: 0x55 (85)"
	batteryRegex := regexp.MustCompile(`Battery Percentage:\s+0x[0-9A-Fa-f]+\s+\((\d+)\)`)

	// Regex for icon: "	Icon: audio-card"
	iconRegex := regexp.MustCompile(`Icon:\s+(.+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for battery percentage
		if matches := batteryRegex.FindStringSubmatch(line); len(matches) == 2 {
			fmt.Sscanf(matches[1], "%d", &battery)
		}

		// Check for icon type
		if matches := iconRegex.FindStringSubmatch(line); len(matches) == 2 {
			icon = strings.TrimSpace(matches[1])
		}
	}

	return battery, icon
}

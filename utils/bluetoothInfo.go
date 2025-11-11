package utils

import (
	"fmt"
	"regexp"
	"strings"
)

type BluetoothDevice struct {
	Name         string `json:"name"`
	MACAddress   string `json:"mac"`
	Battery      int    `json:"battery"`      // Average battery, -1 if not available
	BatteryLeft  int    `json:"batteryLeft"`  // Left earbud battery, -1 if not available
	BatteryRight int    `json:"batteryRight"` // Right earbud battery, -1 if not available
	BatteryCase  int    `json:"batteryCase"`  // Case battery, -1 if not available
	Icon         string `json:"icon"`
	Connected    bool   `json:"connected"`
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
			Name:         name,
			MACAddress:   mac,
			Battery:      -1, // default: not available
			BatteryLeft:  -1,
			BatteryRight: -1,
			BatteryCase:  -1,
			Icon:         "bluetooth",
			Connected:    true,
		}

		// Get detailed info for this device
		infoOutput, err := SpawnProcess("bluetoothctl", []string{"info", mac})
		if err == nil {
			infoStr := string(infoOutput)

			// Extract battery percentage if available (average/main)
			batteryRegex := regexp.MustCompile(`Battery Percentage: [^\(]*\((\d+)\)`)
			if matches := batteryRegex.FindStringSubmatch(infoStr); len(matches) > 1 {
				battery := 0
				if _, err := fmt.Sscanf(matches[1], "%d", &battery); err == nil {
					device.Battery = battery
				}
			}

			// Extract individual battery percentages for Galaxy Buds and similar devices
			// Look for patterns like "Battery Percentage: 0x00nn (nn)" for left, right, case
			batteryLines := strings.Split(infoStr, "\n")
			for _, line := range batteryLines {
				if strings.Contains(line, "Battery Percentage") {
					// Try to extract multiple battery values
					// Galaxy Buds format can vary, try different patterns
					if strings.Contains(strings.ToLower(line), "left") {
						if matches := batteryRegex.FindStringSubmatch(line); len(matches) > 1 {
							fmt.Sscanf(matches[1], "%d", &device.BatteryLeft)
						}
					} else if strings.Contains(strings.ToLower(line), "right") {
						if matches := batteryRegex.FindStringSubmatch(line); len(matches) > 1 {
							fmt.Sscanf(matches[1], "%d", &device.BatteryRight)
						}
					} else if strings.Contains(strings.ToLower(line), "case") {
						if matches := batteryRegex.FindStringSubmatch(line); len(matches) > 1 {
							fmt.Sscanf(matches[1], "%d", &device.BatteryCase)
						}
					}
				}
			}

			// For Galaxy Buds, try using the 'Battery' GATT characteristic directly
			// This might require parsing UUID-based battery info
			parseGalaxyBudsBattery(&device, infoStr)

			// Try to get individual battery info using GalaxyBudsClient or earbuds CLI
			if strings.Contains(strings.ToLower(device.Name), "galaxy buds") ||
				strings.Contains(strings.ToLower(device.Name), "buds") {
				tryGalaxyBudsTools(&device, mac)
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

// parseGalaxyBudsBattery attempts to extract individual battery info for Galaxy Buds
// NOTE: Standard bluetoothctl only exposes combined battery for Galaxy Buds.
// Individual L/R/Case batteries require Samsung's proprietary protocol (e.g., galaxybudsclient).
// This function will work if multiple Battery Percentage entries are present in the output.
func parseGalaxyBudsBattery(device *BluetoothDevice, infoStr string) {
	// Try to find multiple battery percentage entries
	batteryRegex := regexp.MustCompile(`Battery Percentage: 0x([0-9a-fA-F]+) \((\d+)\)`)
	matches := batteryRegex.FindAllStringSubmatch(infoStr, -1)

	// If we find multiple battery readings, assume order: Left, Right, Case
	if len(matches) >= 3 {
		fmt.Sscanf(matches[0][2], "%d", &device.BatteryLeft)
		fmt.Sscanf(matches[1][2], "%d", &device.BatteryRight)
		fmt.Sscanf(matches[2][2], "%d", &device.BatteryCase)
	} else if len(matches) == 2 {
		fmt.Sscanf(matches[0][2], "%d", &device.BatteryLeft)
		fmt.Sscanf(matches[1][2], "%d", &device.BatteryRight)
	}
	// If only 1 match, it's already captured in device.Battery by the caller
}

// tryGalaxyBudsTools attempts to get individual battery info using specialized Galaxy Buds tools
func tryGalaxyBudsTools(device *BluetoothDevice, mac string) {
	// Try GalaxyBudsClient CLI if available (https://github.com/ThePBone/GalaxyBudsClient)
	// Install: yay -S galaxybudsclient-bin
	output, err := SpawnProcess("galaxybudsclient", []string{"--address", mac, "--get-battery"})
	if err == nil {
		parseGalaxyBudsClientOutput(device, string(output))
		return
	}

	// Alternative: Try custom D-Bus battery reading for Samsung devices
	tryDBusBatteryRead(device, mac)
}

// parseGalaxyBudsClientOutput parses output from GalaxyBudsClient
func parseGalaxyBudsClientOutput(device *BluetoothDevice, output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		re := regexp.MustCompile(`(\d+)%?`)

		if strings.Contains(lower, "left") {
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				var percent int
				fmt.Sscanf(matches[1], "%d", &percent)
				device.BatteryLeft = percent
			}
		} else if strings.Contains(lower, "right") {
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				var percent int
				fmt.Sscanf(matches[1], "%d", &percent)
				device.BatteryRight = percent
			}
		} else if strings.Contains(lower, "case") {
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				var percent int
				fmt.Sscanf(matches[1], "%d", &percent)
				device.BatteryCase = percent
			}
		}
	}
}

// tryDBusBatteryRead attempts to read battery via D-Bus
func tryDBusBatteryRead(device *BluetoothDevice, mac string) {
	// Try to read from UPower D-Bus interface
	// Galaxy Buds might expose multiple battery devices
	dbusPath := strings.ReplaceAll(mac, ":", "_")

	// Query all battery devices
	output, err := SpawnProcess("dbus-send", []string{
		"--system",
		"--print-reply",
		"--dest=org.bluez",
		fmt.Sprintf("/org/bluez/hci0/dev_%s", dbusPath),
		"org.freedesktop.DBus.Properties.GetAll",
		"string:org.bluez.Battery1",
	})

	if err == nil {
		// Parse D-Bus output for battery percentage
		// This is a simplified version - full implementation would parse D-Bus properly
		_ = output
	}
}

// parseEarbudsOutput parses JSON output from earbuds tools (legacy - kept for compatibility)
func parseEarbudsOutput(device *BluetoothDevice, output string) {
	// Simple parsing - look for battery values in output
	// earbuds tool outputs format varies, so we'll parse common patterns
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "left") && strings.Contains(lower, "battery") {
			re := regexp.MustCompile(`(\d+)%?`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				var percent int
				fmt.Sscanf(matches[1], "%d", &percent)
				device.BatteryLeft = percent
			}
		} else if strings.Contains(lower, "right") && strings.Contains(lower, "battery") {
			re := regexp.MustCompile(`(\d+)%?`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				var percent int
				fmt.Sscanf(matches[1], "%d", &percent)
				device.BatteryRight = percent
			}
		} else if strings.Contains(lower, "case") && strings.Contains(lower, "battery") {
			re := regexp.MustCompile(`(\d+)%?`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				var percent int
				fmt.Sscanf(matches[1], "%d", &percent)
				device.BatteryCase = percent
			}
		}
	}
}

// tryDirectGalaxyBudsRead is deprecated - use GalaxyBudsClient instead
func tryDirectGalaxyBudsRead(device *BluetoothDevice, mac string) {
	// Placeholder - users should install GalaxyBudsClient for full functionality
	// Install on Arch: yay -S galaxybudsclient-bin
	_ = device
	_ = mac
}

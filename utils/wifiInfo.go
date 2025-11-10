package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type WiFiInfo struct {
	SSID           string  `json:"ssid"`
	SignalStrength int     `json:"signalStrength"` // Signal strength (0-100)
	LinkSpeed      int     `json:"linkSpeed"`      // Link speed in Mbps
	Frequency      string  `json:"frequency"`      // e.g., "5 GHz" or "2.4 GHz"
	Security       string  `json:"security"`       // Security type (WPA2, WPA3, etc.)
	IPAddress      string  `json:"ipAddress"`      // IP address of the device
	Connected      bool    `json:"connected"`
	DownloadSpeed  float64 `json:"downloadSpeed"` // Current download speed in Mbps
	UploadSpeed    float64 `json:"uploadSpeed"`   // Current upload speed in Mbps
	InterfaceName  string  `json:"interface"`     // Network interface name
	UnitOfSpeed    string  `json:"unitOfSpeed"`   // Unit of speed (Mbps, Kbps, etc.)
}

var (
	lastRxBytes   uint64
	lastTxBytes   uint64
	lastCheckTime time.Time
)

// GetWiFiInfo returns current WiFi connection info and network speed
func GetWiFiInfo() (*WiFiInfo, error) {
	// Get active WiFi connection using nmcli
	output, err := SpawnProcess("nmcli", []string{"-t", "-f", "ACTIVE,SSID,SIGNAL,FREQ,DEVICE", "dev", "wifi"})
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	info := &WiFiInfo{
		Connected:   false,
		UnitOfSpeed: "Mbps",
	}

	// Find the active connection (starts with "yes:")
	for _, line := range lines {
		if strings.HasPrefix(line, "yes:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 5 {
				info.Connected = true
				info.SSID = parts[1]

				// Parse signal strength
				if signal, err := strconv.Atoi(parts[2]); err == nil {
					info.SignalStrength = signal
				}

				info.Frequency = parts[3]
				info.InterfaceName = parts[4]
				break
			}
		}
	}

	if !info.Connected {
		return info, nil
	}

	// Get additional connection details (security, IP, link speed)
	getConnectionDetails(info)

	// Get network speed for the interface
	downloadSpeed, uploadSpeed := getCurrentNetworkSpeed(info.InterfaceName)
	info.DownloadSpeed = downloadSpeed
	info.UploadSpeed = uploadSpeed

	return info, nil
}

// getCurrentNetworkSpeed calculates current download/upload speed in Mbps
func getCurrentNetworkSpeed(interfaceName string) (float64, float64) {
	if interfaceName == "" {
		return 0, 0
	}

	rxPath := fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", interfaceName)
	txPath := fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", interfaceName)

	// Read current byte counts
	rxData, err := os.ReadFile(rxPath)
	if err != nil {
		return 0, 0
	}
	txData, err := os.ReadFile(txPath)
	if err != nil {
		return 0, 0
	}

	rxBytes, _ := strconv.ParseUint(strings.TrimSpace(string(rxData)), 10, 64)
	txBytes, _ := strconv.ParseUint(strings.TrimSpace(string(txData)), 10, 64)

	now := time.Now()

	// First call - just store values
	if lastCheckTime.IsZero() {
		lastRxBytes = rxBytes
		lastTxBytes = txBytes
		lastCheckTime = now
		return 0, 0
	}

	// Calculate time difference in seconds
	timeDiff := now.Sub(lastCheckTime).Seconds()
	if timeDiff == 0 {
		return 0, 0
	}

	// Calculate speed in Mbps
	downloadSpeed := float64(rxBytes-lastRxBytes) * 8 / timeDiff / 1_000_000
	uploadSpeed := float64(txBytes-lastTxBytes) * 8 / timeDiff / 1_000_000

	// Update last values
	lastRxBytes = rxBytes
	lastTxBytes = txBytes
	lastCheckTime = now

	return downloadSpeed, uploadSpeed
}

// getConnectionDetails retrieves additional WiFi connection details like security, IP, and link speed
func getConnectionDetails(info *WiFiInfo) {
	// Get active connection name and details using nmcli
	connOutput, err := SpawnProcess("nmcli", []string{"-t", "-f", "NAME,DEVICE", "connection", "show", "--active"})
	if err != nil {
		return
	}

	var connectionName string
	lines := strings.Split(strings.TrimSpace(string(connOutput)), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) >= 2 && parts[1] == info.InterfaceName {
			connectionName = parts[0]
			break
		}
	}

	if connectionName == "" {
		return
	}

	// Get detailed connection info
	detailOutput, err := SpawnProcess("nmcli", []string{"-t", "-f", "802-11-wireless-security.key-mgmt,IP4.ADDRESS,GENERAL.DEVICE", "connection", "show", connectionName})
	if err == nil {
		detailLines := strings.Split(strings.TrimSpace(string(detailOutput)), "\n")
		for _, line := range detailLines {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "802-11-wireless-security.key-mgmt":
					if value != "" && value != "--" {
						info.Security = strings.ToUpper(value)
					} else {
						info.Security = "Open"
					}
				case "IP4.ADDRESS":
					if value != "" && value != "--" {
						// Extract IP address (remove /24 suffix if present)
						ipParts := strings.Split(value, "/")
						info.IPAddress = ipParts[0]
					}
				}
			}
		}
	}

	// Get link speed using iw command
	iwOutput, err := SpawnProcess("iw", []string{"dev", info.InterfaceName, "link"})
	if err == nil {
		iwLines := strings.Split(string(iwOutput), "\n")
		for _, line := range iwLines {
			if strings.Contains(line, "tx bitrate:") {
				// Extract speed (e.g., "tx bitrate: 866.7 MBit/s")
				parts := strings.Fields(line)
				for i, part := range parts {
					if part == "bitrate:" && i+1 < len(parts) {
						if speed, err := strconv.ParseFloat(parts[i+1], 64); err == nil {
							info.LinkSpeed = int(speed)
						}
						break
					}
				}
			}
		}
	}
}

package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type WiFiInfo struct {
	SSID          string  `json:"ssid"`
	Signal        int     `json:"signal"`         // Signal strength (0-100)
	Frequency     string  `json:"frequency"`      // e.g., "5 GHz" or "2.4 GHz"
	DownloadSpeed float64 `json:"download_speed"` // Mbps
	UploadSpeed   float64 `json:"upload_speed"`   // Mbps
	Connected     bool    `json:"connected"`
	InterfaceName string  `json:"interface_name"`
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
		Connected: false,
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
					info.Signal = signal
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

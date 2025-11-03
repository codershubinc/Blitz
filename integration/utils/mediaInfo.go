package utils

import (
	"os/exec"
	"strings"
)

type MediaInfo struct {
	Title    string
	Artist   string
	Album    string
	Artwork  string
	Position string
	Length   string
	Status   string
	Player   string
}

func GetPlayerInfo() (MediaInfo, error) {
	// Run one command to get everything: title, artwork, artist, album, position, length, status, player name
	// Format: title|||artUrl|||artist|||album|||position|||length|||status|||playerName
	cmd := exec.Command("playerctl", "metadata", "--format", "{{title}}|||{{mpris:artUrl}}|||{{artist}}|||{{album}}|||{{duration(position)}}|||{{duration(mpris:length)}}|||{{status}}|||{{playerName}}")
	output, err := cmd.Output()

	if err != nil {
		// playerctl not available or no player running
		return MediaInfo{}, err
	}

	// Split the output by |||
	parts := strings.Split(strings.TrimSpace(string(output)), "|||")

	// Make sure we have all 8 parts (if not, player might not be running)
	if len(parts) < 8 {
		return MediaInfo{}, nil
	}

	// Parse each part
	mediaInfo := MediaInfo{
		Title:    strings.TrimSpace(parts[0]),
		Artwork:  strings.TrimSpace(parts[1]),
		Artist:   strings.TrimSpace(parts[2]),
		Album:    strings.TrimSpace(parts[3]),
		Position: strings.TrimSpace(parts[4]),
		Length:   strings.TrimSpace(parts[5]),
		Status:   strings.TrimSpace(parts[6]),
		Player:   strings.TrimSpace(parts[7]),
	}
	

	return mediaInfo, nil
}

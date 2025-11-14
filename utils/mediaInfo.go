package utils

import (
	"fmt"
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
	output, err := SpawnProcess(
		`playerctl`,
		[]string{"metadata", `--format`, `{{title}}|||{{mpris:artUrl}}|||{{artist}}|||{{album}}|||{{position}}|||{{mpris:length}}|||{{status}}|||{{playerName}}`})
	if err != nil {
		// playerctl not available or no player running
		fmt.Print("Error getting player info:", err)
		return MediaInfo{}, err
	}

	// Split the output by |||
	parts := strings.Split(strings.TrimSpace(string(output)), "|||")

	// Make sure we have all 8 parts (if not, player might not be running)
	if len(parts) < 8 {
		return MediaInfo{}, nil
	}
	// HandleArtworkRequest(strings.TrimSpace(parts[1]))
	artwork, err := HandleArtworkRequest(strings.TrimSpace(parts[1]))
	if err != nil {
		artwork = ""
	}

	// Parse each part
	mediaInfo := MediaInfo{
		Title:    strings.TrimSpace(parts[0]),
		Artwork:  artwork,
		Artist:   strings.TrimSpace(parts[2]),
		Album:    strings.TrimSpace(parts[3]),
		Position: strings.TrimSpace(parts[4]),
		Length:   strings.TrimSpace(parts[5]),
		Status:   strings.TrimSpace(parts[6]),
		Player:   strings.TrimSpace(parts[7]),
	}

	return mediaInfo, nil
}

func GetAllActivePlayers() ([]string, error) {
	// Run playerctl to get the list of all active players
	output, err := SpawnProcess(
		`playerctl`,
		[]string{"-l"},
	)
	if err != nil {
		// playerctl not available or no players running
		fmt.Print("Error getting active players:", err)
		return []string{}, err
	}

	// Split the output by new lines to get individual player names
	players := strings.Split(strings.TrimSpace(string(output)), "\n")

	return players, nil
}

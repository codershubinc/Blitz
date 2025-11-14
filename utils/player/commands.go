package player

import (
	"Quazaar/utils"
	"fmt"
	"log"
)

// PlayerCommand defines the structure of player commands received from clients
type PlayerCommand struct {
	Command string `json:"command"`
}

// HandlePlayerCommand processes player control commands from WebSocket clients
func HandlePlayerCommand(cmdData map[string]interface{}) error {
	command, ok := cmdData["command"].(string)
	if !ok {
		return fmt.Errorf("invalid command format")
	}

	log.Printf("🎮 Executing player command: %s", command)

	switch command {
	case "play":
		return Play()
	case "pause":
		return Pause()
	case "player_toggle", "play-pause":
		return TogglePlayPause()
	case "next", "player_next":
		return Next()
	case "prev", "player_prev":
		return Previous()
	case "volume_up", "player_volume_up":
		return VolumeUp()
	case "volume_down", "player_volume_down":
		return VolumeDown()
	case "stop":
		return Stop()
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// Play starts media playback
func Play() error {
	log.Println("▶️  Play")
	_, err := utils.SpawnProcess("playerctl", []string{"play"})
	if err != nil {
		log.Printf("❌ Play failed: %v", err)
		return err
	}
	log.Println("✅ Play successful")
	return nil
}

// Pause pauses media playback
func Pause() error {
	log.Println("⏸️  Pause")
	_, err := utils.SpawnProcess("playerctl", []string{"pause"})
	if err != nil {
		log.Printf("❌ Pause failed: %v", err)
		return err
	}
	log.Println("✅ Pause successful")
	return nil
}

// TogglePlayPause toggles between play and pause states
func TogglePlayPause() error {
	log.Println("🔄 Toggle Play/Pause")
	_, err := utils.SpawnProcess("playerctl", []string{"play-pause"})
	if err != nil {
		log.Printf("❌ Toggle failed: %v", err)
		return err
	}
	log.Println("✅ Toggle successful")
	return nil
}

// Next skips to the next track
func Next() error {
	log.Println("⏭️  Next Track")
	_, err := utils.SpawnProcess("playerctl", []string{"next"})
	if err != nil {
		log.Printf("❌ Next track failed: %v", err)
		return err
	}
	log.Println("✅ Next track successful")
	return nil
}

// Previous plays the previous track
func Previous() error {
	log.Println("⏮️  Previous Track")
	_, err := utils.SpawnProcess("playerctl", []string{"previous"})
	if err != nil {
		log.Printf("❌ Previous track failed: %v", err)
		return err
	}
	log.Println("✅ Previous track successful")
	return nil
}

// VolumeUp increases the volume
func VolumeUp() error {
	log.Println("🔊 Volume Up")
	_, err := utils.SpawnProcess("playerctl", []string{"volume", "0.05+"})
	if err != nil {
		log.Printf("❌ Volume up failed: %v", err)
		return err
	}
	log.Println("✅ Volume up successful")
	return nil
}

// VolumeDown decreases the volume
func VolumeDown() error {
	log.Println("🔉 Volume Down")
	_, err := utils.SpawnProcess("playerctl", []string{"volume", "0.05-"})
	if err != nil {
		log.Printf("❌ Volume down failed: %v", err)
		return err
	}
	log.Println("✅ Volume down successful")
	return nil
}

// Stop stops media playback
func Stop() error {
	log.Println("⛔ Stop")
	_, err := utils.SpawnProcess("playerctl", []string{"stop"})
	if err != nil {
		log.Printf("❌ Stop failed: %v", err)
		return err
	}
	log.Println("✅ Stop successful")
	return nil
}

// Seek moves the playback position (in seconds)
func Seek(seconds int64) error {
	log.Printf("📍 Seek to %d seconds", seconds)
	_, err := utils.SpawnProcess("playerctl", []string{"position", fmt.Sprintf("%d", seconds)})
	if err != nil {
		log.Printf("❌ Seek failed: %v", err)
		return err
	}
	log.Println("✅ Seek successful")
	return nil
}

// SeekRelative moves the playback position relative to current position (in seconds)
func SeekRelative(seconds int64) error {
	sign := "+"
	if seconds < 0 {
		sign = ""
	}
	seekStr := fmt.Sprintf("%s%d", sign, seconds)
	log.Printf("📍 Seek relative: %s seconds", seekStr)
	_, err := utils.SpawnProcess("playerctl", []string{"position", seekStr})
	if err != nil {
		log.Printf("❌ Seek relative failed: %v", err)
		return err
	}
	log.Println("✅ Seek relative successful")
	return nil
}

// ListAvailableCommands returns all available player commands
func ListAvailableCommands() []string {
	return []string{
		"play",
		"pause",
		"player_toggle",
		"next",
		"prev",
		"volume_up",
		"volume_down",
		"stop",
	}
}

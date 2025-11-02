package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// --- This is where you define your allowed commands ---
// This is a CRITICAL security step.
var ALLOWED_COMMANDS = map[string][]string{
	"update":       {"sudo", "pacman", "-Syu"},
	"list_home":    {"ls", "-l", "/home/swap/"}, // Make sure this path is correct for you
	"status":       {"git", "status"},
	"open_firefox": {"firefox", "--new-window"},
	"open_edge":    {"microsoft-edge-beta"},
	"open_vscode":  {"code-insiders"},
	"open_postman": {"postman"},
}

// This upgrades our HTTP connection to a WebSocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections (not for production!)
	},
}

// Define the message structures for JSON
type ClientMessage struct {
	Command string `json:"command"`
}

type ServerResponse struct {
	Status  string `json:"status"`
	Command string `json:"command,omitempty"`
	Output  string `json:"output,omitempty"`
	Message string `json:"message,omitempty"`
	Artwork string `json:"artwork,omitempty"`
}

// This function handles each client connection
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()
	log.Printf("Client connected: %s", conn.RemoteAddr())

	// We'll use a single writer goroutine to avoid concurrent writes to the websocket.
	messages := make(chan ServerResponse)

	// writer goroutine: writes all outgoing messages to the websocket
	go func() {
		for resp := range messages {
			if err := conn.WriteJSON(resp); err != nil {
				log.Println("Failed to write JSON response:", err)
				return
			}
		}
	}()

	// player poller: periodically checks playerctl and sends updates
	quitPlayerPoll := make(chan struct{})
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				info, artwork, _ := getPlayerInfo()
				messages <- ServerResponse{Status: "player", Output: info, Artwork: artwork}
			case <-quitPlayerPoll:
				return
			}
		}
	}()

	// Loop forever, reading messages from this client
	for {
		// Read a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Client disconnected:", err)
			break
		}

		if messageType == websocket.TextMessage {
			log.Printf("Received message: %s", string(p))

			var msg ClientMessage
			var response ServerResponse

			// Unmarshal the JSON message
			if err := json.Unmarshal(p, &msg); err != nil {
				log.Println("Error parsing JSON:", err)
				response = ServerResponse{Status: "error", Message: "Invalid JSON format"}
				messages <- response
				continue
			}

			// Special-case: if client requests "player_info", return current info immediately
			if msg.Command == "player_info" {
				info, artwork, _ := getPlayerInfo()
				messages <- ServerResponse{Status: "player", Output: info, Artwork: artwork}
				continue
			}

			// Check if the command is in our allow-list
			if commandToRun, ok := ALLOWED_COMMANDS[msg.Command]; ok {
				log.Printf("Running command: %v", commandToRun)

				// Run all commands in the background (like an application launcher)
				cmd := exec.Command(commandToRun[0], commandToRun[1:]...)

				err := cmd.Start()
				if err != nil {
					log.Printf("Command failed to start: %v", err)
					response = ServerResponse{
						Status:  "error",
						Message: err.Error(),
					}
				} else {
					response = ServerResponse{
						Status:  "success",
						Command: msg.Command,
						Output:  "Command launched successfully.",
					}
				}
			} else {
				// Command not allowed
				log.Printf("Error: Command '%s' not allowed.", msg.Command)
				response = ServerResponse{
					Status:  "error",
					Message: "Command '" + msg.Command + "' is not in the allowed list.",
				}
			}

			// Send the response back to the writer goroutine
			messages <- response
		}
	}

	// cleanup when connection loop ends
	close(quitPlayerPoll)
	close(messages)
}

// getPlayerInfo runs playerctl to get the current track, status, and artwork. Returns info string, artwork URL, and error.
func getPlayerInfo() (string, string, error) {
	// Try to get metadata (artist - title)
	metaCmd := exec.Command("playerctl", "metadata", "--format", "{{artist}} - {{title}}")
	outMeta, errMeta := metaCmd.Output()
	meta := strings.TrimSpace(string(outMeta))

	// Try to get status (Playing/Paused)
	statusCmd := exec.Command("playerctl", "status")
	outStatus, errStatus := statusCmd.Output()
	status := strings.TrimSpace(string(outStatus))

	// Try to get artwork URL
	artworkCmd := exec.Command("playerctl", "metadata", "mpris:artUrl")
	outArtwork, _ := artworkCmd.Output()
	artwork := strings.TrimSpace(string(outArtwork))

	if errMeta != nil && errStatus != nil {
		// Return combined error message if both fail
		return "playerctl not available or no player running", "", errMeta
	}

	// Build info string
	if meta == "" && status == "" {
		return "No player running", "", nil
	}

	info := meta
	if info == "" {
		info = "(unknown track)"
	}
	if status != "" {
		info = info + " — " + status
	}
	return info, artwork, nil
}

func main() {
	// Handle WebSocket requests at "/ws" with our wsHandler function
	http.HandleFunc("/ws", wsHandler)

	// Start the server on port 8765, listening on all network interfaces
	log.Println("⚡ Blitz server starting at ws://0.0.0.0:8765/ws")
	err := http.ListenAndServe("0.0.0.0:8765", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

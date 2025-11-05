package main

import (
	"Blitz/integration/utils"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// AllowedCommands --- This is where you define your allowed commands ---
// This is a CRITICAL security step.
var AllowedCommands = map[string][]string{
	"update":        {"sudo", "pacman", "-Syu"},
	"list_home":     {"ls", "-l", "/home/swap/"},
	"status":        {"git", "status"},
	"open_firefox":  {"firefox", "--new-window"},
	"open_edge":     {"microsoft-edge-beta"},
	"open_vscode":   {"code-insiders"},
	"open_postman":  {"postman"},
	"player_play":   {"playerctl", "play"},
	"player_pause":  {"playerctl", "pause"},
	"player_next":   {"playerctl", "next"},
	"player_prev":   {"playerctl", "previous"},
	"player_toggle": {"playerctl", "play-pause"},
}

// This upgrades our HTTP connection to a WebSocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ClientMessage struct {
	Command string `json:"command"`
}

type ServerResponse struct {
	Status  string           `json:"status"`
	Command string           `json:"command,omitempty"`
	Output  interface{}      `json:"output,omitempty"`
	Message string           `json:"message,omitempty"`
	Artwork string           `json:"artwork,omitempty"`
	Player  *utils.MediaInfo `json:"player,omitempty"`
}

// This function handles each client connection
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Failed to close connection:", err)
		}
	}(conn)
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
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				info, _ := utils.GetPlayerInfo()
				artwork, _ := utils.HandleArtworkRequest(info.Artwork)
				messages <- ServerResponse{Status: "player", Player: &info, Artwork: artwork}
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
				info, _ := utils.GetPlayerInfo()
				artwork, _ := utils.HandleArtworkRequest(info.Artwork)
				messages <- ServerResponse{Status: "player", Player: &info, Artwork: artwork}
				continue
			}

			if msg.Command == "play" || msg.Command == "pause" || msg.Command == "play-pause" || msg.Command == "next" || msg.Command == "previous" {
				output, err := playerCtrl(msg.Command)
				if err != nil {
					log.Printf("playerctl command failed: %v", err)
					response = ServerResponse{
						Status:  "error",
						Message: err.Error(),
					}
				} else {
					response = ServerResponse{
						Status:  "success",
						Command: msg.Command,
						Output:  output,
					}
				}
				messages <- response
				continue
			}

			// Check if the command is in our allow-list
			if commandToRun, ok := AllowedCommands[msg.Command]; ok {
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

func playerCtrl(command string) (string, error) {
	if command != "play" && command != "pause" && command != "play-pause" && command != "next" && command != "previous" {
		return "", nil
	}
	cmd := exec.Command("playerctl", command)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func main() {
	// Handle WebSocket requests at "/ws" with our wsHandler function
	http.HandleFunc("/ws", wsHandler)

	// Serve static files (CSS, JS, etc.) from the web directory
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	// Start the server on port 8765, listening on all network interfaces
	log.Println("⚡ Blitz server starting at ws://0.0.0.0:8765/ws")
	err := http.ListenAndServe("0.0.0.0:8765", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

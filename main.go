package main

import (
	"Blitz/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
	Status    string                   `json:"status"`
	Command   string                   `json:"command,omitempty"`
	Output    interface{}              `json:"output,omitempty"`
	Message   string                   `json:"message,omitempty"`
	Artwork   string                   `json:"artwork,omitempty"`
	Player    *utils.MediaInfo         `json:"player,omitempty"`
	Bluetooth *[]utils.BluetoothDevice `json:"bluetooth,omitempty"`
	WiFi      *utils.WiFiInfo          `json:"wifi,omitempty"`
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
		utils.Poller(1*time.Second, quitPlayerPoll, func() {
			info, _ := utils.GetPlayerInfo()

			// Use artwork cache to avoid repeated disk reads
			artworkData, _ := utils.HandleArtworkRequest(info.Artwork)
			info.Artwork = artworkData
			messages <- ServerResponse{Status: "player", Output: info, Artwork: artworkData}
		})
	}()

	// Bluetooth poller: periodically checks Bluetooth devices
	quitBluetoothPoll := make(chan struct{})
	go func() {
		utils.Poller(5*time.Second, quitBluetoothPoll, func() {
			devices, _ := utils.GetBluetoothDevices()
			messages <- ServerResponse{Status: "bluetooth", Bluetooth: &devices}
		})
	}()

	// WiFi poller: periodically checks WiFi status and speed
	quitWiFiPoll := make(chan struct{})
	go func() {
		utils.Poller(3*time.Second, quitWiFiPoll, func() {
			wifiInfo, _ := utils.GetWiFiInfo()
			if wifiInfo != nil {
				messages <- ServerResponse{Status: "wifi", WiFi: wifiInfo}
			}
		})
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

			// Special-case: if client requests "bluetooth_info", return current devices immediately
			if msg.Command == "bluetooth_info" {
				devices, _ := utils.GetBluetoothDevices()
				messages <- ServerResponse{Status: "bluetooth", Output: devices}
				continue
			}

			// Handle bluetooth_info request
			if msg.Command == "bluetooth_info" {
				devices, _ := utils.GetBluetoothDevices()
				messages <- ServerResponse{Status: "bluetooth", Bluetooth: &devices}
				continue
			}

			// Handle wifi_info request
			if msg.Command == "wifi_info" {
				wifiInfo, _ := utils.GetWiFiInfo()
				messages <- ServerResponse{Status: "wifi", WiFi: wifiInfo}
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
	close(quitBluetoothPoll)
	close(quitWiFiPoll)
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

	// Recreate upload directory path in user's Downloads
	homeDir, err := os.UserHomeDir()
	uploadDir := ""
	if err == nil {
		uploadDir = filepath.Join(homeDir, "Downloads", "blitz")
	} else {
		// fallback to a local uploads folder
		uploadDir = "./uploads"
	}

	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Warning: failed to create upload dir %s: %v", uploadDir, err)
	}

	// Upload endpoint
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		handleFileUpload(w, r, uploadDir)
	})

	// Serve static files (CSS, JS, etc.) from the web directory
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	// Start the server on port 8765, listening on all network interfaces
	log.Println("⚡ Blitz server starting at ws://0.0.0.0:8765/ws")
	err = http.ListenAndServe("0.0.0.0:8765", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// handleFileUpload handles multipart/form-data file uploads and saves files to uploadDir.
func handleFileUpload(w http.ResponseWriter, r *http.Request, uploadDir string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit memory used while parsing form (100 MB)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	timestamp := time.Now().Unix()
	dstPath := filepath.Join(uploadDir, fmt.Sprintf("%d_%s", timestamp, filename))

	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "failed to create destination file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"status":   "success",
		"filename": filename,
		"path":     dstPath,
		"size":     header.Size,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

package websocket

import (
	"Quazaar/models"
	"Quazaar/utils/player"
	"fmt"
	"log"
	"net/http"
	"time"
)

func Handle(res http.ResponseWriter, req *http.Request) {
	conn, err := CreateWebSocketConnection(res, req)
	if err != nil {
		http.Error(res, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Create unique client with unbuffered channel (fresh messages only)
	client := &Client{
		Conn: conn,
		Send: make(chan models.ServerResponse), // Unbuffered - fresh messages only
		ID:   fmt.Sprintf("%s-%d", req.RemoteAddr, time.Now().UnixNano()),
	}

	// Register client
	RegisterClient(client)
	defer UnregisterClient(client.ID)

	// No read deadline - connection stays open indefinitely
	// Clients won't timeout due to inactivity

	// Send welcome message
	msg := models.ServerResponse{
		Message: "Welcome to the WebSocket server!",
	}
	if err := SendWebSocketMessage(msg, conn); err != nil {
		log.Printf("Failed to send welcome message to %s", client.ID)
		return
	}

	// Writer goroutine - sends messages to this specific client
	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		for response := range client.Send {
			if err := conn.WriteJSON(response); err != nil {
				log.Printf("Error writing to client %s: %v", client.ID, err)
				// Stop reading on write error
				return
			}
		}
	}()

	// Reader goroutine - receives messages from client
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("Client %s disconnected: %v", client.ID, err)
			break
		}

		log.Printf("📨 Received from %s: %+v", client.ID, msg)

		// Handle player commands
		if command, ok := msg["command"]; ok {
			log.Printf("🎮 Processing command: %v", command)
			if err := player.HandlePlayerCommand(msg); err != nil {
				log.Printf("⚠️  Command failed: %v", err)
				// Send error response to client
				errorMsg := models.ServerResponse{
					Status:  "error",
					Message: "command_failed",
					Data: map[string]string{
						"error": err.Error(),
					},
				}
				conn.WriteJSON(errorMsg)
			} else {
				// Send success response to client
				successMsg := models.ServerResponse{
					Status:  "success",
					Message: "command_executed",
					Data: map[string]string{
						"command": fmt.Sprintf("%v", command),
					},
				}
				log.Printf("✅ Command executed successfully: %v", command)
				conn.WriteJSON(successMsg)
			}
		}
	}
	<-writerDone
}

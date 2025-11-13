package websocket

import (
	"Blitz/models"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// HandlePingPong handles ping/pong command from WebSocket client
func HandlePingPong(conn *websocket.Conn, msg map[string]interface{}) {
	command, ok := msg["command"].(string)
	if !ok {
		return
	}

	if command == "ping" {
		SendPong(conn)
	}
}

// SendPong sends pong response to client
func SendPong(conn *websocket.Conn) {
	response := models.ServerResponse{
		Status:  "success",
		Message: "pong",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"server":    "Blitz WebSocket",
		},
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Printf("❌ Failed to send pong: %v", err)
	} else {
		log.Println("🏓 Pong sent")
	}
}

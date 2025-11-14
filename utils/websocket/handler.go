package websocket

import (
	"Blitz/models"
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

	// Create unique client
	client := &Client{
		Conn: conn,
		Send: make(chan models.ServerResponse, 100),
		ID:   fmt.Sprintf("%s-%d", req.RemoteAddr, time.Now().UnixNano()),
	}

	// Register client
	RegisterClient(client)
	defer UnregisterClient(client.ID)

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
				continue
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

		// Handle ping/pong
		HandlePingPong(conn, msg)
	}
	<-writerDone
}

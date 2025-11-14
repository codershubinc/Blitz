package websocket

import (
	"Blitz/models"
	"log"
	"net/http"
)

func Handle(res http.ResponseWriter, req *http.Request) {
	conn, err := CreateWebSocketConnection(res, req)
	if err != nil {
		http.Error(res, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	msg := models.ServerResponse{
		Message: "Welcome to the WebSocket server!",
	}
	if err := SendWebSocketMessage(msg, conn); err != nil {
		http.Error(res, "Failed to send welcome message", http.StatusInternalServerError)
		return
	}

	chh := GetChannel()
	if chh == nil {
		http.Error(res, "Failed to get response channel", http.StatusInternalServerError)
		return
	}

	// Writer goroutine - sends messages to client
	go func() {
		for response := range chh {
			if err := conn.WriteJSON(response); err != nil {
				continue
			}
		}
	}()

	// Reader goroutine - receives messages from client
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			continue
		}
		log.Printf("Received message: %+v\n", msg)

		// Handle ping/pong
		HandlePingPong(conn, msg)
	}
}

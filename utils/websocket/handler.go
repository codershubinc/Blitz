package websocket

import (
	"Blitz/models"
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
	if err := SendWebSocketMessage(msg); err != nil {
		http.Error(res, "Failed to send welcome message", http.StatusInternalServerError)
		return
	}

	chh := GetChannel()
	if chh == nil {
		http.Error(res, "Failed to get response channel", http.StatusInternalServerError)
		return
	}
 

	// Reader goroutine - receives messages from client
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		// Handle ping/pong
		HandlePingPong(conn, msg)
	}
}

package websocket

import (
	"Blitz/models"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}

func CreateWebSocketConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return nil, err
	}
	log.Println("WebSocket Connection established for :=", conn.LocalAddr())
	return conn, nil
}

func CloseWebSocketConnection(conn *websocket.Conn) {
	if conn != nil {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing WebSocket Connection:", err)
		} else {
			log.Println("WebSocket Connection closed")
		}
	} else {
		log.Println("WebSocket Connection is nil, nothing to close")
	}
}

func SendWebSocketMessage(msg models.ServerResponse, conn *websocket.Conn) error {
	if conn == nil {
		log.Println("WebSocket Connection is nil, cannot send message")
		return nil
	}

	err := conn.WriteJSON(msg)
	if err != nil {
		log.Println("Error sending message over WebSocket:", err)
		return err
	}
	log.Println("Message sent over WebSocket:", msg)
	return nil
}

func IsWebSocketConnected(conn *websocket.Conn) bool {
	if conn == nil {
		log.Println("WebSocket Connection is nil")
		return false
	}
	log.Println("WebSocket Connection is active")
	return true
}

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
var Conn *websocket.Conn

func CreateWebSocketConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	Conn = conn
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return nil, err
	}
	log.Println("WebSocket Connection established for :=", Conn.LocalAddr())
	return Conn, nil
}

func CloseWebSocketConnection() {
	if Conn != nil {
		err := Conn.Close()
		if err != nil {
			log.Println("Error closing WebSocket Connection:", err)
		} else {
			log.Println("WebSocket Connection closed")
		}
	} else {
		log.Println("WebSocket Connection is nil, nothing to close")
	}
}

func SendWebSocketMessage(msg models.ServerResponse) error {
	if Conn == nil {
		log.Println("WebSocket Connection is nil, cannot send message")
		return nil
	}

	err := Conn.WriteJSON(msg)
	if err != nil {
		log.Println("Error sending message over WebSocket:", err)
		return err
	}
	log.Println("Message sent over WebSocket:", msg)
	return nil
}

func IsWebSocketConnected() bool {
	if Conn == nil {
		log.Println("WebSocket Connection is nil")
		return false
	}
	log.Println("WebSocket Connection is active")
	return true
}

package websocket

import (
	"Blitz/models"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	Send chan models.ServerResponse
	ID   string
}

var (
	clients   = make(map[string]*Client)
	clientsMu sync.RWMutex
)

// RegisterClient adds a new client to the broadcast list
func RegisterClient(client *Client) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	clients[client.ID] = client
	log.Printf("✅ Client registered: %s (Total clients: %d)", client.ID, len(clients))
}

// UnregisterClient removes a client from the broadcast list
func UnregisterClient(clientID string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if client, ok := clients[clientID]; ok {
		close(client.Send)
		delete(clients, clientID)
		log.Printf("❌ Client unregistered: %s (Total clients: %d)", clientID, len(clients))
	}
}

// BroadcastMessage sends a message to all connected clients
func BroadcastMessage(msg models.ServerResponse) {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	if len(clients) == 0 {
		log.Println("⚠️  No clients connected, message not sent")
		return
	}

	for _, client := range clients {
		select {
		case client.Send <- msg:
			// Message sent successfully
		default:
			log.Printf("⚠️  Client %s channel full, skipping message", client.ID)
		}
	}
	log.Printf("📡 Broadcast to %d clients", len(clients))
}

// GetClientCount returns the number of connected clients
func GetClientCount() int {
	clientsMu.RLock()
	defer clientsMu.RUnlock()
	return len(clients)
}

// Legacy functions for backward compatibility
func CreateChannel() chan models.ServerResponse {
	log.Println("⚠️  CreateChannel is deprecated, using broadcast pattern")
	return make(chan models.ServerResponse, 100)
}

func GetChannel() chan models.ServerResponse {
	log.Println("⚠️  GetChannel is deprecated, using broadcast pattern")
	return make(chan models.ServerResponse, 100)
}

func CloseChannel() {
	log.Println("⚠️  CloseChannel is deprecated, using broadcast pattern")
}

func WriteChannelMessage(msg models.ServerResponse) {
	BroadcastMessage(msg)
}

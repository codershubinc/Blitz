package main

import (
	"Blitz/utils/websocket"
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Hello Blitz")

	// Setup HTTP routes
	http.HandleFunc("/ws", websocket.Handle)
	http.HandleFunc("/", serveHome)

	// Start the server (this blocks forever)
	fmt.Println("Starting server on http://0.0.0.0:8765")
	fmt.Println("WebSocket endpoint: ws://localhost:8765/ws")
	fmt.Println("Press Ctrl+C to stop the server")

	if err := http.ListenAndServe("0.0.0.0:8765", nil); err != nil {
		log.Fatal("Server error:", err)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "web/index.html")
}

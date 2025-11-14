package main

import (
	"Quazaar/utils/poller"
	"Quazaar/utils/websocket"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	fmt.Println("Hello Quazaar Server ...")

	// Setup HTTP routes
	http.HandleFunc("/ws", websocket.Handle)
	http.HandleFunc("/", serveHome)

	go poller.Handle()

	// Start the server (this blocks forever)
	fmt.Println("Starting server on http://0.0.0.0:8765")
	fmt.Println("WebSocket endpoint: ws://localhost:8765/ws")
	fmt.Println("Press Ctrl+C to stop the server")
	localAddr := os.Getenv("LOCAL_HOST_IP") + ":" + os.Getenv("LOCAL_HOST_PORT")
	fmt.Println("lo", localAddr)

	if err := http.ListenAndServe(localAddr, nil); err != nil {
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
	http.ServeFile(w, r, "temp/web/index.html")
}

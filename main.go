package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ws/spicy-dice", diceGameHandler)

	log.Println("Starting WebSocket server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

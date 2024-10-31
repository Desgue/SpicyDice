package main

import (
	"log"
	"net/http"
)

const (
	PORT string = ":8080"
)

func main() {
	http.HandleFunc("/ws/spicy-dice", diceGameHandler)

	log.Printf("Starting WebSocket server on %s", PORT)
	err := http.ListenAndServe(PORT, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

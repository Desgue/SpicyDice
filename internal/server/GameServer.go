package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Desgue/SpicyDice/internal/config"
	"github.com/Desgue/SpicyDice/internal/domain"
	"github.com/Desgue/SpicyDice/internal/service"
	"github.com/gorilla/websocket"
)

const (
	tickDuration    time.Duration = 30 * time.Second
	readBufferSize  int           = 1024
	writeBufferSize int           = 1024
	writeTimeout    time.Duration = 10 * time.Second
	readTimeout     time.Duration = 60 * time.Second
)

type WsMessage struct {
	Type    domain.MessageType `json:"type"`
	Payload json.RawMessage    `json:"payload"`
}

type WebSocketServer struct {
	service  *service.GameService
	upgrader websocket.Upgrader
}

func NewWebSocketServer(service *service.GameService) *WebSocketServer {
	return &WebSocketServer{
		service: service,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  readBufferSize,
			WriteBufferSize: writeBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}
func (s *WebSocketServer) Run() {
	http.HandleFunc("/ws/spicy-dice", s.Serve)
	port := config.New().Server.Port
	log.Printf("Starting WebSocket server on port :%s", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
func (s *WebSocketServer) Serve(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	conn := connection{
		service:      s.service,
		ws:           ws,
		messagesChan: make(chan WsMessage, 100),
		doneChan:     make(chan struct{}),
	}

	go conn.readPump()
	go conn.writePump()

}

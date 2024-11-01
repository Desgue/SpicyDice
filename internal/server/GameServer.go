package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Desgue/SpicyDice/internal/appErrors"
	"github.com/Desgue/SpicyDice/internal/domain"
	"github.com/Desgue/SpicyDice/internal/service"
	"github.com/gorilla/websocket"
)

const (
	PORT string = ":8080"
)

type WebSocketServer struct {
	service  *service.GameService
	upgrader websocket.Upgrader
}

func NewWebSocketServer(service *service.GameService) *WebSocketServer {
	return &WebSocketServer{
		service: service,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}
func (s *WebSocketServer) Run() {
	http.HandleFunc("/ws/spicy-dice", s.Serve)

	log.Printf("Starting WebSocket server on %s", PORT)
	err := http.ListenAndServe(PORT, nil)
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

	// Ensure connection is closed when handler returns
	defer func() {
		if err := ws.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	s.handleConnection(ws)
}

func (s *WebSocketServer) handleConnection(ws *websocket.Conn) {
	for {
		var message domain.WsMessage
		err := ws.ReadJSON(&message)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		if err = s.handleMessage(ws, message); err != nil {
			log.Printf("Error handling message type '%s': %v", message.Type, err)

			var gameErr *appErrors.GameError
			if !errors.As(err, &gameErr) {
				gameErr = appErrors.NewInternalError(err.Error())
			}

			if err := ws.WriteJSON(gameErr); err != nil {
				log.Printf("Error sending error response: %v", err)
				return
			}
		}
	}
}
func (s *WebSocketServer) handleMessage(ws *websocket.Conn, msg domain.WsMessage) error {
	switch msg.Type {
	case domain.MessageTypeWallet:
		return s.handleWalletMessage(ws, msg)
	case domain.MessageTypePlay:
		return s.handlePlayMessage(ws, msg)
	case domain.MessageTypeEndPlay:
		return s.handleEndPlayMessage(ws, msg)
	default:
		return appErrors.NewInvalidInputError(fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}
func (s *WebSocketServer) handleWalletMessage(ws *websocket.Conn, msg domain.WsMessage) error {
	var payload domain.WalletPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return appErrors.NewInvalidInputError("Invalid wallet payload")
	}

	log.Printf("Handling Wallet Message for User ID: %d", payload.ClientID)

	balance, err := s.service.GetBalance(payload.ClientID)
	if err != nil {
		return err
	}

	return s.sendResponse(ws, domain.MessageTypeWallet, balance)
}

// handlePlayMessage processes play-related messages
func (s *WebSocketServer) handlePlayMessage(ws *websocket.Conn, msg domain.WsMessage) error {
	var payload domain.PlayPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return appErrors.NewInvalidInputError("Invalid play payload")
	}

	log.Printf("Handling Play Message for User ID: %d", payload.ClientID)

	result, err := s.service.ProcessPlay(payload)
	if err != nil {
		return err
	}

	return s.sendResponse(ws, domain.MessageTypePlay, result)
}

// handleEndPlayMessage processes end-play messages
func (s *WebSocketServer) handleEndPlayMessage(ws *websocket.Conn, msg domain.WsMessage) error {
	var payload domain.EndPlayPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return appErrors.NewInvalidInputError("Invalid end-play payload")
	}

	log.Printf("Handling End Play Message for User ID: %d", payload.ClientID)

	endPlayResponse, err := s.service.EndPlay(payload.ClientID)
	if err != nil {
		return err
	}

	return s.sendResponse(ws, domain.MessageTypeEndPlay, endPlayResponse)
}

// sendResponse sends a response back through the WebSocket connection
func (s *WebSocketServer) sendResponse(ws *websocket.Conn, msgType domain.MessageType, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}

	return ws.WriteJSON(domain.WsMessage{
		Type:    msgType,
		Payload: payload,
	})
}

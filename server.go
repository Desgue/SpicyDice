package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	service  *GameService
	upgrader websocket.Upgrader
}

func NewWebSocketServer(service *GameService) *WebSocketServer {
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
		var message WsMessage
		err := ws.ReadJSON(&message)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		if err = s.handleMessage(ws, message); err != nil {
			log.Printf("Error handling message type '%s': %v", message.Type, err)

			var gameErr *GameError
			if !errors.As(err, &gameErr) {
				gameErr = NewInternalError(err.Error())
			}

			if err := ws.WriteJSON(gameErr); err != nil {
				log.Printf("Error sending error response: %v", err)
				return
			}
		}
	}
}
func (s *WebSocketServer) handleMessage(ws *websocket.Conn, msg WsMessage) error {
	switch msg.Type {
	case MessageTypeWallet:
		return s.handleWalletMessage(ws, msg)
	case MessageTypePlay:
		return s.handlePlayMessage(ws, msg)
	case MessageTypeEndPlay:
		return s.handleEndPlayMessage(ws, msg)
	default:
		return NewInvalidInputError(fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}
func (s *WebSocketServer) handleWalletMessage(ws *websocket.Conn, msg WsMessage) error {
	var payload WalletPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return NewInvalidInputError("Invalid wallet payload")
	}

	log.Printf("Handling Wallet Message for User ID: %d", payload.ClientID)

	balance, err := s.service.GetBalance(payload.ClientID)
	if err != nil {
		return err
	}

	return s.sendResponse(ws, MessageTypeWallet, balance)
}

// handlePlayMessage processes play-related messages
func (s *WebSocketServer) handlePlayMessage(ws *websocket.Conn, msg WsMessage) error {
	var payload PlayPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return NewInvalidInputError("Invalid play payload")
	}

	log.Printf("Handling Play Message for User ID: %d", payload.ClientID)

	result, err := s.service.ProcessPlay(payload)
	if err != nil {
		return err
	}

	return s.sendResponse(ws, MessageTypePlay, result)
}

// handleEndPlayMessage processes end-play messages
func (s *WebSocketServer) handleEndPlayMessage(ws *websocket.Conn, msg WsMessage) error {
	var payload EndPlayPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return NewInvalidInputError("Invalid end-play payload")
	}

	log.Printf("Handling End Play Message for User ID: %d", payload.ClientID)

	if err := s.service.EndPlay(payload.ClientID); err != nil {
		return err
	}

	return s.sendResponse(ws, MessageTypeEndPlay, true)
}

// sendResponse sends a response back through the WebSocket connection
func (s *WebSocketServer) sendResponse(ws *websocket.Conn, msgType MessageType, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}

	return ws.WriteJSON(WsMessage{
		Type:    msgType,
		Payload: payload,
	})
}

/* func diceGameHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer func() {
		if err := ws.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()
	for {
		var message WsMessage
		err := ws.ReadJSON(&message)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}
		if err = handleMessageType(ws, message); err != nil {
			log.Printf("Error handling message type '%s': %v", message.Type, err)
			var gameErr *GameError
			if !errors.As(err, &gameErr) {
				gameErr = NewInternalError(err.Error())
			}
			if err := ws.WriteJSON(gameErr); err != nil {
				log.Printf("Error sending error response: %v", err)
				return
			}
		}

	}
}

func handleMessageType(ws *websocket.Conn, message WsMessage) error {
	switch message.Type {
	case "wallet":
		return handleWalletMessage(ws, message)
	case "play":
		return handlePlayMessage(ws, message)
	case "endplay":
		return handleEndPlayMessage(ws, message)
	default:
		return NewInvalidInputError(fmt.Sprintf("Unknown message type: %s", message.Type))
	}
}

func handleWalletMessage(ws *websocket.Conn, msg WsMessage) error {
	var walletPayload WalletPayload
	if err := json.Unmarshal(msg.Payload, &walletPayload); err != nil {
		return err
	}
	log.Printf("Handling Wallet Message for User of Id -> %d", walletPayload.ClientID)
	balance, err := getBalance(walletPayload.ClientID)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(balance)
	if err != nil {
		return err
	}
	ws.WriteJSON(WsMessage{Type: Wallet, Payload: payload})
	return nil
}

func handlePlayMessage(ws *websocket.Conn, msg WsMessage) error {
	var playPayload PlayPayload
	if err := json.Unmarshal(msg.Payload, &playPayload); err != nil {
		return err
	}
	log.Printf("Handling Play Message for User of Id -> %d", playPayload.ClientID)
	playRes, err := processPlay(playPayload)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(playRes)
	if err != nil {
		return err
	}
	ws.WriteJSON(WsMessage{Type: Play, Payload: payload})
	return nil
}
func handleEndPlayMessage(ws *websocket.Conn, msg WsMessage) error {
	var endPlayPayload EndPlayPayload
	if err := json.Unmarshal(msg.Payload, &endPlayPayload); err != nil {
		return err
	}
	log.Printf("Handling End Play Message for User of Id -> %d", endPlayPayload.ClientID)
	if err := endPlay(endPlayPayload.ClientID); err != nil {
		return err
	}

	payload, err := json.Marshal(true)
	if err != nil {
		return err
	}
	ws.WriteJSON(WsMessage{Type: EndPlay, Payload: payload})
	return nil
}
*/

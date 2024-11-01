package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "./frontend/index.html")
}

func diceGameHandler(w http.ResponseWriter, r *http.Request) {
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
	ws.WriteJSON(WsMessage{Type: "wallet", Payload: payload})
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
	ws.WriteJSON(WsMessage{Type: "play", Payload: payload})
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
	ws.WriteJSON(WsMessage{Type: "endplay", Payload: payload})
	return nil
}

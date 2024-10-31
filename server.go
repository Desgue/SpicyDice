package main

import (
	"encoding/json"
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

func diceGameHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer ws.Close()
	for {
		var message WsMessage
		err := ws.ReadJSON(&message)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		messageType := message.Type
		switch messageType {
		case "wallet":
			var walletPayload WalletPayload
			if err := json.Unmarshal(message.Payload, &walletPayload); err != nil {
				log.Printf("Error parsing wallet payload: %v", err)
				continue
			}
			balance, err := getBalance(walletPayload.ClientID)
			if err != nil {
				log.Println(err)
				return
			}
			ws.WriteJSON(balance)
		case "play":
			var playPayload PlayPayload
			if err := json.Unmarshal(message.Payload, &playPayload); err != nil {
				log.Printf("Error parsing play payload: %v", err)
				continue
			}
			playRes, err := processPlay(playPayload)
			if err != nil {
				log.Println(err)
				return
			}
			ws.WriteJSON(playRes)
		case "endplay":
			var endPlayPayload EndPlayPayload
			if err := json.Unmarshal(message.Payload, &endPlayPayload); err != nil {
				log.Printf("Error parsing end play payload: %v", err)
				continue
			}
			// Finalize the play
			err = endPlay(endPlayPayload.ClientID)
			if err != nil {
				log.Println(err)
				return
			}

			// Respond with the endplay result
			response := map[string]interface{}{
				"type":    "endplay",
				"success": true,
			}
			ws.WriteJSON(response)
		default:
			log.Printf("Unknown message type: %s", message.Type)

		}
	}

}

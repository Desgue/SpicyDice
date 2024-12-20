package server

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Desgue/SpicyDice/internal/appErrors"
	"github.com/Desgue/SpicyDice/internal/domain"
	"github.com/Desgue/SpicyDice/internal/service"
	"github.com/gorilla/websocket"
)

// connection implements concurrent safe bidirectional communication
// using separate read/write goroutines with proper cleanup mechanisms
type connection struct {
	service      *service.GameService
	ws           *websocket.Conn
	mu           sync.Mutex
	messagesChan chan (WsMessage)
	doneChan     chan (struct{})
	closeOnce    sync.Once
}

// readPump maintains the read side of the websocket, implementing ping/pong heartbeat
// and message routing. Exits on any error, triggering connection cleanup
func (c *connection) readPump() {
	defer func() {
		log.Println("Closing connection from readPump routine...")
		c.cleanUpOnce()
	}()

	c.ws.SetPongHandler(func(string) error {
		return c.ws.SetReadDeadline(time.Now().Add(readTimeout))
	})

	for {
		c.ws.SetReadDeadline(time.Now().Add(readTimeout))
		var message WsMessage
		err := c.ws.ReadJSON(&message)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		if err = c.handleMessage(message); err != nil {
			log.Printf("Error handling message type '%s': %v", message.Type, err)
			c.writeToChan(domain.MessageTypeError, err)
		}
	}
}

// writePump handles the write side of the websocket with periodic ping messages
// Uses mutex for concurrent write safety and implements graceful shutdown
func (c *connection) writePump() {
	ticker := time.NewTicker(tickDuration)

	defer func() {
		log.Println("Closing connection from writePump routine...")
		ticker.Stop()
		c.cleanUpOnce()
	}()

	c.ws.SetPingHandler(func(string) error {
		c.mu.Lock()
		defer c.mu.Unlock()
		return c.ws.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(time.Second))
	})
	for {

		select {
		case <-ticker.C:
			c.mu.Lock()
			c.ws.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.mu.Unlock()
				log.Printf("Error sending ping: %v", err)
				return
			}
			c.mu.Unlock()

		case message := <-c.messagesChan:
			c.mu.Lock()
			c.ws.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.ws.WriteJSON(message); err != nil {
				log.Printf("error writing json message: %s", err)
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		case <-c.doneChan:
			return
		}

	}

}

// handleMessage routes incoming messages to their appropriate handlers based on message type
// Returns domain specific errors for invalid messages or processing failures
func (c *connection) handleMessage(msg WsMessage) error {
	switch msg.Type {
	case domain.MessageTypeWallet:
		return c.handleWalletMessage(msg)
	case domain.MessageTypePlay:
		return c.handlePlayMessage(msg)
	case domain.MessageTypeEndPlay:
		return c.handleEndPlayMessage(msg)
	default:
		return appErrors.NewInvalidInputError(fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}

// handleWalletMessage processes wallet related requests ensuring payload validity
func (c *connection) handleWalletMessage(msg WsMessage) error {
	var payload domain.WalletRequest
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return appErrors.NewInvalidInputError("Invalid wallet payload")
	}

	log.Printf("Handling Wallet Message for User ID: %d", payload.ClientID)

	balance, err := c.service.GetBalance(payload.ClientID)
	if err != nil {
		return err
	}
	return c.writeToChan(domain.MessageTypeWallet, balance)
}

// handlePlayMessage processes game play requests ensuring payload validity
func (c *connection) handlePlayMessage(msg WsMessage) error {
	var payload domain.PlayRequest
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return appErrors.NewInvalidInputError("Invalid play payload")
	}

	log.Printf("Handling Play Message for User ID: %d", payload.ClientID)

	result, err := c.service.ProcessPlay(payload, service.Dice{Sides: 6})
	if err != nil {
		return err
	}

	return c.writeToChan(domain.MessageTypePlay, result)
}

// handleEndPlayMessage processes session end requests ensuring payload validity
func (c *connection) handleEndPlayMessage(msg WsMessage) error {
	var payload domain.EndPlayRequest
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return appErrors.NewInvalidInputError("Invalid end-play payload")
	}

	log.Printf("Handling End Play Message for User ID: %d", payload.ClientID)

	endPlayResponse, err := c.service.EndPlay(payload.ClientID)
	if err != nil {
		return err
	}

	return c.writeToChan(domain.MessageTypeEndPlay, endPlayResponse)
}

// Returns error if connection is closed or message buffer is full
func (c *connection) writeToChan(msgType domain.MessageType, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling incomming message: %w", err)
	}
	select {
	case c.messagesChan <- WsMessage{Type: msgType, Payload: payload}:
		return nil
	case <-c.doneChan:
		return fmt.Errorf("connection closed")
	default:
		return fmt.Errorf("message channel full")
	}
}

// cleanUpOnce ensures connection cleanup happens only once
// Uses sync.Once to prevent duplicate cleanup operations
func (c *connection) cleanUpOnce() {
	c.closeOnce.Do(func() {
		log.Println("Closing connection...")
		close(c.doneChan)
		if err := c.ws.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
		close(c.messagesChan)
	})
}

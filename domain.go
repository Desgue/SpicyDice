package main

import (
	"encoding/json"
	"time"
)

type Repository interface {
	GetBalance(playerID int) (float64, error)
	DeductBalance(playerID int, amount float64) (float64, error)
	IncreaseBalance(playerID int, amount float64) (float64, error)
	CreateGameSession(sess GameSessionRequest) (GameSession, error)
	GetActiveSession(playerID int) (*GameSession, error)
	CloseCurrentGameSession(clientID int) error
}

type MessageType string
type BetType string

func (m MessageType) IsValid() bool {
	switch m {
	case MessageTypeWallet, MessageTypePlay, MessageTypeEndPlay:
		return true
	default:
		return false
	}
}
func (m BetType) IsValid() bool {
	switch m {
	case Odd, Even:
		return true
	default:
		return false
	}
}

const (
	MessageTypeWallet  MessageType = "wallet"
	MessageTypePlay    MessageType = "play"
	MessageTypeEndPlay MessageType = "endplay"
	Even               BetType     = "even"
	Odd                BetType     = "odd"
)

type WsMessage struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type WalletPayload struct {
	ClientID int `json:"client_id"`
}

type WalletResponse struct {
	ClientID int     `json:"client_id"`
	Balance  float64 `json:"balance"`
}

type PlayPayload struct {
	ClientID  int     `json:"client_id"`
	BetAmount float64 `json:"bet_amount"`
	BetType   BetType `json:"bet_type"`
}

type PlayResponse struct {
	DiceResult int     `json:"dice_result"`
	Won        bool    `json:"won"`
	Balance    float64 `json:"balance"`
}
type EndPlayResponse struct {
	ClientID int     `json:"client_id"`
	Balance  float64 `json:"balance"`
}
type EndPlayPayload struct {
	ClientID int `json:"client_id"`
}

type GameSession struct {
	SessionID    int        `json:"session_id"`
	PlayerID     int        `json:"player_id"`
	BetAmount    float64    `json:"bet_amount"`
	DiceResult   int        `json:"dice_result"`
	Won          bool       `json:"won"`
	Active       bool       `json:"active"`
	SessionStart time.Time  `json:"session_start"`
	SessionEnd   *time.Time `json:"session_end,omitempty"` // Nullable
}

type GameSessionRequest struct {
	PlayerID     int       `json:"player_id"`
	BetAmount    float64   `json:"bet_amount"`
	DiceResult   int       `json:"dice_result"`
	Won          bool      `json:"won"`
	Active       bool      `json:"active"`
	SessionStart time.Time `json:"session_start"`
}

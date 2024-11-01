package main

import "encoding/json"

type Repository interface {
	GetBalance(userId int) (*float64, error)
	ProcessPlay()
	EndPlay()
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
	DiceResult int  `json:"dice_result"`
	Won        bool `json:"won"`
}

type EndPlayPayload struct {
	ClientID int `json:"client_id"`
}

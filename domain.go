package main

import "encoding/json"

type MessageType string
type BetType string

const (
	Wallet  MessageType = "wallet"
	Play    MessageType = "play"
	EndPlay MessageType = "endplay"
	Even    BetType     = "even"
	Odd     BetType     = "odd"
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

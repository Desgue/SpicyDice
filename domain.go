package main

import "encoding/json"

type WsMessage struct {
	Type    string          `json:"type"` // wallet | play | endplay
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
	BetType   string  `json:"bet_type"` // "even" or "odd"
}

type PlayResponse struct {
	DiceResult int  `json:"dice_result"`
	Won        bool `json:"won"`
}

type EndPlayPayload struct {
	ClientID int `json:"client_id"`
}

package domain

import (
	"time"
)

type MessageType string
type BetType string

func (m MessageType) IsValid() bool {
	switch m {
	case MessageTypeWallet, MessageTypePlay, MessageTypeEndPlay, MessageTypeError:
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
	MessageTypeError   MessageType = "error"
	MessageTypeWallet  MessageType = "wallet"
	MessageTypePlay    MessageType = "play"
	MessageTypeEndPlay MessageType = "endplay"
	Even               BetType     = "even"
	Odd                BetType     = "odd"
)

type WalletRequest struct {
	ClientID int `json:"client_id"`
}

type WalletResponse struct {
	ClientID int     `json:"client_id"`
	Balance  float64 `json:"balance"`
}

type PlayRequest struct {
	ClientID  int     `json:"client_id"`
	BetAmount float64 `json:"bet_amount"`
	BetType   BetType `json:"bet_type"`
}

type PlayResponse struct {
	DiceResult int     `json:"dice_result"`
	Won        bool    `json:"won"`
	Balance    float64 `json:"balance"`
	BetAmount  float64 `json:"bet_amount"`
}
type EndPlayResponse struct {
	ClientID int `json:"client_id"`
}

type EndPlayRequest struct {
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

type PlayTransaction struct {
	Message    PlayRequest
	DiceResult int
	Won        bool
}

type BalanceUpdate struct {
	PlayerID     int
	ChangeAmount float64
}

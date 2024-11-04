package domain

import (
	"time"
)

// MessageType represents the supported message types for game communication
type MessageType string

// BetType represents the available betting options
type BetType string

// IsValid checks if the message type is among the supported operations
func (m MessageType) IsValid() bool {
	switch m {
	case MessageTypeWallet, MessageTypePlay, MessageTypeEndPlay, MessageTypeError:
		return true
	default:
		return false
	}
}

// IsValid ensures the bet type matches the allowed game rules
func (m BetType) IsValid() bool {
	switch m {
	case Odd, Even:
		return true
	default:
		return false
	}
}

// System-wide constants for message and bet types
const (
	MessageTypeError   MessageType = "error"
	MessageTypeWallet  MessageType = "wallet"
	MessageTypePlay    MessageType = "play"
	MessageTypeEndPlay MessageType = "endplay"
	Even               BetType     = "even"
	Odd                BetType     = "odd"
)

// WalletRequest initiates a balance check operation
type WalletRequest struct {
	ClientID int `json:"client_id"`
}

// WalletResponse carries the current balance state
type WalletResponse struct {
	ClientID int     `json:"client_id"`
	Balance  float64 `json:"balance"`
}

// PlayRequest encapsulates the necessary information to start a game round
type PlayRequest struct {
	ClientID  int     `json:"client_id"`
	BetAmount float64 `json:"bet_amount"`
	BetType   BetType `json:"bet_type"`
}

// PlayResponse contains the game round results and updated balance
type PlayResponse struct {
	DiceResult int     `json:"dice_result"`
	Won        bool    `json:"won"`
	Balance    float64 `json:"balance"`
	BetAmount  float64 `json:"bet_amount"`
}

// EndPlayResponse confirms the termination of a game session
type EndPlayResponse struct {
	ClientID int `json:"client_id"`
}

// EndPlayRequest signals the intention to terminate the current game session
type EndPlayRequest struct {
	ClientID int `json:"client_id"`
}

// GameSession represents the state and metadata of an active or completed game
type GameSession struct {
	SessionID    int        `json:"session_id"`
	PlayerID     int        `json:"player_id"`
	BetAmount    float64    `json:"bet_amount"`
	DiceResult   int        `json:"dice_result"`
	Won          bool       `json:"won"`
	Active       bool       `json:"active"`
	SessionStart time.Time  `json:"session_start"`
	SessionEnd   *time.Time `json:"session_end,omitempty"`
}

// GameSessionRequest contains the required data to initialize a new game session
type GameSessionRequest struct {
	PlayerID     int       `json:"player_id"`
	BetAmount    float64   `json:"bet_amount"`
	DiceResult   int       `json:"dice_result"`
	Won          bool      `json:"won"`
	Active       bool      `json:"active"`
	SessionStart time.Time `json:"session_start"`
}

// PlayTransaction combines the player's bet with the game outcome
type PlayTransaction struct {
	Message    PlayRequest
	DiceResult int
	Won        bool
}

// BalanceUpdate represents a modification to a player's account balance
type BalanceUpdate struct {
	PlayerID     int
	ChangeAmount float64
}

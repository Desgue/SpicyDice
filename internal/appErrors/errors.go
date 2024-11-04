package appErrors

import "fmt"

// Error codes for game-related operations
const (
	InternalErrorCode = iota + 1000
	InvalidInputErrorCode
	InsufficientFundsErrorCode
	InvalidBetAmountErrorCode
	UserNotFoundErrorCode
	ActiveSessionErrorCode
	DiceRollErrorCode
)

// GameError provides structured error information for client feedback
type GameError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error satisfies the error interface and formats the error message
func (e *GameError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Details)
}

// NewInternalError creates errors for unexpected system failures
func NewInternalError(details string) *GameError {
	return &GameError{
		Code:    InternalErrorCode,
		Message: "Internal server error",
		Details: details,
	}
}

// NewInvalidInputError creates errors for malformed request data
func NewInvalidInputError(details string) *GameError {
	return &GameError{
		Code:    InvalidBetAmountErrorCode,
		Message: "Invalid input provided",
		Details: details,
	}
}

// NewInsufficientFundsError creates errors when bet exceeds player balance
func NewInsufficientFundsError(details string) *GameError {
	return &GameError{
		Code:    InsufficientFundsErrorCode,
		Message: "Bet amount exceeds available balance",
		Details: details,
	}
}

// NewInvalidBetAmountError creates errors when bet doesn't meet game rules
func NewInvalidBetAmountError(details string) *GameError {
	return &GameError{
		Code:    InvalidBetAmountErrorCode,
		Message: "Invalid bet amount",
		Details: details,
	}
}

// NewUserNotFoundError creates errors for non-existent player lookups
func NewUserNotFoundError(details string) *GameError {
	return &GameError{
		Code:    UserNotFoundErrorCode,
		Message: "User not found",
		Details: details,
	}
}

// NewActiveSessionError creates errors for concurrent session conflicts
func NewActiveSessionError(details string) *GameError {
	return &GameError{
		Code:    ActiveSessionErrorCode,
		Message: "Active session error",
		Details: details,
	}
}

// NewDiceRollError creates errors for randomization failures
func NewDiceRollError(details string) *GameError {
	return &GameError{
		Code:    DiceRollErrorCode,
		Message: "Error rolling dice",
		Details: details,
	}
}

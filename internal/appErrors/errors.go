package appErrors

import "fmt"

const (
	InternalErrorCode = iota + 1000
	InvalidInputErrorCode
	InsufficientFundsErrorCode
	InvalidBetAmountErrorCode
	UserNotFoundErrorCode
	ActiveSessionErrorCode
)

type GameError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *GameError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Details)
}

func NewInternalError(details string) *GameError {
	return &GameError{
		Code:    InternalErrorCode,
		Message: "Internal server error",
		Details: details,
	}
}
func NewInvalidInputError(details string) *GameError {
	return &GameError{
		Code:    InvalidBetAmountErrorCode,
		Message: "Invalid input provided",
		Details: details,
	}
}

func NewInsufficientFundsError(details string) *GameError {
	return &GameError{
		Code:    InsufficientFundsErrorCode,
		Message: "Bet amount exceeds available balance",
		Details: details,
	}
}

func NewInvalidBetAmountError(details string) *GameError {
	return &GameError{
		Code:    InvalidBetAmountErrorCode,
		Message: "Invalid bet amount",
		Details: details,
	}
}

func NewUserNotFoundError(details string) *GameError {
	return &GameError{
		Code:    UserNotFoundErrorCode,
		Message: "User not found",
		Details: details,
	}
}

func NewActiveSessionError(details string) *GameError {
	return &GameError{
		Code:    ActiveSessionErrorCode,
		Message: "Player already has an active session",
		Details: details,
	}
}

package main

import "fmt"

type GameError struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *GameError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Details)
}

func NewInternalError(details string) *GameError {
	return &GameError{
		Message: "Internal server error",
		Details: details,
	}
}
func NewInvalidInputError(details string) *GameError {
	return &GameError{
		Message: "Invalid input provided",
		Details: details,
	}
}

func NewInsufficientFundsError(details string) *GameError {
	return &GameError{
		Message: "Bet amount exceeds available balance",
		Details: details,
	}
}

func NewInvalidBetAmountError(details string) *GameError {
	return &GameError{
		Message: "Invalid bet amount",
		Details: details,
	}
}

func NewUserNotFoundError(details string) *GameError {
	return &GameError{
		Message: "User not found",
		Details: details,
	}
}

func NewActiveSessionError(details string) *GameError {
	return &GameError{
		Message: "Player already has an active session",
		Details: details,
	}
}

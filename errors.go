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

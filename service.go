package main

import (
	"fmt"
	"log"
	"math/rand"
)

const (
	MinBetAmount = 1.0
	MaxBetAmount = 1000.0
)

var balance = 1000.0

// Placeholder function to get the user balance
func getBalance(userID int) (WalletResponse, error) {
	return WalletResponse{ClientID: userID, Balance: balance}, nil // Replace with actual database call
}

// Placeholder function to process the play
func processPlay(msg PlayPayload) (PlayResponse, error) {
	log.Printf("\nProcessing play for user id -> %d\nBet Amount -> %g\nBet Type -> %s", msg.ClientID, msg.BetAmount, msg.BetType)

	if err := validateBetAmount(msg.BetAmount, balance); err != nil {
		return PlayResponse{}, err
	}

	balance -= msg.BetAmount

	diceSides := 6 // TODO: Implement more then 6 sided dice?
	diceResult := rollDice(diceSides)
	won := outcome(msg.BetType, diceResult)

	return PlayResponse{DiceResult: diceResult, Won: won}, nil
}

// Placeholder function to end the play session
func endPlay(clientId int) error {
	log.Printf("Updating session for client id -> %d", clientId)

	return nil // Replace with actual session finalization logic
}

func validateBetAmount(betAmount, balance float64) error {
	var details string

	if betAmount > balance {
		details = fmt.Sprintf("bet amount %.2f exceeds available balance %.2f", betAmount, balance)
		return NewInsufficientFundsError(details)
	}

	if betAmount < 0 {
		details = fmt.Sprintf("bet amount cannot be negative: %.2f", betAmount)
		return NewInvalidBetAmountError(details)
	}
	if betAmount == 0 {
		details = "bet amount cannot be zero"
		return NewInvalidBetAmountError(details)
	}

	if betAmount < MinBetAmount {
		details = fmt.Sprintf("minimum bet amount is %.2f", MinBetAmount)
		return NewInvalidBetAmountError(details)
	}

	if betAmount > MaxBetAmount {
		details = fmt.Sprintf("maximum bet amount is %.2f", MaxBetAmount)
		return NewInvalidBetAmountError(details)
	}

	return nil
}

func rollDice(sides int) int {
	return rand.Intn(sides) + 1
}

func outcome(betType BetType, diceResult int) bool {
	isEven := diceResult%2 == 0
	return (betType == Even && isEven) || (betType == Odd && !isEven)
}

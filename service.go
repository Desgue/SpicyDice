package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

const (
	MinBetAmount = 1.0
	MaxBetAmount = 1000.0
)

type GameService struct {
	repo Repository
}

func NewGameService(repo Repository) *GameService {
	return &GameService{
		repo: repo,
	}
}

func (gs *GameService) GetBalance(userID int) (WalletResponse, error) {
	balance, err := gs.repo.GetBalance(userID)
	if err != nil {
		return WalletResponse{}, NewInternalError(err.Error())
	}
	return WalletResponse{ClientID: userID, Balance: balance}, nil
}

func (gs *GameService) ProcessPlay(msg PlayPayload) (PlayResponse, error) {
	log.Printf("\nProcessing play for user id -> %d\nBet Amount -> %g\nBet Type -> %s", msg.ClientID, msg.BetAmount, msg.BetType)

	// VALIDATE USER DO NOT HAS AN ACTIVE SESSION BEFORE PLACING A BET
	activeSession, err := gs.repo.GetActiveSession(msg.ClientID)
	if err != nil {
		return PlayResponse{}, NewInternalError(err.Error())
	}
	if activeSession != nil {
		return PlayResponse{}, NewActiveSessionError("Cannot place a bet because the player already has an active session.")
	}

	balance, err := gs.repo.GetBalance(msg.ClientID)
	if err != nil {
		return PlayResponse{}, NewInternalError(err.Error())
	}

	// VALIDATE USER BET AMOUNT
	if err := gs.validateBetAmount(msg.BetAmount, balance); err != nil {
		return PlayResponse{}, err
	}

	// GAME LOGIC
	newBalance, err := gs.repo.DeductBalance(msg.ClientID, msg.BetAmount)
	if err != nil {
		return PlayResponse{}, NewInternalError(err.Error())
	}
	diceSides := 6 // TODO: Implement more than 6 sided dice?
	diceResult := gs.rollDice(diceSides)
	haveWon := gs.calculateOutcome(msg.BetType, diceResult)

	if _, err = gs.repo.CreateGameSession(GameSessionRequest{
		PlayerID:     msg.ClientID,
		BetAmount:    msg.BetAmount,
		DiceResult:   diceResult,
		Won:          haveWon,
		Active:       true,
		SessionStart: time.Now(),
	}); err != nil {
		return PlayResponse{}, NewInternalError(err.Error())
	}

	return PlayResponse{DiceResult: diceResult, Won: haveWon, Balance: newBalance}, nil
}

func (gs *GameService) EndPlay(clientID int) (EndPlayResponse, error) {
	log.Printf("\nFinishing play session for client id -> %d", clientID)

	// VALIDATE USER HAS AN ACTIVE SESSION BEFORE ENDING THE PLAY SESSION
	activeSession, err := gs.repo.GetActiveSession(clientID)
	if err != nil {
		return EndPlayResponse{}, NewInternalError(err.Error())
	}
	if activeSession == nil {
		return EndPlayResponse{}, NewActiveSessionError(fmt.Sprintf("Client ID %d does not have an active session.", clientID))
	}

	var balance float64
	if activeSession.Won {
		// TODO: Add multiplier options ?
		multiplier := 2.0
		balance, err = gs.repo.IncreaseBalance(clientID, activeSession.BetAmount*multiplier)
		if err != nil {
			return EndPlayResponse{}, NewInternalError(err.Error())
		}
	} else {
		balance, err = gs.repo.GetBalance(clientID)
		if err != nil {
			return EndPlayResponse{}, err
		}
	}

	if err := gs.repo.CloseCurrentGameSession(clientID); err != nil {
		return EndPlayResponse{}, NewInternalError(err.Error())
	}

	return EndPlayResponse{ClientID: clientID, Balance: balance}, nil
}

// PRIVATE METHODS

func (gs *GameService) validateBetAmount(betAmount, balance float64) error {
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

func (gs *GameService) rollDice(sides int) int {
	return rand.Intn(sides) + 1
}

func (gs *GameService) calculateOutcome(betType BetType, diceResult int) bool {
	isEven := diceResult%2 == 0
	return (betType == Even && isEven) || (betType == Odd && !isEven)
}

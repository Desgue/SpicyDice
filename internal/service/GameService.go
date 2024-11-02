package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/Desgue/SpicyDice/internal/appErrors"
	"github.com/Desgue/SpicyDice/internal/domain"
)

const (
	MinBetAmount = 1.0
	MaxBetAmount = 1000.0
)

type GameService struct {
	repo domain.Repository
}

func NewGameService(repo domain.Repository) *GameService {
	return &GameService{
		repo: repo,
	}
}

func (gs *GameService) GetBalance(userID int) (domain.WalletResponse, error) {
	balance, err := gs.repo.GetBalance(userID)
	if err != nil {
		return domain.WalletResponse{}, appErrors.NewInternalError(err.Error())
	}
	return domain.WalletResponse{ClientID: userID, Balance: balance}, nil
}

func (gs *GameService) ProcessPlay(msg domain.PlayPayload) (domain.PlayResponse, error) {
	log.Printf("\nProcessing play for user id -> %d\nBet Amount -> %g\nBet Type -> %s", msg.ClientID, msg.BetAmount, msg.BetType)

	balance, err := gs.repo.GetBalance(msg.ClientID)
	if err != nil {
		return domain.PlayResponse{}, appErrors.NewInternalError(err.Error())
	}

	// VALIDATE USER BET AMOUNT
	if err := gs.validateBetAmount(msg.BetAmount, balance); err != nil {
		return domain.PlayResponse{}, err
	}

	// GAME LOGIC
	diceSides := 6 // TODO: Implement more than 6 sided dice?
	diceResult, err := gs.rollDice(diceSides)
	if err != nil {
		return domain.PlayResponse{}, appErrors.NewDiceRollError(err.Error())
	}
	haveWon := gs.calculateOutcome(msg.BetType, diceResult)

	_, newBalance, err := gs.repo.ExecutePlayTransaction(domain.PlayTransaction{
		Message:    msg,
		DiceResult: diceResult,
		Won:        haveWon,
	})
	gameRrr := &appErrors.GameError{}
	if err != nil {
		if errors.As(err, &gameRrr) {
			return domain.PlayResponse{}, err
		}
		return domain.PlayResponse{}, appErrors.NewInternalError(fmt.Sprintf("Error while executing play transaction: %s", err))
	}

	return domain.PlayResponse{DiceResult: diceResult, Won: haveWon, Balance: newBalance}, nil
}

func (gs *GameService) EndPlay(clientID int) (domain.EndPlayResponse, error) {
	log.Printf("\nFinishing play session for client id -> %d", clientID)

	// VALIDATE USER HAS AN ACTIVE SESSION BEFORE ENDING THE PLAY SESSION
	activeSession, err := gs.repo.GetActiveSession(clientID)
	if err != nil {
		return domain.EndPlayResponse{}, appErrors.NewInternalError(err.Error())
	}
	if activeSession == nil {
		return domain.EndPlayResponse{}, appErrors.NewActiveSessionError(fmt.Sprintf("Client ID %d does not have an active session.", clientID))
	}

	if err := gs.repo.CloseCurrentGameSession(clientID); err != nil {
		return domain.EndPlayResponse{}, appErrors.NewInternalError(err.Error())
	}

	return domain.EndPlayResponse{ClientID: clientID}, nil
}

// PRIVATE METHODS

func (gs *GameService) validateBetAmount(betAmount, balance float64) error {
	var details string

	if betAmount > balance {
		details = fmt.Sprintf("bet amount %.2f exceeds available balance %.2f", betAmount, balance)
		return appErrors.NewInsufficientFundsError(details)
	}

	if betAmount < 0 {
		details = fmt.Sprintf("bet amount cannot be negative: %.2f", betAmount)
		return appErrors.NewInvalidBetAmountError(details)
	}
	if betAmount == 0 {
		details = "bet amount cannot be zero"
		return appErrors.NewInvalidBetAmountError(details)
	}

	if betAmount < MinBetAmount {
		details = fmt.Sprintf("minimum bet amount is %.2f", MinBetAmount)
		return appErrors.NewInvalidBetAmountError(details)
	}

	if betAmount > MaxBetAmount {
		details = fmt.Sprintf("maximum bet amount is %.2f", MaxBetAmount)
		return appErrors.NewInvalidBetAmountError(details)
	}

	return nil
}

func (gs *GameService) rollDice(sides int) (int, error) {
	bigI, err := rand.Int(rand.Reader, big.NewInt(int64(sides)))
	if err != nil {
		return 0, err
	}
	roll := int(bigI.Int64()) + 1
	return roll, nil
}

func (gs *GameService) calculateOutcome(betType domain.BetType, diceResult int) bool {
	isEven := diceResult%2 == 0
	return (betType == domain.Even && isEven) || (betType == domain.Odd && !isEven)
}

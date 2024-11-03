package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/Desgue/SpicyDice/internal/appErrors"
	"github.com/Desgue/SpicyDice/internal/config"
	"github.com/Desgue/SpicyDice/internal/domain"
)

type GameService struct {
	repo domain.Repository
}

var (
	MinBetAmount = config.New().Game.MinBetAmount
	MaxBetAmount = config.New().Game.MaxBetAmount
)

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

func (gs *GameService) ProcessPlay(msg domain.PlayPayload, dice DiceRoller) (domain.PlayResponse, error) {
	log.Printf("\nProcessing play for user id -> %d\nBet Amount -> %g\nBet Type -> %s", msg.ClientID, msg.BetAmount, msg.BetType)

	balance, err := gs.repo.GetBalance(msg.ClientID)
	if err != nil {
		return domain.PlayResponse{}, appErrors.NewInternalError(err.Error())
	}
	if err := gs.validateBetAmount(msg.BetAmount, balance); err != nil {
		return domain.PlayResponse{}, err
	}

	diceResult, err := dice.Roll()
	if err != nil {
		return domain.PlayResponse{}, appErrors.NewDiceRollError(err.Error())
	}
	haveWon := gs.calculateOutcome(msg.BetType, diceResult)

	_, newBalance, err := gs.repo.ProcessPlay(domain.PlayTransaction{
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

	return domain.PlayResponse{DiceResult: diceResult, Won: haveWon, Balance: newBalance, BetAmount: msg.BetAmount}, nil
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

	if betAmount < config.New().Game.MinBetAmount {
		details = fmt.Sprintf("minimum bet amount is %.2f", MinBetAmount)
		return appErrors.NewInvalidBetAmountError(details)
	}

	if betAmount > config.New().Game.MaxBetAmount {
		details = fmt.Sprintf("maximum bet amount is %.2f", MaxBetAmount)
		return appErrors.NewInvalidBetAmountError(details)
	}

	return nil
}

type DiceRoller interface {
	Roll() (int, error)
}
type Dice struct {
	Sides int
}

func (d Dice) Roll() (int, error) {
	bigI, err := rand.Int(rand.Reader, big.NewInt(int64(d.Sides)))
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

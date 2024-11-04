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
	"github.com/Desgue/SpicyDice/internal/repository"
)

// GameService orchestrates game logic and wallet operations while maintaining transactional integrity
type GameService struct {
	repo repository.Repository
}

var (
	MinBetAmount = config.New().Game.MinBetAmount
	MaxBetAmount = config.New().Game.MaxBetAmount
)

// NewGameService follows the repository pattern for data persistence operations
func NewGameService(repo repository.Repository) *GameService {
	return &GameService{
		repo: repo,
	}
}

// GetBalance retrieves current balance ensuring player exists in the system
func (gs *GameService) GetBalance(playerID int) (domain.WalletResponse, error) {
	log.Printf("\nGetting balance for client id -> %d", playerID)
	balance, err := gs.repo.GetBalance(playerID)
	if err != nil {
		return domain.WalletResponse{}, appErrors.NewInternalError(err.Error())
	}
	return domain.WalletResponse{ClientID: playerID, Balance: balance}, nil
}

// ProcessPlay handles the complete game cycle: validation, dice roll, outcome calculation and balance update
// Returns error if any game rules are violated or system errors occur
func (gs *GameService) ProcessPlay(msg domain.PlayRequest, dice DiceRoller) (domain.PlayResponse, error) {
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

// EndPlay enforces game session closure rules and maintains data consistency
func (gs *GameService) EndPlay(clientID int) (domain.EndPlayResponse, error) {
	log.Printf("\nFinishing play session for client id -> %d", clientID)

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

// validateBetAmount enforces betting rules including minimum/maximum limits and available balance
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

// DiceRoller defines the contract for dice rolling implementations
type DiceRoller interface {
	Roll() (int, error)
}

// Dice implements secure random number generation for fair game outcomes
type Dice struct {
	Sides int
}

// Roll uses crypto/rand for cryptographically secure random number generation
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

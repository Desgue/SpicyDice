package service

import (
	"fmt"
	"log"
	"math/rand"
	"time"

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

	// VALIDATE USER DO NOT HAS AN ACTIVE SESSION BEFORE PLACING A BET
	activeSession, err := gs.repo.GetActiveSession(msg.ClientID)
	if err != nil {
		return domain.PlayResponse{}, appErrors.NewInternalError(err.Error())
	}
	if activeSession != nil {
		return domain.PlayResponse{}, appErrors.NewActiveSessionError("Cannot place a bet because the player already has an active session.")
	}

	balance, err := gs.repo.GetBalance(msg.ClientID)
	if err != nil {
		return domain.PlayResponse{}, appErrors.NewInternalError(err.Error())
	}

	// VALIDATE USER BET AMOUNT
	if err := gs.validateBetAmount(msg.BetAmount, balance); err != nil {
		return domain.PlayResponse{}, err
	}

	// GAME LOGIC
	newBalance, err := gs.repo.DeductBalance(msg.ClientID, msg.BetAmount)
	if err != nil {
		return domain.PlayResponse{}, appErrors.NewInternalError(err.Error())
	}
	diceSides := 6 // TODO: Implement more than 6 sided dice?
	diceResult := gs.rollDice(diceSides)
	haveWon := gs.calculateOutcome(msg.BetType, diceResult)

	if _, err = gs.repo.CreateGameSession(domain.GameSessionRequest{
		PlayerID:     msg.ClientID,
		BetAmount:    msg.BetAmount,
		DiceResult:   diceResult,
		Won:          haveWon,
		Active:       true,
		SessionStart: time.Now(),
	}); err != nil {
		return domain.PlayResponse{}, appErrors.NewInternalError(err.Error())
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

	var balance float64
	if activeSession.Won {
		// TODO: Add multiplier options ?
		multiplier := 2.0
		balance, err = gs.repo.IncreaseBalance(clientID, activeSession.BetAmount*multiplier)
		if err != nil {
			return domain.EndPlayResponse{}, appErrors.NewInternalError(err.Error())
		}
	} else {
		balance, err = gs.repo.GetBalance(clientID)
		if err != nil {
			return domain.EndPlayResponse{}, err
		}
	}

	if err := gs.repo.CloseCurrentGameSession(clientID); err != nil {
		return domain.EndPlayResponse{}, appErrors.NewInternalError(err.Error())
	}

	return domain.EndPlayResponse{ClientID: clientID, Balance: balance}, nil
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

func (gs *GameService) rollDice(sides int) int {
	return rand.Intn(sides) + 1
}

func (gs *GameService) calculateOutcome(betType domain.BetType, diceResult int) bool {
	isEven := diceResult%2 == 0
	return (betType == domain.Even && isEven) || (betType == domain.Odd && !isEven)
}

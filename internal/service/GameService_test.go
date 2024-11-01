package service

import (
	"fmt"
	"testing"

	"github.com/Desgue/SpicyDice/internal/appErrors"
	"github.com/Desgue/SpicyDice/internal/domain"
	"github.com/Desgue/SpicyDice/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessPlay(t *testing.T) {
	tests := []struct {
		name             string
		clientID         int
		betAmount        float64
		betType          domain.BetType
		setupMock        func(*repository.MockRepository)
		expectedError    string
		expectedBalance  float64
		expectedDiceRoll int
		expectedOutcome  bool
	}{
		{
			name:      "Valid Play",
			clientID:  1,
			betAmount: 100.0,
			betType:   domain.Even,
			setupMock: func(mockRepo *repository.MockRepository) {
				mockRepo.On("GetActiveSession", 1).Return(nil, nil)
				mockRepo.On("GetBalance", 1).Return(500.0, nil)
				mockRepo.On("DeductBalance", 1, 100.0).Return(400.0, nil)
				mockRepo.On("CreateGameSession", mock.Anything).Return(domain.GameSession{}, nil)
			},
			expectedError:   "",
			expectedBalance: 400.0,
		},
		{
			name:      "Insufficient Balance",
			clientID:  1,
			betAmount: 100.0,
			betType:   domain.Even,
			setupMock: func(mockRepo *repository.MockRepository) {
				mockRepo.On("GetActiveSession", 1).Return(nil, nil)
				mockRepo.On("GetBalance", 1).Return(50.0, nil)
			},
			expectedError: appErrors.NewInsufficientFundsError(fmt.Sprintf("bet amount %.2f exceeds available balance %.2f", 100.0, 50.0)).Error(),
		},
		{
			name:      "Session Reuse",
			clientID:  1,
			betAmount: 100.0,
			betType:   domain.Even,
			setupMock: func(mockRepo *repository.MockRepository) {
				mockRepo.On("GetActiveSession", 1).Return(&domain.GameSession{PlayerID: 1, Active: true}, nil)
			},
			expectedError: "active session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repository.MockRepository)
			service := NewGameService(mockRepo)
			tt.setupMock(mockRepo)

			response, err := service.ProcessPlay(domain.PlayPayload{ClientID: tt.clientID, BetAmount: tt.betAmount, BetType: tt.betType})

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBalance, response.Balance)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestValidateBetAmount(t *testing.T) {
	tests := []struct {
		name          string
		betAmount     float64
		balance       float64
		expectedError string
	}{
		{
			name:          "Valid Bet",
			betAmount:     100.0,
			balance:       500.0,
			expectedError: "",
		},
		{
			name:          "Bet Exceeds Balance",
			betAmount:     600.0,
			balance:       500.0,
			expectedError: "exceeds available balance",
		},
		{
			name:          "Negative Bet Amount",
			betAmount:     -50.0,
			balance:       500.0,
			expectedError: "bet amount cannot be negative",
		},
		{
			name:          "Zero Bet Amount",
			betAmount:     0.0,
			balance:       500.0,
			expectedError: "bet amount cannot be zero",
		},
		{
			name:          "Bet Below Minimum",
			betAmount:     0.5,
			balance:       500.0,
			expectedError: "minimum bet amount",
		},
		{
			name:          "Bet Above Maximum",
			betAmount:     1500.0,
			balance:       2000.0,
			expectedError: "maximum bet amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &GameService{}

			err := service.validateBetAmount(tt.betAmount, tt.balance)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

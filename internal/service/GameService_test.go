package service

import (
	"testing"

	"github.com/Desgue/SpicyDice/internal/appErrors"
	"github.com/Desgue/SpicyDice/internal/config"
	"github.com/Desgue/SpicyDice/internal/domain"
	"github.com/Desgue/SpicyDice/internal/repository"
	"github.com/stretchr/testify/assert"
)

type FakeDice struct {
}

func (fd FakeDice) Roll() (int, error) {
	return 1, nil
}

const (
	TestBalance             = 200.0
	TestValidBet            = 100.0
	TestInvalidBet          = 300.0
	TestPostValidBetBalance = TestBalance + TestValidBet
	TestBalanceWhenError    = 0.0
)

func TestProcessPlay_BusinessLogic(t *testing.T) {
	testCases := []struct {
		name              string
		payload           domain.PlayPayload
		setupMock         func(*repository.MockRepository)
		expectedBalance   float64
		expectedWin       bool
		expectError       bool
		expectedErrorCode int
	}{
		{
			name: "insuficient_funds",
			payload: domain.PlayPayload{
				ClientID:  1,
				BetAmount: TestInvalidBet,
				BetType:   domain.Even,
			},
			setupMock: func(mockRepo *repository.MockRepository) {
				mockRepo.On("GetBalance", 1).Return(TestBalance, nil)

			},
			expectedBalance:   TestBalance,
			expectedWin:       false,
			expectError:       true,
			expectedErrorCode: appErrors.InsufficientFundsErrorCode,
		},
		{
			name: "successful_bet_within_balance",
			payload: domain.PlayPayload{
				ClientID:  1,
				BetAmount: TestValidBet,
				BetType:   domain.Odd,
			},
			setupMock: func(mockRepo *repository.MockRepository) {
				mockRepo.On("GetBalance", 1).Return(TestBalance, nil)
				mockRepo.On("ExecutePlayTransaction", domain.PlayTransaction{
					Message: domain.PlayPayload{
						ClientID:  1,
						BetAmount: 100,
						BetType:   domain.Odd,
					},
					DiceResult: 1,
					Won:        true,
				}).Return(domain.GameSession{}, TestPostValidBetBalance, nil)
			},
			expectedBalance: TestPostValidBetBalance,
			expectedWin:     true,
			expectError:     false,
		},
		{
			name: "valid_bet_active_session_exists",
			payload: domain.PlayPayload{
				ClientID:  1,
				BetAmount: TestValidBet,
				BetType:   domain.Odd,
			},
			setupMock: func(mockRepo *repository.MockRepository) {
				mockRepo.On("GetBalance", 1).Return(TestBalance, nil)
				mockRepo.On("ExecutePlayTransaction", domain.PlayTransaction{
					Message: domain.PlayPayload{
						ClientID:  1,
						BetAmount: TestValidBet,
						BetType:   domain.Odd,
					},
					DiceResult: 1,
					Won:        true,
				}).Return(domain.GameSession{}, 0.0, appErrors.NewActiveSessionError(""))
			},
			expectedWin:       false,
			expectError:       true,
			expectedErrorCode: appErrors.ActiveSessionErrorCode,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repository.MockRepository)
			service := NewGameService(mockRepo)
			tt.setupMock(mockRepo)

			res, err := service.ProcessPlay(tt.payload, FakeDice{})
			assert.Equal(t, res.Won, tt.expectedWin)
			if tt.expectError {
				assert.Equal(t, err.(*appErrors.GameError).Code, tt.expectedErrorCode)
				assert.Empty(t, res, domain.PlayPayload{})
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestValidateBetAmount(t *testing.T) {
	tests := []struct {
		name              string
		betAmount         float64
		balance           float64
		expectError       bool
		expectedErrorCode int
	}{
		{
			name:        "Valid Bet",
			betAmount:   100.0,
			balance:     500.0,
			expectError: false,
		},
		{
			name:              "Bet Exceeds Balance",
			betAmount:         600.0,
			balance:           500.0,
			expectError:       true,
			expectedErrorCode: appErrors.InsufficientFundsErrorCode,
		},
		{
			name:              "Negative Bet Amount",
			betAmount:         -50.0,
			balance:           500.0,
			expectError:       true,
			expectedErrorCode: appErrors.InvalidBetAmountErrorCode,
		},
		{
			name:              "Zero Bet Amount",
			betAmount:         0.0,
			balance:           500.0,
			expectError:       true,
			expectedErrorCode: appErrors.InvalidBetAmountErrorCode,
		},
		{
			name:              "Bet Below Minimum",
			betAmount:         config.New().Game.MinBetAmount - 1,
			balance:           500.0,
			expectError:       true,
			expectedErrorCode: appErrors.InvalidBetAmountErrorCode,
		},
		{
			name:              "Bet Above Maximum",
			betAmount:         config.New().Game.MaxBetAmount + 1,
			balance:           2000.0,
			expectError:       true,
			expectedErrorCode: appErrors.InvalidBetAmountErrorCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &GameService{}

			err := service.validateBetAmount(tt.betAmount, tt.balance)

			if tt.expectError {
				assert.Equal(t, err.(*appErrors.GameError).Code, tt.expectedErrorCode)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

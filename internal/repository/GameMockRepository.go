package repository

import (
	"github.com/Desgue/SpicyDice/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetBalance(playerID int) (float64, error) {
	args := m.Called(playerID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRepository) GetActiveSession(playerID int) (*domain.GameSession, error) {
	args := m.Called(playerID)
	if session, ok := args.Get(0).(*domain.GameSession); ok {
		return session, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRepository) CloseCurrentGameSession(clientID int) error {
	args := m.Called(clientID)
	return args.Error(0)
}

func (m *MockRepository) ProcessPlay(transaction domain.PlayTransaction) (domain.GameSession, float64, error) {
	args := m.Called(transaction)
	return args.Get(0).(domain.GameSession), args.Get(1).(float64), args.Error(2)
}

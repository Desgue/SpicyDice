package repository

import (
	"github.com/Desgue/SpicyDice/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockRepository implements the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetBalance(playerID int) (float64, error) {
	args := m.Called(playerID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRepository) DeductBalance(playerID int, amount float64) (float64, error) {
	args := m.Called(playerID, amount)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRepository) IncreaseBalance(playerID int, amount float64) (float64, error) {
	args := m.Called(playerID, amount)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRepository) CreateGameSession(sess domain.GameSessionRequest) (domain.GameSession, error) {
	args := m.Called(sess)
	return args.Get(0).(domain.GameSession), args.Error(1)
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
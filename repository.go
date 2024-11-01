package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type GameRepository struct {
	db *sql.DB
}

func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{
		db: db,
	}
}

func (gr *GameRepository) GetBalance(userID int) (float64, error) {
	var balance float64
	query := `SELECT balance FROM player WHERE id = $1`
	if err := gr.db.QueryRow(query, userID).Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, NewUserNotFoundError(fmt.Sprintf("No player found with ID: %d", userID))
		}
		return 0, NewInternalError(fmt.Sprintf("Database error: %v", err))
	}
	return balance, nil
}
func (gr *GameRepository) DeductBalance(userId int, amount float64) (float64, error) {
	var newBalance float64
	query := `
		UPDATE player
		SET balance = balance - $1
		WHERE id = $2 
		AND balance >= $1
		RETURNING balance
		;`
	if err := gr.db.QueryRow(query, amount, userId).Scan(&newBalance); err != nil {
		return 0, err
	}

	return newBalance, nil

}
func (gr *GameRepository) ProcessPlay() {}
func (gr *GameRepository) EndPlay()     {}

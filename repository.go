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

func (gr *GameRepository) GetBalance(playerID int) (float64, error) {
	var balance float64
	query := `SELECT balance FROM player WHERE id = $1`
	if err := gr.db.QueryRow(query, playerID).Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, NewUserNotFoundError(fmt.Sprintf("No player found with ID: %d", playerID))
		}
		return 0, NewInternalError(fmt.Sprintf("Database error: %v", err))
	}
	return balance, nil
}
func (gr *GameRepository) DeductBalance(playerID int, amount float64) (float64, error) {
	var newBalance float64
	query := `
		UPDATE player
		SET balance = balance - $1
		WHERE id = $2 
		AND balance >= $1
		RETURNING balance
		;`
	if err := gr.db.QueryRow(query, amount, playerID).Scan(&newBalance); err != nil {
		return 0, fmt.Errorf("error deducting balance for player id %d", playerID)
	}

	return newBalance, nil

}
func (gr *GameRepository) IncreaseBalance(playerID int, amount float64) (float64, error) {
	var newBalance float64
	query := `
		UPDATE player
		SET balance = balance + $1
		WHERE id = $2 
		RETURNING balance
		;`
	if err := gr.db.QueryRow(query, amount, playerID).Scan(&newBalance); err != nil {
		return 0, fmt.Errorf("error increasing balance for player id %d", playerID)
	}

	return newBalance, nil

}

func (gr *GameRepository) CreateGameSession(sess GameSessionRequest) (GameSession, error) {
	var session GameSession
	query := `
		INSERT INTO game_session (player_id, bet_amount, dice_result, won, active, session_start)
		VALUES
		($1, $2, $3, $4, $5, $6)
		RETURNING session_id, player_id, bet_amount, dice_result, won, active, session_start, session_end
	;`

	err := gr.db.QueryRow(
		query,
		sess.PlayerID,
		sess.BetAmount,
		sess.DiceResult,
		sess.Won,
		sess.Active,
		sess.SessionStart,
	).
		Scan(
			&session.SessionID,
			&session.PlayerID,
			&session.BetAmount,
			&session.DiceResult,
			&session.Won,
			&session.Active,
			&session.SessionStart,
			&session.SessionEnd,
		)
	if err != nil {
		return GameSession{}, fmt.Errorf("error creating game session for player id %d", sess.PlayerID)
	}

	return session, nil

}
func (gr *GameRepository) GetActiveSession(playerID int) (*GameSession, error) {
	var session GameSession
	query := `
		SELECT session_id, player_id, bet_amount, dice_result, won, active, session_start, session_end FROM game_session 
		WHERE player_id = $1
		AND active = true
	;`

	err := gr.db.QueryRow(
		query,
		playerID).
		Scan(
			&session.SessionID,
			&session.PlayerID,
			&session.BetAmount,
			&session.DiceResult,
			&session.Won,
			&session.Active,
			&session.SessionStart,
			&session.SessionEnd,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error retrieving active session for player id %d", playerID)
	}

	return &session, nil
}

func (gr *GameRepository) CloseCurrentGameSession(clientID int) error {
	query := `
		UPDATE game_session
		SET active = false, session_end = NOW()
		WHERE player_id = $1 AND active = true
		;`

	result, err := gr.db.Exec(query, clientID)
	if err != nil {
		return fmt.Errorf("error updating game session: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no active session found for player id %d", clientID)
	}

	return nil
}

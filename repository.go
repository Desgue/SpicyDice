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

func (gr *GameRepository) CreateGameSession(sess GameSessionRequest) (GameSession, error) {
	var session GameSession
	query := `
		INSERT INTO game_session (player_id, bet_amount, dice_result, won, active, session_start)
		VALUES
		($1, $2, $3, $4, $5, $6)
		RETURNING (session_id, player_id, bet_amount, dice_result, won, active, session_start, session_end)
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
		return GameSession{}, err
	}

	return session, nil

}
func (gr *GameRepository) GetGameSession(playerID int) (GameSession, error) {
	var session GameSession
	query := `
		SELECT session_id, player_id, bet_amount, dice_result, won, active, session_start, session_end FROM game_session 
		WHERE session_id = $1
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
		return GameSession{}, err
	}

	return session, nil
}
func (gr *GameRepository) EndPlay() {}

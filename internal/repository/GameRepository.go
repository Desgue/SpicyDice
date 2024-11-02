package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Desgue/SpicyDice/internal/appErrors"
	"github.com/Desgue/SpicyDice/internal/domain"
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
			return 0, appErrors.NewUserNotFoundError(fmt.Sprintf("No player found with ID: %d", playerID))
		}
		return 0, appErrors.NewInternalError(fmt.Sprintf("Database error: %v", err))
	}
	return balance, nil
}

func (gr *GameRepository) updateBalance(tx *sql.Tx, playerID int, newBalance float64) (float64, error) {
	var endBalance float64
	query := `
		UPDATE player
		SET balance = $1
		WHERE id = $2 
		RETURNING balance
		;`
	if err := tx.QueryRow(query, newBalance, playerID).Scan(&endBalance); err != nil {
		return 0, fmt.Errorf("error updating balance for player id %d", playerID)
	}

	return endBalance, nil

}
func (gr *GameRepository) GetActiveSession(playerID int) (*domain.GameSession, error) {
	var session domain.GameSession
	query := `
		SELECT session_id, player_id, bet_amount, dice_result, won, active, session_start, session_end FROM game_session 
		WHERE player_id = $1
		AND active = true
	;`

	err := gr.db.QueryRow(
		query,
		playerID,
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error retrieving active session for player id %d", playerID)
	}

	return &session, nil
}

func (gr *GameRepository) CreateGameSession(sess domain.GameSessionRequest) (domain.GameSession, error) {
	tx, err := gr.db.Begin()
	if err != nil {
		return domain.GameSession{}, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()
	return gr.createGameSession(tx, sess)
}

func (gr *GameRepository) createGameSession(tx *sql.Tx, sess domain.GameSessionRequest) (domain.GameSession, error) {
	var session domain.GameSession
	query := `
		INSERT INTO game_session (player_id, bet_amount, dice_result, won, active, session_start)
		VALUES
		($1, $2, $3, $4, $5, $6)
		RETURNING session_id, player_id, bet_amount, dice_result, won, active, session_start, session_end
	;`

	err := tx.QueryRow(
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
		return domain.GameSession{}, fmt.Errorf("error creating game session for player id %d", sess.PlayerID)
	}

	return session, nil

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

func (gr *GameRepository) ExecutePlayTransaction(msg domain.PlayPayload, diceResult int, won bool, balance float64) (domain.GameSession, float64, error) {
	var session domain.GameSession
	var multiplier = 2.0

	tx, err := gr.db.Begin()
	if err != nil {
		return domain.GameSession{}, 0, appErrors.NewInternalError(fmt.Sprintf("error creating database transaction: %s", err))
	}
	defer tx.Rollback()

	// Check for active session
	activeSession, err := gr.GetActiveSession(msg.ClientID)
	if err != nil {
		return domain.GameSession{}, 0, appErrors.NewInternalError(err.Error())
	}
	if activeSession != nil {
		return domain.GameSession{}, 0, appErrors.NewActiveSessionError("Player already has an active session")
	}

	// Create the game session
	session, err = gr.createGameSession(tx, domain.GameSessionRequest{
		PlayerID:     msg.ClientID,
		BetAmount:    msg.BetAmount,
		DiceResult:   diceResult,
		Won:          won,
		Active:       true,
		SessionStart: time.Now(),
	})
	if err != nil {
		return domain.GameSession{}, 0, err
	}

	// Update balance with win multiplier if applicable
	if won {
		balance += (msg.BetAmount * multiplier) - msg.BetAmount
	} else {
		balance -= msg.BetAmount
	}

	// Save new balance to DB
	_, err = gr.updateBalance(tx, msg.ClientID, balance)
	if err != nil {
		return domain.GameSession{}, 0, err
	}

	if err = tx.Commit(); err != nil {
		return domain.GameSession{}, 0, appErrors.NewInternalError("error commiting transaction")
	}

	return session, balance, nil
}

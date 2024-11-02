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

func (gr *GameRepository) ExecutePlayTransaction(t domain.PlayTransaction) (domain.GameSession, float64, error) {
	var session domain.GameSession
	var multiplier = 2.0
	var changeAmount float64

	tx, err := gr.db.Begin()
	if err != nil {
		return domain.GameSession{}, 0, appErrors.NewInternalError(fmt.Sprintf("error creating database transaction: %s", err))
	}
	defer tx.Rollback()

	activeSession, err := gr.GetActiveSession(t.Message.ClientID)
	if err != nil {
		return domain.GameSession{}, 0, appErrors.NewInternalError(err.Error())
	}
	if activeSession != nil {
		return domain.GameSession{}, 0, appErrors.NewActiveSessionError("Player already has an active session")
	}

	session, err = gr.createGameSession(tx, domain.GameSessionRequest{
		PlayerID:     t.Message.ClientID,
		BetAmount:    t.Message.BetAmount,
		DiceResult:   t.DiceResult,
		Won:          t.Won,
		Active:       true,
		SessionStart: time.Now(),
	})
	if err != nil {
		return domain.GameSession{}, 0, err
	}

	if t.Won {
		changeAmount = (t.Message.BetAmount * multiplier) - t.Message.BetAmount
	} else {
		changeAmount = -t.Message.BetAmount
	}

	err = gr.updateBalance(tx, domain.BalanceUpdate{
		PlayerID:     t.Message.ClientID,
		ChangeAmount: changeAmount,
	})
	if err != nil {
		return domain.GameSession{}, 0, err
	}

	if err = tx.Commit(); err != nil {
		return domain.GameSession{}, 0, fmt.Errorf("failed to commit play transaction: %w", err)
	}

	return session, 0, nil
}

func (gr *GameRepository) updateBalance(tx *sql.Tx, update domain.BalanceUpdate) error {
	var currBalance float64
	balanceLockQuery := `
		SELECT balance FROM player
		WHERE id = $1
		FOR UPDATE
		;`
	if err := tx.QueryRow(balanceLockQuery, update.PlayerID).Scan(&currBalance); err != nil {
		return fmt.Errorf("error locking row: %w", err)
	}

	newBalance := currBalance + update.ChangeAmount
	if err := validateBalance(newBalance); err != nil {
		return err
	}

	updateQuery := `
	UPDATE player 
	SET 
		balance = $1
	WHERE 
		id = $2 
	`

	result, err := tx.Exec(updateQuery, newBalance, update.PlayerID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if affected == 0 {
		return ErrUnaffectedRows
	}
	return nil

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

func validateBalance(balance float64) error {
	if balance < 0 {
		return ErrNegativeBalance
	}
	return nil
}

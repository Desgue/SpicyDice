package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/Desgue/SpicyDice/internal/appErrors"
	"github.com/Desgue/SpicyDice/internal/domain"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testDBConfig helps us to setup test scenarios with a nice chain of functions
// Useful to prepare the database before each test in a clean way
type testDBConfig struct {
	ctx        context.Context
	t          *testing.T
	db         *sql.DB
	operations []func(*sql.Tx) error
}

// NewTestConfig creates the DB Config
// It already includes the table creation in the operations list
func NewTestConfig(ctx context.Context, t *testing.T, db *sql.DB) *testDBConfig {
	cfg := &testDBConfig{
		ctx: ctx,
		t:   t,
		db:  db,
	}
	cfg.operations = append(cfg.operations, cfg.createTables)
	return cfg
}

// createTables makes our test database schema with indexes
// We need this to guarantee db consistency during the tests
func (cfg *testDBConfig) createTables(tx *sql.Tx) error {
	playerTable := `
	CREATE TABLE IF NOT EXISTS player (
		id SERIAL PRIMARY KEY,
		balance decimal(10,2)
	  );`
	gameSessionTable := `
	  CREATE TABLE IF NOT EXISTS game_session (
		session_id  SERIAL PRIMARY KEY,
		player_id int,
		bet_amount decimal(10,2),
		dice_result int,
		won boolean,
		active boolean,
		session_start timestamptz,
		session_end timestamptz DEFAULT NULL,
		FOREIGN KEY (player_id) REFERENCES player (id) ON DELETE CASCADE
	  );`
	createUniqueIndex := `
	  CREATE UNIQUE INDEX unique_active_player_session ON game_session (player_id)
	  WHERE active = true;
	   
	  `

	if _, err := tx.ExecContext(cfg.ctx, playerTable); err != nil {
		return err
	}
	if _, err := tx.ExecContext(cfg.ctx, gameSessionTable); err != nil {
		return err
	}
	if _, err := tx.ExecContext(cfg.ctx, createUniqueIndex); err != nil {
		return err
	}
	return nil
}

// WithPlayer adds a new player to the operations queue
// Good to test different player scenarios with clean setup
func (cfg *testDBConfig) WithPlayer(id int, balance float64) *testDBConfig {
	insertPlayer := func(tx *sql.Tx) error {
		query := `
		INSERT INTO player (id, balance)
		VALUES ($1, $2);`
		if _, err := tx.ExecContext(cfg.ctx, query, id, balance); err != nil {
			return fmt.Errorf("WithPlpayer: error inserting player: %w", err)
		}
		return nil
	}
	cfg.operations = append(cfg.operations, insertPlayer)
	return cfg
}

// WithActiveSession adds a game session to the operations queue
// Good to test cases where player already has active session
func (cfg *testDBConfig) WithActiveSession(session domain.GameSession) *testDBConfig {
	insertSession := func(tx *sql.Tx) error {
		query := `
		INSERT INTO game_session (session_id, player_id, bet_amount, dice_result, won, active, session_start)
		VALUES
		($1, $2, $3, $4, $5, $6, $7);`
		if _, err := tx.ExecContext(
			cfg.ctx,
			query,
			session.SessionID,
			session.PlayerID,
			session.BetAmount,
			session.DiceResult,
			session.Won,
			session.Active,
			session.SessionStart); err != nil {

			return fmt.Errorf("WithActiveSession: error inserting active session: %w", err)
		}
		return nil
	}
	cfg.operations = append(cfg.operations, insertSession)
	return cfg
}

// Setup runs all operations in sequence inside a transaction
func (cfg *testDBConfig) Setup() error {
	tx, err := cfg.db.Begin()
	if err != nil {
		return fmt.Errorf("Setup: error starting transaction: %w", err)
	}
	defer tx.Rollback()

	for _, operation := range cfg.operations {
		if err := operation(tx); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// Cleanup removes all test data and resets operations list
// Important to mantain a clean state after every test
func (cfg *testDBConfig) Cleanup() error {
	tx, err := cfg.db.Begin()
	if err != nil {
		return fmt.Errorf("Cleanup: error starting transaction: %w", err)
	}
	defer tx.Rollback()

	queries := []string{
		`DELETE FROM game_session;`,
		`DELETE FROM player;`,
	}

	for _, query := range queries {
		if _, err := tx.ExecContext(cfg.ctx, query); err != nil {
			return fmt.Errorf("Cleanup: error executing query %s: %w", query, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Cleanup: error committing transaction: %w", err)
	}

	cfg.operations = []func(*sql.Tx) error{cfg.createTables}
	return nil
}

// TestProcessPlay checks if game rules are working right in database
// Here we use docker to create a postgres container to ensure a clean state every test
func TestProcessPlay(t *testing.T) {
	ctx := context.Background()
	dbName := "postgres"
	dbUser := "postgres"
	dbPassword := "postgres"

	postgresContainer, err := postgres.Run(ctx,
		"docker.io/postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	connStr, _ := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("error open database: %s", err.Error())
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatalf("could not reach database: %s", err)
	}
	testDB := NewTestConfig(ctx, t, db)
	repo := NewGameRepository(db)

	testCases := []struct {
		name                    string
		expectError             bool
		expectedErrorCode       int
		expectedBalance         float64
		expectedSessionResponse bool
		setupFunc               func(cfg *testDBConfig) *testDBConfig
		transaction             domain.PlayTransaction
	}{
		// Validates that a player cannot start a new game session while having an active one
		{
			name:                    "active_session_exists",
			expectError:             true,
			expectedErrorCode:       appErrors.ActiveSessionErrorCode,
			expectedBalance:         0.0,
			expectedSessionResponse: true,
			setupFunc: func(cfg *testDBConfig) *testDBConfig {
				// Simulates an existing active session for the player
				playerId, sessionId := 1, 1
				config := cfg.WithPlayer(playerId, 1000).WithActiveSession(domain.GameSession{
					SessionID:    sessionId,
					PlayerID:     playerId,
					BetAmount:    100,
					DiceResult:   1,
					Won:          true,
					Active:       true,
					SessionStart: time.Now(),
				})

				return config
			},
			transaction: domain.PlayTransaction{
				Message: domain.PlayRequest{
					ClientID:  1,
					BetAmount: 100,
					BetType:   domain.Odd,
				},
				DiceResult: 1,
				Won:        true,
			},
		},
	}

	for _, tc := range testCases {
		// Simulates an existing active session for the player
		cfg := tc.setupFunc(testDB)
		if err := cfg.Setup(); err != nil {
			t.Fatalf("error configuring test database: %s", err)
		}

		session, balance, err := repo.ProcessPlay(tc.transaction)

		assert.Equal(t, tc.expectedBalance, balance)

		if tc.expectedSessionResponse {
			assert.NotEmpty(t, session)
		} else {
			assert.Empty(t, session)
		}
		if tc.expectError {
			assert.Equal(t, tc.expectedErrorCode, err.(*appErrors.GameError).Code)
			assert.Zero(t, balance)

		} else {
			assert.Nil(t, err)
		}

		if err := cfg.Cleanup(); err != nil {
			t.Fatal(err)
		}
	}
}

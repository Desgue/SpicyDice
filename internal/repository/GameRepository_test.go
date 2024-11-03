package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testDBConfig struct {
	ctx        context.Context
	t          *testing.T
	db         *sql.DB
	NumPlayers int
	Balance    int
}

func (cfg *testDBConfig) configTestDatabase() error {
	cfg.t.Helper()
	tx, err := cfg.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()
	if err := cfg.createTables(tx); err != nil {
		return fmt.Errorf("error creating table: %w", err)
	}

	if err := cfg.populateTables(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commiting transaction: %w", err)
	}
	return nil
}
func (cfg *testDBConfig) createTables(tx *sql.Tx) error {
	playerTable := `
	CREATE TABLE IF NOT EXISTS player (
		id SERIAL PRIMARY KEY,
		balance decimal(10,2)
	  );`
	gameSessionTable := `
	  CREATE TABLE IF NOT EXISTS  game_session (
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
	initTables := func(t *testing.T) error {
		t.Helper()
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
	return initTables(cfg.t)
}
func (cfg *testDBConfig) populateTables(tx *sql.Tx) error {
	// variables needed: Number of players, Balance
	insertQuery := `
	INSERT INTO player (balance)
	VALUES ($1);`

	insertData := func(t *testing.T) error {
		t.Helper()
		if _, err := tx.ExecContext(cfg.ctx, insertQuery, cfg.Balance); err != nil {
			return err
		}

		return nil
	}
	if err := insertData(cfg.t); err != nil {
		return err
	}
	return nil
}

func TestExecutePlayTransaction(t *testing.T) {
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
		log.Fatalf("failed to start container: %s", err)
	}

	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	connStr, _ := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("error open database: %s", err.Error())
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("could not reach database: %s", err)
	}
	/* repo := NewGameRepository(db) */
	config := testDBConfig{
		ctx:        ctx,
		db:         db,
		t:          t,
		NumPlayers: 10,
		Balance:    2000,
	}
	if err := config.configTestDatabase(); err != nil {
		log.Fatalf("error configuring test database: %s", err)
	}

}

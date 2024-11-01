package main

import (
	"database/sql"
)

type GameRepository struct {
	db *sql.DB
}

func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{
		db: db,
	}
}

func (gr *GameRepository) GetBalance()  {}
func (gr *GameRepository) ProcessPlay() {}
func (gr *GameRepository) EndPlay()     {}

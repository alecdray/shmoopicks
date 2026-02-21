package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

const migrationsDir = "db/migrations"

type DB struct {
	*sql.DB
}

func NewDB(filepath string) (*DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := runMigrations(db); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{db}, nil
}

func runMigrations(db *sql.DB) error {
	err := goose.Up(db, migrationsDir)

	if errors.Is(err, goose.ErrNoMigrationFiles) {
		// Do nothing, no migrations found
	} else if err != nil {
		return err
	}

	return nil
}

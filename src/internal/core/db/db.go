package db

import (
	"database/sql"
	"errors"
	"shmoopicks/src/internal/core/db/sqlc"
	"time"

	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3"
)

const migrationsDir = "db/migrations"

type DB struct {
	sql     *sql.DB
	queries *sqlc.Queries
}

func NewDB(filepath string) (*DB, error) {
	sqlDb, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	if err := sqlDb.Ping(); err != nil {
		return nil, err
	}

	sqlDb.SetMaxOpenConns(25)
	sqlDb.SetMaxIdleConns(5)
	sqlDb.SetConnMaxLifetime(5 * time.Minute)

	queries := sqlc.New(sqlDb)
	db := &DB{sql: sqlDb, queries: queries}

	if err := db.runMigrations(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) Sql() *sql.DB {
	return db.sql
}

func (db *DB) Queries() *sqlc.Queries {
	return db.queries
}

func (db *DB) Close() error {
	return db.sql.Close()
}

func (db *DB) runMigrations() error {
	err := goose.Up(db.sql, migrationsDir)

	if errors.Is(err, goose.ErrNoMigrationFiles) {
		// Do nothing, no migrations found
	} else if err != nil {
		return err
	}

	return nil
}

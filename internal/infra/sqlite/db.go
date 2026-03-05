package sqlite

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type Config struct {
	DSN string
}

func NewDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite", cfg.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

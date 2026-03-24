package db

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewConnection(connectionString string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

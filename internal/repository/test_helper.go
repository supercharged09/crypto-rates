package repository

import (
	"database/sql"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// setupTestDB подключение к тестовой БД
func setupTestDB() (*sql.DB, error) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://crypto:crypto@localhost:5432/crypto_rates?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

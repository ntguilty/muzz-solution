package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

func NewSQLDB() (*sql.DB, error) {
	psqlInfo := os.Getenv("POSTGRES_DSN")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("error during creating postgres default db: %w", err)

	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return db, nil
}

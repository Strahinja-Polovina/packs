package database

import (
	"fmt"

	"github.com/Strahinja-Polovina/packs/pkg/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewConnection(dbConfig *config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dbConfig.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

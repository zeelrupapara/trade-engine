package database

import (
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/zeelrupapara/trade-engine/config"
)

var dbInstance *goqu.Database

const (
	POSTGRES = "postgres"
)

// Connect establishes a new PostgreSQL connection using Goqu and returns the database instance.
func Connect(cfg config.DBConfig) (*goqu.Database, error) {
	// Build PostgreSQL connection URL
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Db,
		cfg.QueryString,
	)

	// Open SQL DB connection
	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("✅ Connected to PostgreSQL at", dbURL)

	// Register Goqu dialect and return instance
	dbInstance = goqu.Dialect("postgres").DB(sqlDB)
	return dbInstance, nil
}

// GetDB returns the already connected Goqu database instance.
func GetDB() *goqu.Database {
	return dbInstance
}

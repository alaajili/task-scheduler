package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/alaajili/task-scheduler/shared/config"
)

var sqlOpen = sql.Open

// DB wraps sql.DB to provide additional functionality if needed.
type DB struct {
	*sql.DB
}

// NewPostgresDB creates a new DB instance for PostgreSQL.
func NewPostgresDB(cfg config.DatabaseConfig) (*DB, error) {
	db, err := sqlOpen("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure the database connection pool
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(time.Hour)

	// Tetst the database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	return &DB{DB: db}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}

// HealthCheck performs a simple health check on the database.
func (db *DB) HealthCheck(ctx context.Context) error {
	return db.PingContext(ctx)
}

package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// DB is the global database connection instance
var DB *sql.DB

// Connect establishes a connection to the PostgreSQL database
func Connect() error {
	var connStr string

	// Check if DATABASE_URL is set (used in production)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		// Use DATABASE_URL directly (production mode)
		connStr = databaseURL
	} else {
		// Build connection string from individual env vars (local development)
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSLMODE")

		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "5433"
		}
		if sslmode == "" {
			sslmode = "disable"
		}

		connStr = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode,
		)
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	// Configure connection pool
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	return nil
}

// Health checks database connectivity with a timeout
// Returns nil if healthy, error otherwise
func Health(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Create a context with timeout if none provided
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := DB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

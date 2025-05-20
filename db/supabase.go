package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is the global database connection pool
var DB *pgxpool.Pool

// InitializeDB initializes the database connection pool using pgxpool for Supabase
func InitializeDB() (*pgxpool.Pool, error) {
	// Get connection string from environment variable
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	log.Println("Connecting to database using DATABASE_URL...")

	// Create the connection with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Configure the connection pool
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	// Set pool configuration
	config.MaxConns = 10                      // Maximum number of connections in the pool
	config.MinConns = 2                       // Minimum number of connections in the pool
	config.MaxConnLifetime = 1 * time.Hour    // Maximum connection lifetime
	config.MaxConnIdleTime = 30 * time.Minute // Maximum idle time for a connection

	// Connect using pgxpool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable UUID extension if not exists (optional)
	_, err = pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	if err != nil {
		log.Printf("Note: Could not create uuid-ossp extension (it might already exist): %v", err)
	}

	log.Println("Successfully connected to the database pool")
	DB = pool
	return pool, nil
}

// CloseDB closes the database connection pool
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

// GetDB returns the database connection pool
func GetDB() *pgxpool.Pool {
	return DB
}

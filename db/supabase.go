package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

// DB is the global database connection
var DB *pgx.Conn

// InitializeDB initializes the database connection using pgx for Supabase
func InitializeDB() (*pgx.Conn, error) {
	// Get connection string from environment variable
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	log.Println("Connecting to database using DATABASE_URL...")

	// Create the connection with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect using pgx directly
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Test the connection
	if err := conn.Ping(ctx); err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable UUID extension if not exists (optional)
	_, err = conn.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	if err != nil {
		log.Printf("Note: Could not create uuid-ossp extension (it might already exist): %v", err)
	}

	log.Println("Successfully connected to the database")
	DB = conn
	return conn, nil
}

// CloseDB closes the database connection
func CloseDB() {
	if DB != nil {
		DB.Close(context.Background())
	}
}

// GetDB returns the database connection
func GetDB() *pgx.Conn {
	return DB
}

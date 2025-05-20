package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BaseRepository struct {
	db *pgxpool.Pool
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *pgxpool.Pool) *BaseRepository {
	return &BaseRepository{db: db}
}

// Query executes a query that returns rows
func (r *BaseRepository) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	// Acquire a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection from pool: %w", err)
	}
	defer conn.Release()

	return conn.Query(ctx, query, args...)
}

// QueryRow executes a query that returns at most one row
func (r *BaseRepository) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	// Acquire a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil // Can't return error with pgx.Row interface
	}
	// Note: We can't defer conn.Release() here because the row may be read after this function returns
	// The connection will be released when the row is closed by the caller

	return conn.QueryRow(ctx, query, args...)
}

// QueryRowTx is a safer alternative that gets a connection and handles releasing it
func (r *BaseRepository) QueryRowTx(ctx context.Context, query string, scanFn func(pgx.Row) error, args ...interface{}) error {
	// Acquire a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection from pool: %w", err)
	}
	defer conn.Release()

	row := conn.QueryRow(ctx, query, args...)
	return scanFn(row)
}

// Exec executes a query without returning any rows
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) error {
	// Acquire a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection from pool: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, query, args...)
	return err
}

// Begin starts a transaction
func (r *BaseRepository) Begin(ctx context.Context) (pgx.Tx, error) {
	// Acquire a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection from pool: %w", err)
	}
	// Note: We deliberately don't defer conn.Release() here because the connection
	// will be used for the entire transaction and should be released after tx.Commit()/Rollback()

	return conn.Begin(ctx)
}

// WithTx executes a function within a transaction
func (r *BaseRepository) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	// Acquire a connection from the pool
	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection from pool: %w", err)
	}
	defer conn.Release()

	// Begin transaction on the acquired connection
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	err = fn(tx)
	return err
}

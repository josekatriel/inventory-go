package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type BaseRepository struct {
	db *pgx.Conn
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *pgx.Conn) *BaseRepository {
	return &BaseRepository{db: db}
}

// Query executes a query that returns rows
func (r *BaseRepository) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return r.db.Query(ctx, query, args...)
}

// QueryRow executes a query that returns at most one row
func (r *BaseRepository) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return r.db.QueryRow(ctx, query, args...)
}

// Exec executes a query without returning any rows
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// Begin starts a transaction
func (r *BaseRepository) Begin(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

// WithTx executes a function within a transaction
func (r *BaseRepository) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := r.Begin(ctx)
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

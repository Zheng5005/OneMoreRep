package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a pgx connection pool and provides transaction helpers.
type DB struct {
	pool *pgxpool.Pool
}

// NewDB creates a new DB instance from a connection string.
func NewDB(ctx context.Context, connString string) (*DB, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{pool: pool}, nil
}

// Close releases all resources held by the connection pool.
func (d *DB) Close() {
	d.pool.Close()
}

// Ping checks if the database is reachable.
func (d *DB) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx)
}

// Pool returns the underlying connection pool.
func (d *DB) Pool() *pgxpool.Pool {
	return d.pool
}

// WithTx executes the given function inside a database transaction.
// The transaction is committed if fn returns nil, otherwise it is rolled back.
func (d *DB) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

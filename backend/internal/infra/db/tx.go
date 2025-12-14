package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxManager manages database transactions
type TxManager interface {
	// WithTx executes fn within a transaction
	// If fn returns an error, the transaction is rolled back
	// Otherwise, the transaction is committed
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

// PgxTxManager is a PostgreSQL transaction manager using pgx
type PgxTxManager struct {
	pool *pgxpool.Pool
}

// NewPgxTxManager creates a new PgxTxManager
func NewPgxTxManager(pool *pgxpool.Pool) *PgxTxManager {
	return &PgxTxManager{pool: pool}
}

// txKey is the context key for storing the transaction
type txKeyType struct{}

var txKey = txKeyType{}

// WithTx executes fn within a transaction
func (m *PgxTxManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	// Begin transaction
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}

	// Store transaction in context
	txCtx := context.WithValue(ctx, txKey, tx)

	// Execute function
	err = fn(txCtx)
	if err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	// Commit on success
	return tx.Commit(ctx)
}

// GetTx retrieves the transaction from context, or returns the pool if no transaction exists
// This is a helper for Repository implementations
func GetTx(ctx context.Context, pool *pgxpool.Pool) pgxQuery {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return pool
}

// pgxQuery is an interface that both pgxpool.Pool and pgx.Tx implement
// This allows repositories to work with either a pool or a transaction
type pgxQuery interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

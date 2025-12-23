package services

import "context"

// TxManager is an interface for managing database transactions.
// This interface allows the Application layer to be independent of
// the specific database implementation.
type TxManager interface {
	// WithTx executes the given function within a transaction.
	// If the function returns an error, the transaction is rolled back.
	// If the function returns nil, the transaction is committed.
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

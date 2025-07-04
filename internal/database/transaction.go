package database

import (
	"context"
	"database/sql"
	"fmt"
)

// TransactionFunc represents a function that can be executed within a transaction
type TransactionFunc func(tx *sql.Tx) error

// WithTransaction executes the given function within a database transaction
// It automatically handles transaction rollback on error and commit on success
func WithTransaction(db *sql.DB, fn TransactionFunc) error {
	return WithTransactionContext(context.Background(), db, fn)
}

// WithTransactionContext executes the given function within a database transaction with context support
// It automatically handles transaction rollback on error and commit on success
func WithTransactionContext(ctx context.Context, db *sql.DB, fn TransactionFunc) error {
	// Start transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is handled properly
	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				// Log rollback error but don't override the panic
				fmt.Printf("failed to rollback transaction after panic: %v\n", rollbackErr)
			}
			panic(p) // Re-panic
		}
	}()

	// Execute the function within the transaction
	if err := fn(tx); err != nil {
		// Rollback on error
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction failed: %v, rollback failed: %w", err, rollbackErr)
		}
		return fmt.Errorf("transaction failed: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TransactionOptions represents options for transaction configuration
type TransactionOptions struct {
	Isolation sql.IsolationLevel
	ReadOnly  bool
}

// WithTransactionOptions executes the given function within a database transaction with custom options
func WithTransactionOptions(ctx context.Context, db *sql.DB, opts *TransactionOptions, fn TransactionFunc) error {
	var txOpts *sql.TxOptions
	if opts != nil {
		txOpts = &sql.TxOptions{
			Isolation: opts.Isolation,
			ReadOnly:  opts.ReadOnly,
		}
	}

	// Start transaction with options
	tx, err := db.BeginTx(ctx, txOpts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is handled properly
	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				fmt.Printf("failed to rollback transaction after panic: %v\n", rollbackErr)
			}
			panic(p) // Re-panic
		}
	}()

	// Execute the function within the transaction
	if err := fn(tx); err != nil {
		// Rollback on error
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("transaction failed: %v, rollback failed: %w", err, rollbackErr)
		}
		return fmt.Errorf("transaction failed: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

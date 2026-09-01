package tx

import (
	"context"

	"github.com/HieuNguyenVV/book-hive/internal/log"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/database"
)

// WithTx is the helper function to execute a function with a transaction
func WithTransaction(ctx context.Context, logger log.Logger, fu func(ctx context.Context) error) error {
	// Create a new transaction context
	trxCtx := new(database.TransactionCtx)
	// Set the transaction context in the context
	ctx = context.WithValue(ctx, database.TransactionCtxKey, trxCtx)
	// Execute the function with the transaction context
	err := fu(ctx)
	// If the function returns an error, rollback the transaction
	if err != nil {
		if rollbackErr := trxCtx.Rollback(); rollbackErr != nil {
			logger.Errorf("failed to rollback transaction: %v", rollbackErr)
		}
		return err
	}
	// Commit the transaction
	if err := trxCtx.Commit(); err != nil {
		if rollbackErr := trxCtx.Rollback(); rollbackErr != nil {
			logger.Errorf("failed to rollback transaction: %v", rollbackErr)
		}
		return err
	}
	return nil
}

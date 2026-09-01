package middleware

import (
	"context"
	"database/sql"
	"strings"

	"github.com/HieuNguyenVV/book-hive/internal/log"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/database"
	"github.com/gin-gonic/gin"
)

const ginTxKey = "tx_ctx"

// Tx starts a request-scoped transaction and commits or rolls it back after the handler finishes.
func Tx(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipTransaction(c.Request.URL.Path) {
			c.Next()
			return
		}

		trx := &database.TransactionCtx{}
		reqCtx := context.WithValue(c.Request.Context(), database.TransactionCtxKey, trx)
		c.Request = c.Request.WithContext(reqCtx)
		c.Set(ginTxKey, trx)

		defer func() {
			if r := recover(); r != nil {
				rollbackTransaction(trx, logger)
				panic(r)
			}
			finalizeTransaction(trx, c, logger)
		}()

		c.Next()
	}
}

func skipTransaction(path string) bool {
	switch {
	case path == "/health", path == "/readyz", path == "/ws":
		return true
	case strings.HasPrefix(path, "/swagger"):
		return true
	default:
		return false
	}
}

func finalizeTransaction(trx *database.TransactionCtx, c *gin.Context, logger log.Logger) {
	if trx == nil || trx.Conn == nil {
		return
	}

	if shouldRollback(c) {
		if err := trx.Rollback(); err != nil && err != sql.ErrTxDone {
			logger.Errorf("failed to rollback transaction: %v", err)
			return
		}
		logger.Info("transaction rolled back")
		return
	}

	if err := trx.Commit(); err != nil {
		if err != sql.ErrTxDone {
			logger.Errorf("failed to commit transaction: %v", err)
			if rollbackErr := trx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
				logger.Errorf("failed to rollback transaction after commit failure: %v", rollbackErr)
			}
		}
		return
	}
	logger.Info("transaction committed")
}

func rollbackTransaction(trx *database.TransactionCtx, logger log.Logger) {
	if trx == nil || trx.Conn == nil {
		return
	}
	if err := trx.Rollback(); err != nil && err != sql.ErrTxDone {
		logger.Errorf("failed to rollback transaction: %v", err)
		return
	}
	logger.Info("transaction rolled back after panic")
}

func shouldRollback(c *gin.Context) bool {
	if c.IsAborted() || c.Writer.Status() >= 400 {
		return true
	}
	return len(c.Errors) > 0
}

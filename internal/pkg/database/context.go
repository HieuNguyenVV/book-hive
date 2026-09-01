package database

import (
	"sync"

	"github.com/jmoiron/sqlx"
)

// TransactionCtxKeyType is the key type for the transaction context
type TransactionCtxKeyType string

// CustomerSettingCtx is the context for the customer setting
// IsJobAfterTxCommit is the flag to indicate if the job should be executed after the transaction is committed
type CustomerSettingCtx struct {
	IsJobAfterTxCommit bool
}

// CustomerSettingCtxKeyType is the key type for the customer setting context
type CustomerSettingCtxKeyType string

const (
	// TransactionCtxKey is the key for the transaction context
	TransactionCtxKey TransactionCtxKeyType = "transaction_ctx"

	// CustomerSettingCtxKey is the key for the customer setting context
	CustomerSettingCtxKey CustomerSettingCtxKeyType = "customer_setting_ctx"
)

type TransactionCtx struct {
	Mu   sync.Mutex
	Conn *sqlx.Tx
}

// Commit the transaction
func (t *TransactionCtx) Commit() error {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	if t.Conn != nil {
		err := t.Conn.Commit()
		if err == nil {
			t.Conn = nil
		}
		return err
	}
	return nil
}

// Rollback the transaction
func (t *TransactionCtx) Rollback() error {
	t.Mu.Lock()
	defer t.Mu.Unlock()
	if t.Conn != nil {
		err := t.Conn.Rollback()
		if err == nil {
			t.Conn = nil
		}
		return err
	}
	return nil
}

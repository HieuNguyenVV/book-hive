package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/HieuNguyenVV/book-hive/internal/pkg/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Postgres is the type for the postgres database
// It contains the master and slave databases
type Postgres struct {
	writeDB *sqlx.DB
	readDB  *sqlx.DB
}

// MasterDB is the type for the master database
type MasterDB *sqlx.DB

// SlaveDB is the type for the slave database
type SlaveDB *sqlx.DB

// Conn is the interface for the database connection
type Conn interface {
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
	DriverName() string
	Get(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	MustExec(query string, args ...interface{}) sql.Result
	MustExecContext(ctx context.Context, query string, args ...interface{}) sql.Result
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
	Preparex(query string) (*sqlx.Stmt, error)
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	Rebind(query string) string
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// NewMasterDB creates a new master database
func NewMasterDB(cfg config.PostgresNodeConfig) (MasterDB, error) {
	dns := cfg.DSN()
	sqlDB, err := sql.Open("postgres", dns)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	db := sqlx.NewDb(sqlDB, "postgres")
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// NewSlaveDB creates a new slave database
func NewSlaveDB(cfg config.PostgresNodeConfig) (SlaveDB, error) {
	dns := cfg.DSN()
	sqlDB, err := sql.Open("postgres", dns)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	db := sqlx.NewDb(sqlDB, "postgres")
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// NewPostgres creates a new postgres database
func NewPostgres(cfg config.PostgresConfig) (*Postgres, error) {
	masterDB, err := NewMasterDB(cfg.Master)
	if err != nil {
		return nil, err
	}
	slaveDB, err := NewSlaveDB(cfg.Slaves)
	if err != nil {
		return nil, err
	}
	return &Postgres{writeDB: masterDB, readDB: slaveDB}, nil
}

// Ping pings the database to check if the connection is successful
func (p *Postgres) Ping() error {
	if p.writeDB != nil {
		if err := p.writeDB.Ping(); err != nil {
			return fmt.Errorf("failed to ping write database: %w", err)
		}
	}
	if p.readDB != nil {
		if err := p.readDB.Ping(); err != nil {
			return fmt.Errorf("failed to ping read database: %w", err)
		}
	}
	return nil
}

// Shutdown shuts down the postgres database
func (p *Postgres) Shutdown() {
	if p.writeDB != nil {
		_ = p.writeDB.Close()
	}
	if p.readDB != nil {
		_ = p.readDB.Close()
	}
}

// MasterDB returns the master database
func (p *Postgres) MasterDB() MasterDB {
	return p.writeDB
}

// SlaveDB returns the slave database
func (p *Postgres) SlaveDB() SlaveDB {
	return p.readDB
}

// GetReadConnection returns the read connection
func (p *Postgres) GetReadConnection(ctx context.Context) (Conn, error) {
	// If the transaction context is set, return the transaction connection
	if transactionCtx, ok := ctx.Value(TransactionCtxKey).(*TransactionCtx); ok {
		conn := transactionCtx.Conn
		if conn != nil {
			return conn, nil
		}
	}
	// If the customer setting context is set and the job should be executed after the transaction is committed, return the write connection
	if customerSettingCtx, ok := ctx.Value(CustomerSettingCtxKey).(*CustomerSettingCtx); ok {
		// If the job should be executed after the transaction is committed, return the write connection
		if customerSettingCtx.IsJobAfterTxCommit {
			return p.writeDB, nil
		}
	}
	return p.readDB, nil
}

// GetWriteConnection returns the write connection
func (p *Postgres) GetWriteConnection(ctx context.Context) (Conn, error) {
	if transactionCtx, ok := ctx.Value(TransactionCtxKey).(*TransactionCtx); ok {
		if transactionCtx.Conn == nil {
			transactionCtx.Mu.Lock()
			defer transactionCtx.Mu.Unlock()

			if transactionCtx.Conn == nil {
				conn, err := p.writeDB.Beginx()
				if err != nil {
					return nil, err
				}
				transactionCtx.Conn = conn
			}
		}
		return transactionCtx.Conn, nil
	}

	return p.writeDB, nil
}

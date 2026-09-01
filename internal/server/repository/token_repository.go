package repository

import (
	"context"
	"database/sql"
	"time"

	appErrors "github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/database"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
)

type TokenRepository interface {
	CreateToken(ctx context.Context, token *model.Token) (*model.Token, error)
	GetTokenByToken(ctx context.Context, token string) (*model.Token, error)
	RevokeToken(ctx context.Context, token string) error
}

type tokenRepository struct {
	db *database.Postgres
}

func NewTokenRepository(db *database.Postgres) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) CreateToken(ctx context.Context, token *model.Token) (*model.Token, error) {
	conn, err := r.db.GetWriteConnection(ctx)
	if err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err).Reform("failed to get write connection")
	}

	query := `
		INSERT INTO tokens (user_id, token, revoked_at, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, token, revoked_at, expires_at, created_at, updated_at
	`
	err = conn.QueryRowxContext(ctx, query, token.UserID, token.Token, token.RevokedAt, token.ExpiresAt, token.CreatedAt, token.UpdatedAt).StructScan(token)
	if err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err).Reform("failed to create token")
	}

	return token, nil
}

func (r *tokenRepository) GetTokenByToken(ctx context.Context, token string) (*model.Token, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err).Reform("failed to get read connection")
	}

	query := `
		SELECT id, user_id, token, revoked_at, expires_at, created_at, updated_at
		FROM tokens
		WHERE token = $1
	`

	tokenModel := &model.Token{}
	err = conn.QueryRowxContext(ctx, query, token).StructScan(tokenModel)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, appErrors.ErrDatabase.Wrap(err).Reform("failed to get token by token: %s", token)
	}
	return tokenModel, nil
}

func (r *tokenRepository) RevokeToken(ctx context.Context, token string) error {
	conn, err := r.db.GetWriteConnection(ctx)
	if err != nil {
		return appErrors.ErrDatabase.Wrap(err).Reform("failed to get write connection")
	}

	now := time.Now().Unix()
	query := `
		UPDATE tokens
		SET revoked_at = ?, updated_at = ?
		WHERE token = ?
	`
	_, err = conn.Exec(conn.Rebind(query), now, now, token)
	if err != nil {
		return appErrors.ErrDatabase.Wrap(err).Reform("failed to revoke token: %s", token)
	}
	return nil
}

package model

type Token struct {
	ID        string `json:"id" db:"id"`
	UserID    string `json:"user_id" db:"user_id"`
	Token     string `json:"token" db:"token"`
	RevokedAt int64  `json:"revoked_at" db:"revoked_at"`
	ExpiresAt int64  `json:"expires_at" db:"expires_at"`
	CreatedAt int64  `json:"created_at" db:"created_at"`
	UpdatedAt int64  `json:"updated_at" db:"updated_at"`
}

package model

type Member struct {
	ID        string  `json:"id" db:"id"`
	ChannelID string  `json:"channel_id" db:"channel_id"`
	UserID    string  `json:"user_id" db:"user_id"`
	Role      string  `json:"role" db:"role"`
	State     string  `json:"state" db:"state"`
	JoinedAt  int64   `json:"joined_at" db:"joined_at"`
	InviterID *string `json:"inviter_id,omitempty" db:"inviter_id"`
	CreatedAt int64   `json:"created_at" db:"created_at"`
	UpdatedAt int64   `json:"updated_at" db:"updated_at"`
}

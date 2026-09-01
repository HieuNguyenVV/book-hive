package model

import "github.com/HieuNguyenVV/book-hive/internal/pkg/database"

const GlobalPubSubTopic = "message"

const (
	ChannelTypeGroup = "group"
	ChannelTypeOpen  = "open"
)

const (
	MemberRoleOperator = "operator"
	MemberRoleMember   = "member"
	MemberStateJoined  = "joined"
)

type Message struct {
	ID          string           `json:"id" db:"id"`
	ChannelID   string           `json:"channel_id" db:"channel_id"`
	SenderID    string           `json:"sender_id" db:"sender_id"`
	Content     string           `json:"content" db:"content"`
	MessageType string           `json:"message_type" db:"message_type"`
	CustomType  string           `json:"custom_type" db:"custom_type"`
	Data        database.JSONMap `json:"data" db:"data"`
	ReqID       *string          `json:"req_id" db:"req_id"`
	CreatedAt   int64            `json:"created_at" db:"created_at"`
	UpdatedAt   int64            `json:"updated_at" db:"updated_at"`
}

type NewMessage struct {
	ChannelID   string    `json:"channel_id"`
	ChannelURL  string    `json:"channel_url"`
	ChannelName string    `json:"channel_name"`
	IsDistinct  bool      `json:"is_distinct"`
	ChannelType string    `json:"channel_type"`
	MsgID       string    `json:"msg_id"`
	Message     string    `json:"message"`
	UserID      string    `json:"user_id"`
	Sender      UserBrief `json:"sender"`
	Ts          int64     `json:"ts"`
	ReqID       *string   `json:"req_id,omitempty"`
}

type CreateDirectChannelRequest struct {
	TargetUserID string `json:"target_user_id"`
}

package repository

import (
	"context"
	"encoding/json"
	"time"

	appErrors "github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/database"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
)

type MessageRepository interface {
	CreateMessage(ctx context.Context, message *model.Message) (*model.Message, error)
	UpdateChannelLastMessage(ctx context.Context, channelID, messageID string) error
	GetMessagesByChannelID(ctx context.Context, channelID string, limit, offset int) ([]*model.MessageWithSender, int, error)
	GetMessageWithSenderByID(ctx context.Context, messageID string) (*model.MessageWithSender, error)
}

type messageRepository struct {
	db *database.Postgres
}

func NewMessageRepository(db *database.Postgres) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) CreateMessage(ctx context.Context, message *model.Message) (*model.Message, error) {
	conn, err := r.db.GetWriteConnection(ctx)
	if err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err)
	}

	dataJSON, err := json.Marshal(defaultMap(message.Data))
	if err != nil {
		return nil, appErrors.ErrInvalidValue.Wrap(err)
	}

	now := time.Now().Unix()
	stmt := `
		INSERT INTO messages (
			channel_id, sender_id, content, message_type, custom_type, data, req_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, channel_id, sender_id, content, message_type, custom_type, data, req_id, created_at, updated_at
	`

	var created model.Message
	if err := conn.QueryRowxContext(ctx, stmt,
		message.ChannelID,
		message.SenderID,
		message.Content,
		message.MessageType,
		message.CustomType,
		dataJSON,
		message.ReqID,
		now,
		now,
	).StructScan(&created); err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err).Reform("failed to create message")
	}
	return &created, nil
}

func (r *messageRepository) UpdateChannelLastMessage(ctx context.Context, channelID, messageID string) error {
	conn, err := r.db.GetWriteConnection(ctx)
	if err != nil {
		return appErrors.ErrDatabase.Wrap(err)
	}

	_, err = conn.Exec(
		conn.Rebind(`UPDATE channels SET last_message_id = ?, update_at = ? WHERE id = ?`),
		messageID,
		time.Now().Unix(),
		channelID,
	)
	if err != nil {
		return appErrors.ErrDatabase.Wrap(err).Reform("failed to update channel last message")
	}
	return nil
}

func defaultMap(data database.JSONMap) database.JSONMap {
	if data == nil {
		return database.JSONMap{}
	}
	return data
}

type messageWithSenderRow struct {
	ID              string  `db:"id"`
	ChannelID       string  `db:"channel_id"`
	Content         string  `db:"content"`
	MessageType     string  `db:"message_type"`
	CustomType      string  `db:"custom_type"`
	ReqID           *string `db:"req_id"`
	CreatedAt       int64   `db:"created_at"`
	UpdatedAt       int64   `db:"updated_at"`
	SenderID        string  `db:"sender_id"`
	SenderFullName  string  `db:"sender_full_name"`
	SenderEmail     string  `db:"sender_email"`
	SenderFirstName string  `db:"sender_first_name"`
	SenderLastName  string  `db:"sender_last_name"`
}

func mapMessageWithSender(row messageWithSenderRow) *model.MessageWithSender {
	return &model.MessageWithSender{
		ID:          row.ID,
		ChannelID:   row.ChannelID,
		Content:     row.Content,
		MessageType: row.MessageType,
		CustomType:  row.CustomType,
		ReqID:       row.ReqID,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		Sender: model.UserBrief{
			ID:        row.SenderID,
			FullName:  row.SenderFullName,
			Email:     row.SenderEmail,
			FirstName: row.SenderFirstName,
			LastName:  row.SenderLastName,
		},
	}
}

func (r *messageRepository) GetMessagesByChannelID(ctx context.Context, channelID string, limit, offset int) ([]*model.MessageWithSender, int, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return nil, 0, appErrors.ErrDatabase.Wrap(err)
	}

	countStmt := conn.Rebind(`SELECT COUNT(*) FROM messages WHERE channel_id = ?`)
	var total int
	if err := conn.QueryRowxContext(ctx, countStmt, channelID).Scan(&total); err != nil {
		return nil, 0, appErrors.ErrDatabase.Wrap(err).Reform("failed to count channel messages")
	}
	if total == 0 {
		return []*model.MessageWithSender{}, 0, nil
	}

	stmt := conn.Rebind(`
		SELECT
			msg.id, msg.channel_id, msg.content, msg.message_type, msg.custom_type, msg.req_id,
			msg.created_at, msg.updated_at,
			u.id AS sender_id,
			u.full_name AS sender_full_name,
			u.email AS sender_email,
			u.first_name AS sender_first_name,
			u.last_name AS sender_last_name
		FROM messages msg
		INNER JOIN users u ON u.id = msg.sender_id
		WHERE msg.channel_id = ?
		ORDER BY msg.created_at DESC
		LIMIT ? OFFSET ?
	`)

	rows := make([]messageWithSenderRow, 0)
	if err := conn.SelectContext(ctx, &rows, stmt, channelID, limit, offset); err != nil {
		return nil, 0, appErrors.ErrDatabase.Wrap(err).Reform("failed to get channel messages")
	}

	messages := make([]*model.MessageWithSender, len(rows))
	for i, row := range rows {
		messages[i] = mapMessageWithSender(row)
	}
	return messages, total, nil
}

func (r *messageRepository) GetMessageWithSenderByID(ctx context.Context, messageID string) (*model.MessageWithSender, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err)
	}

	stmt := conn.Rebind(`
		SELECT
			msg.id, msg.channel_id, msg.content, msg.message_type, msg.custom_type, msg.req_id,
			msg.created_at, msg.updated_at,
			u.id AS sender_id,
			u.full_name AS sender_full_name,
			u.email AS sender_email,
			u.first_name AS sender_first_name,
			u.last_name AS sender_last_name
		FROM messages msg
		INNER JOIN users u ON u.id = msg.sender_id
		WHERE msg.id = ?
	`)

	var row messageWithSenderRow
	if err := conn.GetContext(ctx, &row, stmt, messageID); err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err).Reform("failed to get message")
	}
	return mapMessageWithSender(row), nil
}

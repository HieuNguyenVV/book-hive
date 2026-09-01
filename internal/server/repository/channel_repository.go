package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/database"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
)

type ChannelRepository interface {
	CreateChannel(ctx context.Context, channel *model.Channel) (*model.Channel, error)
	GetChannelByID(ctx context.Context, id string) (*model.Channel, error)
	GetChannels(ctx context.Context, request *model.ListChannelsRequest) ([]*model.Channel, int, error)
	GetChannelsByUserID(ctx context.Context, userID string) ([]*model.Channel, error)
	GetJoinedChannelsByUserID(ctx context.Context, userID string, request *model.ListChannelsRequest, limit, offset int) ([]*model.Channel, int, error)
	GetChannelByURL(ctx context.Context, url string) (*model.Channel, error)
	FindDistinctDirectChannel(ctx context.Context, userAID, userBID string) (*model.Channel, error)
	UpdateChannel(ctx context.Context, id string, request *model.UpdateChannelRequest, updateBy string) (*model.Channel, error)
	DeleteChannel(ctx context.Context, id string) error
	ChannelURLExists(ctx context.Context, url string, excludeID string) (bool, error)
}

type channelRepository struct {
	db *database.Postgres
}

func NewChannelRepository(db *database.Postgres) ChannelRepository {
	return &channelRepository{db: db}
}

func (r *channelRepository) GetChannelByID(ctx context.Context, id string) (*model.Channel, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return nil, errors.ErrDatabase.Wrap(err)
	}

	stmt := conn.Rebind(`
		SELECT
			c.id,
			c.channel_url,
			c.cover_url,
			c.name,
			c.channel_type,
			c.metadata,
			c.last_message_id,
			c.last_read_message_id,
			c.is_public,
			c.is_distinct,
			c.create_by,
			c.create_at,
			c.update_by,
			c.update_at
		FROM
			channels c
		WHERE
			c.id = ?
	`)
	var channel model.Channel
	if err := conn.GetContext(ctx, &channel, stmt, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound.Reform("channel not found: %s", id)
		}
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to get channel: %s", id)
	}
	return &channel, nil
}

func (r *channelRepository) GetChannelByURL(ctx context.Context, url string) (*model.Channel, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return nil, errors.ErrDatabase.Wrap(err)
	}

	stmt := conn.Rebind(`
		SELECT
			c.id,
			c.channel_url,
			c.cover_url,
			c.name,
			c.channel_type,
			c.metadata,
			c.last_message_id,
			c.last_read_message_id,
			c.is_public,
			c.is_distinct,
			c.create_by,
			c.create_at,
			c.update_by,
			c.update_at
		FROM
			channels c
		WHERE
			c.channel_url = ?
	`)
	var channel model.Channel
	if err := conn.GetContext(ctx, &channel, stmt, url); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound.Reform("channel not found: %s", url)
		}
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to get channel: %s", url)
	}
	return &channel, nil
}

func (r *channelRepository) CreateChannel(ctx context.Context, channel *model.Channel) (*model.Channel, error) {
	conn, err := r.db.GetWriteConnection(ctx)
	if err != nil {
		return nil, errors.ErrDatabase.Wrap(err)
	}

	stmt := `
		INSERT INTO channels (
			channel_url,
			cover_url,
			name,
			channel_type,
			metadata,
			last_message_id,
			last_read_message_id,
			is_public,
			is_distinct,
			create_by,
			create_at,
			update_by,
			update_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, channel_url, cover_url, name, channel_type, metadata,
		last_message_id, last_read_message_id, is_public, is_distinct, create_by, create_at, update_by, update_at
	`
	var createdChannel model.Channel
	metadata := channel.Metadata
	if metadata == nil {
		metadata = database.JSONMap{}
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, errors.ErrInvalidValue.Wrap(err).Reform("invalid metadata")
	}

	if err := conn.QueryRowxContext(ctx, stmt,
		channel.ChannelURL,
		nullIfEmptyPtr(channel.CoverURL),
		channel.Name,
		channel.ChannelType,
		metadataJSON,
		nullIfEmptyPtr(channel.LastMessageID),
		nullIfEmptyPtr(channel.LastReadMessageID),
		channel.IsPublic,
		channel.IsDistinct,
		channel.CreateBy,
		channel.CreateAt,
		channel.UpdateBy,
		channel.UpdateAt,
	).StructScan(&createdChannel); err != nil {
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to create channel")
	}
	return &createdChannel, nil
}

func (r *channelRepository) GetChannels(ctx context.Context, request *model.ListChannelsRequest) ([]*model.Channel, int, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return nil, 0, errors.ErrDatabase.Wrap(err)
	}

	args := make([]any, 0)
	conditions := make([]string, 0)

	if request.Search != nil && *request.Search != "" {
		search := "%" + *request.Search + "%"
		conditions = append(conditions, "(name LIKE ? OR channel_url LIKE ?)")
		args = append(args, search, search)
	}
	if request.ChannelType != nil {
		conditions = append(conditions, "channel_type = ?")
		args = append(args, *request.ChannelType)
	}
	if request.IsPublic != nil {
		conditions = append(conditions, "is_public = ?")
		args = append(args, *request.IsPublic)
	}
	if request.ChannelURL != nil && *request.ChannelURL != "" {
		conditions = append(conditions, "channel_url = ?")
		args = append(args, *request.ChannelURL)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	limitOffset := ""
	if request.Limit != nil {
		limitOffset += fmt.Sprintf(" LIMIT %d", *request.Limit)
	}
	if request.Offset != nil {
		limitOffset += fmt.Sprintf(" OFFSET %d", *request.Offset)
	}

	cte := `
		WITH filtered AS (
			SELECT
				id, channel_url, cover_url, name, channel_type, metadata,
				last_message_id, last_read_message_id, is_public, is_distinct,
				create_by, create_at, update_by, update_at
			FROM channels` + whereClause + `
		),
		total AS (
			SELECT COUNT(*) AS total_count FROM filtered
		)
	`

	stmt := cte + `
		SELECT
			f.id, f.channel_url, f.cover_url, f.name, f.channel_type, f.metadata,
			f.last_message_id, f.last_read_message_id, f.is_public, f.is_distinct,
			f.create_by, f.create_at, f.update_by, f.update_at,
			(SELECT total_count FROM total) AS total_count
		FROM filtered f
		ORDER BY f.create_at DESC` + limitOffset

	type channelRow struct {
		model.Channel
		TotalCount int `db:"total_count"`
	}

	rows := make([]channelRow, 0)
	if err := conn.Select(&rows, conn.Rebind(stmt), args...); err != nil {
		return nil, 0, errors.ErrDatabase.Wrap(err).Reform("failed to get channels")
	}
	if len(rows) == 0 {
		return []*model.Channel{}, 0, nil
	}

	channels := make([]*model.Channel, len(rows))
	for i, row := range rows {
		channel := row.Channel
		channels[i] = &channel
	}
	return channels, rows[0].TotalCount, nil
}

func (r *channelRepository) GetChannelsByUserID(ctx context.Context, userID string) ([]*model.Channel, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return nil, errors.ErrDatabase.Wrap(err)
	}

	stmt := conn.Rebind(`
		SELECT
			c.id,
			c.channel_url,
			c.cover_url,
			c.name,	
			c.channel_type,
			c.metadata,
			c.last_message_id,
			c.last_read_message_id,
			c.is_public,
			c.is_distinct,
			c.create_by,
			c.create_at,
			c.update_by,
			c.update_at
		FROM
			channels c
		WHERE
			c.create_by = ?
	`)
	var channels []*model.Channel
	if err := conn.SelectContext(ctx, &channels, conn.Rebind(stmt), userID); err != nil {
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to get channel by user id: %s", userID)
	}
	return channels, nil
}

func (r *channelRepository) GetJoinedChannelsByUserID(ctx context.Context, userID string, request *model.ListChannelsRequest, limit, offset int) ([]*model.Channel, int, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return nil, 0, errors.ErrDatabase.Wrap(err)
	}

	args := []any{userID, model.MemberStateJoined}
	conditions := make([]string, 0)

	if request != nil {
		if request.Search != nil && *request.Search != "" {
			search := "%" + *request.Search + "%"
			conditions = append(conditions, "(c.name LIKE ? OR c.channel_url LIKE ?)")
			args = append(args, search, search)
		}
		if request.ChannelType != nil {
			conditions = append(conditions, "c.channel_type = ?")
			args = append(args, *request.ChannelType)
		}
		if request.IsPublic != nil {
			conditions = append(conditions, "c.is_public = ?")
			args = append(args, *request.IsPublic)
		}
		if request.ChannelURL != nil && *request.ChannelURL != "" {
			conditions = append(conditions, "c.channel_url = ?")
			args = append(args, *request.ChannelURL)
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	filteredArgs := append([]any{}, args...)
	filteredArgs = append(filteredArgs, limit, offset)

	cte := `
		WITH filtered AS (
			SELECT
				c.id, c.channel_url, c.cover_url, c.name, c.channel_type, c.metadata,
				c.last_message_id, c.last_read_message_id, c.is_public, c.is_distinct,
				c.create_by, c.create_at, c.update_by, c.update_at
			FROM channels c
			INNER JOIN members m ON m.channel_id = c.id AND m.user_id = ? AND m.state = ?` + whereClause + `
			ORDER BY c.update_at DESC
			LIMIT ? OFFSET ?
		),
		total AS (
			SELECT COUNT(*) AS total_count
			FROM channels c
			INNER JOIN members m ON m.channel_id = c.id AND m.user_id = ? AND m.state = ?` + whereClause + `
		)
		SELECT
			f.id, f.channel_url, f.cover_url, f.name, f.channel_type, f.metadata,
			f.last_message_id, f.last_read_message_id, f.is_public, f.is_distinct,
			f.create_by, f.create_at, f.update_by, f.update_at,
			(SELECT total_count FROM total) AS total_count
		FROM filtered f
	`

	totalArgs := append([]any{}, args...)
	queryArgs := append(filteredArgs, totalArgs...)

	type channelRow struct {
		model.Channel
		TotalCount int `db:"total_count"`
	}

	rows := make([]channelRow, 0)
	if err := conn.SelectContext(ctx, &rows, conn.Rebind(cte), queryArgs...); err != nil {
		return nil, 0, errors.ErrDatabase.Wrap(err).Reform("failed to get joined channels")
	}
	if len(rows) == 0 {
		return []*model.Channel{}, 0, nil
	}

	channels := make([]*model.Channel, len(rows))
	for i, row := range rows {
		channel := row.Channel
		channels[i] = &channel
	}
	return channels, rows[0].TotalCount, nil
}

func (r *channelRepository) UpdateChannel(ctx context.Context, id string, request *model.UpdateChannelRequest, updateBy string) (*model.Channel, error) {
	conn, err := r.db.GetWriteConnection(ctx)
	if err != nil {
		return nil, errors.ErrDatabase.Wrap(err)
	}

	keys := make([]string, 0)
	values := make([]any, 0)

	keys = append(keys, "update_by", "update_at")
	values = append(values, updateBy, time.Now().Unix())

	if request.CoverURL != nil {
		keys = append(keys, "cover_url")
		values = append(values, *request.CoverURL)
	}
	if request.Name != nil {
		keys = append(keys, "name")
		values = append(values, *request.Name)
	}
	if request.ChannelType != nil {
		keys = append(keys, "channel_type")
		values = append(values, *request.ChannelType)
	}
	if request.Metadata != nil {
		metadataJSON, err := json.Marshal(request.Metadata)
		if err != nil {
			return nil, errors.ErrInvalidValue.Wrap(err).Reform("invalid metadata")
		}
		keys = append(keys, "metadata")
		values = append(values, metadataJSON)
	}
	if request.LastMessageID != nil {
		keys = append(keys, "last_message_id")
		values = append(values, nullIfEmpty(*request.LastMessageID))
	}
	if request.LastReadMessageID != nil {
		keys = append(keys, "last_read_message_id")
		values = append(values, nullIfEmpty(*request.LastReadMessageID))
	}
	if request.IsPublic != nil {
		keys = append(keys, "is_public")
		values = append(values, *request.IsPublic)
	}

	setParts := make([]string, len(keys))
	for i, key := range keys {
		setParts[i] = key + " = ?"
	}

	stmt := fmt.Sprintf(`
		UPDATE channels
		SET %s
		WHERE id = ?
		RETURNING id, channel_url, cover_url, name, channel_type, metadata,
			last_message_id, last_read_message_id, is_public, is_distinct, create_by, create_at, update_by, update_at
	`, strings.Join(setParts, ", "))

	values = append(values, id)

	channel := &model.Channel{}
	if err := conn.QueryRowxContext(ctx, conn.Rebind(stmt), values...).StructScan(channel); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound.Reform("channel not found: %s", id)
		}
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to update channel: %s", id)
	}
	return channel, nil
}

func (r *channelRepository) DeleteChannel(ctx context.Context, id string) error {
	conn, err := r.db.GetWriteConnection(ctx)
	if err != nil {
		return errors.ErrDatabase.Wrap(err)
	}

	if _, err = conn.Exec(conn.Rebind(`DELETE FROM messages WHERE channel_id = ?`), id); err != nil {
		return errors.ErrDatabase.Wrap(err).Reform("failed to delete channel messages: %s", id)
	}

	if _, err = conn.Exec(conn.Rebind(`DELETE FROM members WHERE channel_id = ?`), id); err != nil {
		return errors.ErrDatabase.Wrap(err).Reform("failed to delete channel members: %s", id)
	}

	result, err := conn.Exec(conn.Rebind(`DELETE FROM channels WHERE id = ?`), id)
	if err != nil {
		return errors.ErrDatabase.Wrap(err).Reform("failed to delete channel: %s", id)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.ErrDatabase.Wrap(err).Reform("failed to delete channel: %s", id)
	}
	if rowsAffected == 0 {
		return errors.ErrNotFound.Reform("channel not found: %s", id)
	}
	return nil
}

func (r *channelRepository) ChannelURLExists(ctx context.Context, url string, excludeID string) (bool, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return false, errors.ErrDatabase.Wrap(err)
	}

	stmt := `SELECT EXISTS(SELECT 1 FROM channels WHERE channel_url = ?`
	args := []any{url}
	if excludeID != "" {
		stmt += ` AND id <> ?`
		args = append(args, excludeID)
	}
	stmt += `)`

	var exists bool
	if err := conn.QueryRowxContext(ctx, conn.Rebind(stmt), args...).Scan(&exists); err != nil {
		return false, errors.ErrDatabase.Wrap(err).Reform("failed to check channel url")
	}
	return exists, nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullIfEmptyPtr(value *string) any {
	if value == nil || *value == "" {
		return nil
	}
	return *value
}

func (r *channelRepository) FindDistinctDirectChannel(ctx context.Context, userAID, userBID string) (*model.Channel, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return nil, errors.ErrDatabase.Wrap(err)
	}

	stmt := conn.Rebind(`
		SELECT
			c.id, c.channel_url, c.cover_url, c.name, c.channel_type, c.metadata,
			c.last_message_id, c.last_read_message_id, c.is_public, c.is_distinct,
			c.create_by, c.create_at, c.update_by, c.update_at
		FROM channels c
		INNER JOIN members m1 ON m1.channel_id = c.id AND m1.user_id = ? AND m1.state = ?
		INNER JOIN members m2 ON m2.channel_id = c.id AND m2.user_id = ? AND m2.state = ?
		WHERE c.is_distinct = true AND c.channel_type = ?
		LIMIT 1
	`)

	var channel model.Channel
	err = conn.GetContext(ctx, &channel, stmt, userAID, model.MemberStateJoined, userBID, model.MemberStateJoined, model.ChannelTypeGroup)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to find direct channel")
	}
	return &channel, nil
}

package repository

import (
	"context"
	"time"

	appErrors "github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/database"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
	"github.com/jmoiron/sqlx"
)

type MemberRepository interface {
	CreateMember(ctx context.Context, member *model.Member) (*model.Member, error)
	IsMember(ctx context.Context, channelID, userID string) (bool, error)
	GetJoinedChannelIDsByUserID(ctx context.Context, userID string) ([]string, error)
	GetMembersWithUserByChannelID(ctx context.Context, channelID string) ([]*model.ChannelMemberDetail, error)
	GetMembersWithUserByChannelIDs(ctx context.Context, channelIDs []string) (map[string][]*model.ChannelMemberDetail, error)
	CountMembers(ctx context.Context, channelID string) (int, error)
	CountJoinedChannelsByUserID(ctx context.Context, userID string) (int, error)
}

type memberRepository struct {
	db *database.Postgres
}

func NewMemberRepository(db *database.Postgres) MemberRepository {
	return &memberRepository{db: db}
}

func (r *memberRepository) CreateMember(ctx context.Context, member *model.Member) (*model.Member, error) {
	conn, err := r.db.GetWriteConnection(ctx)
	if err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err)
	}

	now := time.Now().Unix()
	stmt := `
		INSERT INTO members (
			channel_id, user_id, role, state, joined_at, inviter_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, channel_id, user_id, role, state, joined_at, inviter_id, created_at, updated_at
	`

	var created model.Member
	if err := conn.QueryRowxContext(ctx, stmt,
		member.ChannelID,
		member.UserID,
		member.Role,
		member.State,
		now,
		nullIfEmptyPtr(member.InviterID),
		now,
		now,
	).StructScan(&created); err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err).Reform("failed to create member")
	}
	return &created, nil
}

func (r *memberRepository) IsMember(ctx context.Context, channelID, userID string) (bool, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return false, appErrors.ErrDatabase.Wrap(err)
	}

	stmt := `
		SELECT EXISTS(
			SELECT 1 FROM members
			WHERE channel_id = ? AND user_id = ? AND state = ?
		)
	`
	var exists bool
	if err := conn.QueryRowxContext(ctx, conn.Rebind(stmt), channelID, userID, model.MemberStateJoined).Scan(&exists); err != nil {
		return false, appErrors.ErrDatabase.Wrap(err)
	}
	return exists, nil
}

func (r *memberRepository) GetJoinedChannelIDsByUserID(ctx context.Context, userID string) ([]string, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err)
	}

	stmt := conn.Rebind(`
		SELECT channel_id FROM members
		WHERE user_id = ? AND state = ?
		ORDER BY joined_at DESC
	`)

	ids := make([]string, 0)
	if err := conn.SelectContext(ctx, &ids, stmt, userID, model.MemberStateJoined); err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err).Reform("failed to get joined channels")
	}
	return ids, nil
}

func (r *memberRepository) CountMembers(ctx context.Context, channelID string) (int, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return 0, appErrors.ErrDatabase.Wrap(err)
	}

	var count int
	if err := conn.QueryRowxContext(ctx, conn.Rebind(`
		SELECT COUNT(*) FROM members WHERE channel_id = ? AND state = ?
	`), channelID, model.MemberStateJoined).Scan(&count); err != nil {
		return 0, appErrors.ErrDatabase.Wrap(err)
	}
	return count, nil
}

func (r *memberRepository) GetMembersWithUserByChannelID(ctx context.Context, channelID string) ([]*model.ChannelMemberDetail, error) {
	members, err := r.GetMembersWithUserByChannelIDs(ctx, []string{channelID})
	if err != nil {
		return nil, err
	}
	return members[channelID], nil
}

func (r *memberRepository) GetMembersWithUserByChannelIDs(ctx context.Context, channelIDs []string) (map[string][]*model.ChannelMemberDetail, error) {
	result := make(map[string][]*model.ChannelMemberDetail)
	if len(channelIDs) == 0 {
		return result, nil
	}

	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err)
	}

	query, args, err := sqlx.In(`
		SELECT
			m.id, m.channel_id, m.user_id, m.role, m.state, m.joined_at,
			u.full_name AS user_full_name,
			u.email AS user_email,
			u.first_name AS user_first_name,
			u.last_name AS user_last_name
		FROM members m
		INNER JOIN users u ON u.id = m.user_id
		WHERE m.channel_id IN (?) AND m.state = ?
		ORDER BY m.joined_at ASC
	`, channelIDs, model.MemberStateJoined)
	if err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err).Reform("failed to bind members query")
	}
	query = conn.Rebind(query)

	type memberRow struct {
		ID            string `db:"id"`
		ChannelID     string `db:"channel_id"`
		UserID        string `db:"user_id"`
		Role          string `db:"role"`
		State         string `db:"state"`
		JoinedAt      int64  `db:"joined_at"`
		UserFullName  string `db:"user_full_name"`
		UserEmail     string `db:"user_email"`
		UserFirstName string `db:"user_first_name"`
		UserLastName  string `db:"user_last_name"`
	}

	rows := make([]memberRow, 0)
	if err := conn.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err).Reform("failed to get channel members")
	}

	for _, row := range rows {
		result[row.ChannelID] = append(result[row.ChannelID], &model.ChannelMemberDetail{
			ID:        row.ID,
			ChannelID: row.ChannelID,
			UserID:    row.UserID,
			Role:      row.Role,
			State:     row.State,
			JoinedAt:  row.JoinedAt,
			User: model.UserBrief{
				ID:        row.UserID,
				FullName:  row.UserFullName,
				Email:     row.UserEmail,
				FirstName: row.UserFirstName,
				LastName:  row.UserLastName,
			},
		})
	}
	return result, nil
}

func (r *memberRepository) CountJoinedChannelsByUserID(ctx context.Context, userID string) (int, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return 0, appErrors.ErrDatabase.Wrap(err)
	}

	var count int
	if err := conn.QueryRowxContext(ctx, conn.Rebind(`
		SELECT COUNT(*) FROM members WHERE user_id = ? AND state = ?
	`), userID, model.MemberStateJoined).Scan(&count); err != nil {
		return 0, appErrors.ErrDatabase.Wrap(err)
	}
	return count, nil
}

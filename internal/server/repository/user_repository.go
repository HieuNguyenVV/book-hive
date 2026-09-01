package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	appErrors "github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/database"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUsers(ctx context.Context, request *model.ListUsersRequest) ([]*model.User, int, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) (*model.User, error)
	CheckUserExists(ctx context.Context, email string) (bool, error)
	UpdateUser(ctx context.Context, id string, user *model.UpdateUserRequest) (*model.User, error)
}

type userRepository struct {
	db *database.Postgres
}

func NewUserRepository(db *database.Postgres) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	if id == "" {
		return nil, appErrors.ErrInvalidArgument.Wrap(errors.New("id is required"))
	}

	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err)
	}

	query := `
		SELECT id, first_name, last_name, full_name, email, password, role, status, 
		COALESCE(last_login_at, 0) AS last_login_at, deleted_at, created_at, updated_at 
		FROM users 
		WHERE id = $1
	`
	user := &model.User{}
	err = conn.QueryRowxContext(ctx, query, id).StructScan(user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, appErrors.ErrDatabase.Wrap(err).Reform("failed to get user by id: %s", id)
	}
	return user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	if email == "" {
		return nil, appErrors.ErrInvalidArgument.Wrap(errors.New("email is required"))
	}

	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err)
	}

	query := `
		SELECT id, first_name, last_name, full_name, email, password, role, status,
		COALESCE(last_login_at, 0) AS last_login_at, deleted_at, created_at, updated_at 
		FROM users 
		WHERE email = $1
	`
	user := &model.User{}
	err = conn.QueryRowxContext(ctx, query, email).StructScan(user)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, appErrors.ErrDatabase.Wrap(err).Reform("failed to get user by email: %s", email)
	}
	return user, nil
}

func (r *userRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	conn, err := r.db.GetWriteConnection(ctx)
	if err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err)
	}

	query := `
		INSERT INTO users (first_name, last_name, full_name, email, password, role, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, first_name, last_name, full_name, email, password, role, status,
		COALESCE(last_login_at, 0) AS last_login_at, deleted_at, created_at, updated_at
	`
	err = conn.QueryRowxContext(
		ctx,
		query,
		user.FirstName,
		user.LastName,
		user.FullName,
		user.Email,
		user.Password,
		user.Role,
		user.Status,
	).StructScan(user)
	if err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err).Reform("failed to create user")
	}
	return user, nil
}

func (r *userRepository) CheckUserExists(ctx context.Context, email string) (bool, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return false, appErrors.ErrDatabase.Wrap(err)
	}

	query := `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`
	var exists bool
	err = conn.QueryRowxContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, appErrors.ErrDatabase.Wrap(err).Reform("failed to check user exists: %s", email)
	}
	return exists, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, id string, user *model.UpdateUserRequest) (*model.User, error) {
	conn, err := r.db.GetWriteConnection(ctx)
	if err != nil {
		return nil, appErrors.ErrDatabase.Wrap(err)
	}

	keys := make([]string, 0)
	values := make([]interface{}, 0)

	keys = append(keys, "updated_at")
	values = append(values, time.Now().Unix())

	if user.FirstName != nil {
		keys = append(keys, "first_name")
		values = append(values, *user.FirstName)
	}
	if user.LastName != nil {
		keys = append(keys, "last_name")
		values = append(values, *user.LastName)
	}
	if user.FullName != nil {
		keys = append(keys, "full_name")
		values = append(values, *user.FullName)
	}
	if user.Role != nil {
		keys = append(keys, "role")
		values = append(values, *user.Role)
	}
	if user.Status != nil {
		keys = append(keys, "status")
		values = append(values, *user.Status)
	}
	if user.Password != nil {
		keys = append(keys, "password")
		values = append(values, *user.Password)
	}
	if user.LastLoginAt != nil {
		keys = append(keys, "last_login_at")
		values = append(values, *user.LastLoginAt)
	}

	setParts := make([]string, len(keys))
	for i, key := range keys {
		setParts[i] = key + "=?"
	}

	userModel := &model.User{}
	statement := fmt.Sprintf(
		"UPDATE users SET %s WHERE id = ? RETURNING id, first_name, last_name, full_name, email, password, role, status, COALESCE(last_login_at, 0) AS last_login_at, deleted_at, created_at, updated_at",
		strings.Join(setParts, ", "),
	)
	values = append(values, id)
	err = conn.QueryRowxContext(ctx, conn.Rebind(statement), values...).StructScan(userModel)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, appErrors.ErrNotFound.Reform("user not found: %s", id)
		}
		return nil, appErrors.ErrDatabase.Wrap(err).Reform("failed to update user: %s", id)
	}
	return userModel, nil
}

func (r *userRepository) GetUsers(ctx context.Context, request *model.ListUsersRequest) ([]*model.User, int, error) {
	conn, err := r.db.GetReadConnection(ctx)
	if err != nil {
		return nil, 0, appErrors.ErrDatabase.Wrap(err)
	}

	args := make([]any, 0)
	conditions := make([]string, 0)

	if request.Search != nil && *request.Search != "" {
		search := "%" + *request.Search + "%"
		conditions = append(conditions, "(first_name LIKE ? OR last_name LIKE ? OR full_name LIKE ? OR email LIKE ?)")
		args = append(args, search, search, search, search)
	}
	if request.Role != nil {
		conditions = append(conditions, "role = ?")
		args = append(args, *request.Role)
	}
	if request.Status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, *request.Status)
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
				id, first_name, last_name, full_name, email, password, role, status,
				COALESCE(last_login_at, 0) AS last_login_at, deleted_at, created_at, updated_at
			FROM users` + whereClause + `
		),
		total AS (
			SELECT COUNT(*) AS total_count FROM filtered
		)
	`

	stmt := cte + `
		SELECT
			f.id, f.first_name, f.last_name, f.full_name, f.email, f.password, f.role, f.status,
			f.last_login_at, f.deleted_at, f.created_at, f.updated_at,
			(SELECT total_count FROM total) AS total_count
		FROM filtered f
		ORDER BY f.created_at DESC` + limitOffset

	type userRow struct {
		model.User
		TotalCount int `db:"total_count"`
	}

	rows := make([]userRow, 0)
	if err := conn.Select(&rows, conn.Rebind(stmt), args...); err != nil {
		return nil, 0, appErrors.ErrDatabase.Wrap(err).Reform("failed to get users")
	}
	if len(rows) == 0 {
		return []*model.User{}, 0, nil
	}

	users := make([]*model.User, len(rows))
	for i, row := range rows {
		user := row.User
		users[i] = &user
	}
	return users, rows[0].TotalCount, nil
}

package model

type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
	UserRoleGuest UserRole = "guest"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusPending  UserStatus = "pending"
	UserStatusBlocked  UserStatus = "blocked"
	UserStatusDeleted  UserStatus = "deleted"
)

type User struct {
	ID          string     `json:"id" db:"id"`
	FirstName   string     `json:"first_name" db:"first_name"`
	LastName    string     `json:"last_name" db:"last_name"`
	FullName    string     `json:"full_name" db:"full_name"`
	Email       string     `json:"email" db:"email"`
	Password    string     `json:"-" db:"password"`
	Role        UserRole   `json:"role" db:"role"`
	Status      UserStatus `json:"status" db:"status"`
	LastLoginAt int64      `json:"last_login_at" db:"last_login_at"`
	DeletedAt   int64      `json:"deleted_at" db:"deleted_at"`
	CreatedAt   int64      `json:"created_at" db:"created_at"`
	UpdatedAt   int64      `json:"updated_at" db:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type UpdateUserRequest struct {
	FirstName   *string     `json:"first_name"`
	LastName    *string     `json:"last_name"`
	FullName    *string     `json:"full_name"`
	Role        *UserRole   `json:"role"`
	Status      *UserStatus `json:"status"`
	Password    *string     `json:"password"`
	LastLoginAt *int64      `json:"last_login_at"`
}

type ListUsersRequest struct {
	Search *string     `json:"search"`
	Role   *UserRole   `json:"role"`
	Status *UserStatus `json:"status"`
	Limit  *int        `json:"limit"`
	Offset *int        `json:"offset"`
}

type ListUsersResponse struct {
	Results  []*User `json:"results"`
	Total    int     `json:"total"`
	Previous string  `json:"previous"`
	Next     string  `json:"next"`
}

type LoginWebSocketResponse struct {
	UserID       string     `json:"user_id"`
	LoginTs      int64      `json:"login_ts"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	FullName     string     `json:"full_name"`
	Email        string     `json:"email"`
	Role         UserRole   `json:"role"`
	Status       UserStatus `json:"status"`
	LastLoginAt  int64      `json:"last_login_at"`
	Key          string     `json:"key"`
	PingInterval int32      `json:"ping_interval"`
	PongTimeout  int32      `json:"pong_timeout"`
}

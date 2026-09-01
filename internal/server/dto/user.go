package dto

import "github.com/HieuNguyenVV/book-hive/internal/server/model"

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
} // @name LoginRequest

func (r *LoginRequest) ToModel() model.LoginRequest {
	return model.LoginRequest{
		Email:    r.Email,
		Password: r.Password,
	}
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
} // @name LoginResponse

func (r *LoginResponse) FromModel(model *model.LoginResponse) {
	r.AccessToken = model.AccessToken
	r.RefreshToken = model.RefreshToken
	r.ExpiresIn = model.ExpiresIn
	r.TokenType = model.TokenType
}

type RegisterRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	FullName  string `json:"full_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,password"`
} // @name RegisterRequest

func (r *RegisterRequest) ToModel() model.RegisterRequest {
	return model.RegisterRequest{
		FirstName: r.FirstName,
		LastName:  r.LastName,
		FullName:  r.FullName,
		Email:     r.Email,
		Password:  r.Password,
	}
}

type UserResponse struct {
	ID          string           `json:"id"`
	FirstName   string           `json:"first_name"`
	LastName    string           `json:"last_name"`
	FullName    string           `json:"full_name"`
	Email       string           `json:"email"`
	Role        model.UserRole   `json:"role"`
	Status      model.UserStatus `json:"status"`
	LastLoginAt int64            `json:"last_login_at"`
	CreatedAt   int64            `json:"created_at"`
	UpdatedAt   int64            `json:"updated_at"`
} // @name UserResponse

func (r *UserResponse) FromModel(model *model.User) {
	r.ID = model.ID
	r.FirstName = model.FirstName
	r.LastName = model.LastName
	r.FullName = model.FullName
	r.Email = model.Email
	r.Role = model.Role
	r.Status = model.Status
	r.LastLoginAt = model.LastLoginAt
	r.CreatedAt = model.CreatedAt
	r.UpdatedAt = model.UpdatedAt
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,password"`
	ConfirmPassword string `json:"confirm_password" binding:"required,password"`
} // @name ChangePasswordRequest

func (r *ChangePasswordRequest) ToModel() model.ChangePasswordRequest {
	return model.ChangePasswordRequest{
		OldPassword:     r.OldPassword,
		NewPassword:     r.NewPassword,
		ConfirmPassword: r.ConfirmPassword,
	}
}

type UpdateUserRequest struct {
	FirstName *string           `json:"first_name" binding:"required"`
	LastName  *string           `json:"last_name" binding:"required"`
	FullName  *string           `json:"full_name" binding:"required"`
	Role      *model.UserRole   `json:"role" binding:"required"`
	Status    *model.UserStatus `json:"status" binding:"required"`
} // @name UpdateUserRequest

func (r *UpdateUserRequest) ToModel() model.UpdateUserRequest {
	return model.UpdateUserRequest{
		FirstName: r.FirstName,
		LastName:  r.LastName,
		FullName:  r.FullName,
		Role:      r.Role,
		Status:    r.Status,
	}
}

type ListUsersRequest struct {
	Token  *string           `form:"token"`
	Search *string           `form:"search"`
	Role   *model.UserRole   `form:"role"`
	Status *model.UserStatus `form:"status"`
	Limit  *int              `form:"limit"`
	Offset *int              `form:"offset"`
} // @name ListUsersRequest

type ListUsersResponse struct {
	Results  []*model.User `json:"results"`
	Total    int           `json:"total"`
	Previous string        `json:"previous"`
	Next     string        `json:"next"`
} // @name ListUsersResponse

func (r *ListUsersResponse) FromModel(model *model.ListUsersResponse) {
	r.Results = model.Results
	r.Total = model.Total
	r.Previous = model.Previous
	r.Next = model.Next
}

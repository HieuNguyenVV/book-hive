package controller

import (
	"net/http"

	"github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/HieuNguyenVV/book-hive/internal/log"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/pagination"
	"github.com/HieuNguyenVV/book-hive/internal/server/dto"
	"github.com/HieuNguyenVV/book-hive/internal/server/middleware"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
	"github.com/HieuNguyenVV/book-hive/internal/server/service"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	logger      log.Logger
	userService service.UserService
}

func NewUserController(logger log.Logger, userService service.UserService) *UserController {
	return &UserController{logger: logger, userService: userService}
}

// HandleLogin handles the login request
// @Summary Login
// @Description Login to the system
// @Tags User
// @Accept json
// @Produce json
// @Param login_request body LoginRequest true "Login request"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} AppError
// @Failure 401 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/user/login [post]
func (c *UserController) HandleLogin(ctx *gin.Context) {
	request := &dto.LoginRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		c.logger.Errorf("failed to bind login request: %v", err)
		handleBindError(ctx, err)
		return
	}
	response, err := c.userService.Login(ctx.Request.Context(), request.ToModel())
	if err != nil {
		c.logger.Errorf("failed to login: %v", err)
		handleError(ctx, err)
		return
	}
	responseDto := dto.LoginResponse{}
	responseDto.FromModel(response)
	ctx.JSON(http.StatusOK, responseDto)
}

// HandleRegister handles the register request
// @Summary Register
// @Description Register a new user
// @Tags User
// @Accept json
// @Produce json
// @Param register_request body RegisterRequest true "Register request"
// @Success 200 {object} UserResponse
// @Failure 400 {object} AppError
// @Failure 401 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/user/register [post]
func (c *UserController) HandleRegister(ctx *gin.Context) {
	request := &dto.RegisterRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		c.logger.Errorf("failed to bind register request: %v", err)
		handleBindError(ctx, err)
		return
	}
	response, err := c.userService.Register(ctx.Request.Context(), request.ToModel())
	if err != nil {
		c.logger.Errorf("failed to register: %v", err)
		handleError(ctx, err)
		return
	}
	responseDto := dto.UserResponse{}
	responseDto.FromModel(response)
	ctx.JSON(http.StatusOK, responseDto)
}

// HandleGetUser handles the get user request
// @Summary Get User
// @Description Get a user by id
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 400 {object} AppError
// @Failure 401 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/user/me [get]
// @Security AccessToken
func (c *UserController) HandleGetUser(ctx *gin.Context) {
	user, ok := ctx.Get(middleware.CtxUserKey)
	if !ok {
		c.logger.Errorf("user is required")
		handleError(ctx, errors.ErrInvalidArgument.Reform("user is required"))
		return
	}

	response, err := c.userService.GetUserByID(ctx.Request.Context(), user.(*model.User).ID)
	if err != nil {
		c.logger.Errorf("failed to get user: %v", err)
		handleError(ctx, err)
		return
	}
	responseDto := dto.UserResponse{}
	responseDto.FromModel(response)
	ctx.JSON(http.StatusOK, responseDto)
}

// HandleRefreshToken handles the refresh token request
// @Summary Refresh Token
// @Description Refresh a token
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} LoginResponse
// @Failure 400 {object} AppError
// @Failure 401 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/user/refresh-token [post]
// @Security RefreshToken
func (c *UserController) HandleRefreshToken(ctx *gin.Context) {
	tokenVal, ok := ctx.Get(middleware.CtxRefreshTokenKey)
	if !ok {
		c.logger.Errorf("refresh token is required")
		handleError(ctx, errors.ErrInvalidArgument.Reform("refresh token is required"))
		return
	}
	token, _ := tokenVal.(string)
	if token == "" {
		c.logger.Errorf("refresh token is required")
		handleError(ctx, errors.ErrInvalidArgument.Reform("refresh token is required"))
		return
	}
	response, err := c.userService.RefreshToken(ctx.Request.Context(), token)
	if err != nil {
		c.logger.Errorf("failed to refresh token: %v", err)
		handleError(ctx, err)
		return
	}
	responseDto := dto.LoginResponse{}
	responseDto.FromModel(response)
	ctx.JSON(http.StatusOK, responseDto)
}

// HandleLogout handles the logout request
// @Summary Logout
// @Description Logout a user
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} AppError
// @Failure 401 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/user/logout [post]
// @Security AccessToken
func (c *UserController) HandleLogout(ctx *gin.Context) {
	token := ctx.GetHeader(middleware.RefreshTokenHeader)
	if token == "" {
		c.logger.Errorf("refresh token is required")
		handleError(ctx, errors.ErrInvalidArgument.Reform("refresh token is required"))
		return
	}
	err := c.userService.Logout(ctx.Request.Context(), token)
	if err != nil {
		c.logger.Errorf("failed to logout: %v", err)
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// HandleChangePassword handles the change password request
// @Summary Change Password
// @Description Change a user's password
// @Tags User
// @Accept json
// @Produce json
// @Param change_password_request body ChangePasswordRequest true "Change password request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} AppError
// @Failure 401 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/user/change-password [post]
// @Security AccessToken
func (c *UserController) HandleChangePassword(ctx *gin.Context) {
	userVal, ok := ctx.Get(middleware.CtxUserKey)
	if !ok {
		c.logger.Errorf("user id is required")
		handleError(ctx, errors.ErrInvalidArgument.Reform("user id is required"))
		return
	}
	user := userVal.(*model.User)

	request := &dto.ChangePasswordRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		c.logger.Errorf("failed to bind change password request: %v", err)
		handleBindError(ctx, err)
		return
	}
	err := c.userService.ChangePassword(ctx.Request.Context(), user.ID, request.ToModel())
	if err != nil {
		c.logger.Errorf("failed to change password: %v", err)
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// HandleUpdateUser handles the update user request
// @Summary Update User
// @Description Update a user
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param update_user_request body UpdateUserRequest true "Update user request"
// @Success 200 {object} UserResponse
// @Failure 400 {object} AppError
// @Failure 401 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/user/{id} [patch]
// @Security AccessToken
func (c *UserController) HandleUpdateUser(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		c.logger.Errorf("user id is required")
		handleError(ctx, errors.ErrInvalidArgument.Reform("user id is required"))
		return
	}

	request := &dto.UpdateUserRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		c.logger.Errorf("failed to bind update user request: %v", err)
		handleBindError(ctx, err)
		return
	}
	response, err := c.userService.UpdateUser(ctx.Request.Context(), id, request.ToModel())
	if err != nil {
		c.logger.Errorf("failed to update user: %v", err)
		handleError(ctx, err)
		return
	}
	responseDto := dto.UserResponse{}
	responseDto.FromModel(response)
	ctx.JSON(http.StatusOK, responseDto)
}

// HandleListUsers handles the list users request
// @Summary List Users
// @Description List users
// @Tags User
// @Accept json
// @Produce json
// @Param token query string false "Pagination token"
// @Param search query string false "Search keyword"
// @Param role query string false "User role"
// @Param status query string false "User status"
// @Param limit query int false "Page size"
// @Param offset query int false "Page offset"
// @Success 200 {object} ListUsersResponse
// @Failure 400 {object} AppError
// @Failure 401 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/user/list [get]
// @Security AccessToken
func (c *UserController) HandleListUsers(ctx *gin.Context) {
	request := &dto.ListUsersRequest{}
	if err := ctx.ShouldBindQuery(request); err != nil {
		c.logger.Errorf("failed to bind list users request: %v", err)
		handleBindError(ctx, err)
		return
	}
	limit := pagination.DefaultPaginationLimit
	offset := pagination.DefaultPaginationOffset
	if request.Limit != nil && *request.Limit != 0 {
		limit = *request.Limit
	}
	if request.Offset != nil {
		offset = *request.Offset
	}

	if request.Token != nil && *request.Token != "" {
		err := pagination.GetLimitOffsetFromToken(*request.Token, &limit, &offset)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	response, totalCount, err := c.userService.ListUsers(ctx.Request.Context(), model.ListUsersRequest{
		Search: request.Search,
		Role:   request.Role,
		Status: request.Status,
		Limit:  &limit,
		Offset: &offset,
	})
	if err != nil {
		c.logger.Errorf("failed to list users: %v", err)
		handleError(ctx, err)
		return
	}
	responseDto := dto.ListUsersResponse{}
	responseDto.FromModel(&model.ListUsersResponse{
		Results:  response,
		Total:    totalCount,
		Previous: pagination.GetPrevToken(&limit, &offset, totalCount),
		Next:     pagination.GetNextToken(&limit, &offset, totalCount),
	})
	ctx.JSON(http.StatusOK, responseDto)
}

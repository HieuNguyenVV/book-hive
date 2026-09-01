package service

import (
	"context"
	"time"

	"github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/HieuNguyenVV/book-hive/internal/helper/tx"
	"github.com/HieuNguyenVV/book-hive/internal/log"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/config"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/jwt"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
	"github.com/HieuNguyenVV/book-hive/internal/server/repository"
	"github.com/HieuNguyenVV/book-hive/internal/server/utils"
)

type UserService interface {
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	ListUsers(ctx context.Context, request model.ListUsersRequest) ([]*model.User, int, error)
	Login(ctx context.Context, request model.LoginRequest) (*model.LoginResponse, error)
	Register(ctx context.Context, request model.RegisterRequest) (*model.User, error)
	RefreshToken(ctx context.Context, token string) (*model.LoginResponse, error)
	Logout(ctx context.Context, token string) error
	ChangePassword(ctx context.Context, userID string, request model.ChangePasswordRequest) error
	UpdateUser(ctx context.Context, userID string, request model.UpdateUserRequest) (*model.User, error)
	LoginWebSocket(ctx context.Context, userID string, token string) (*model.LoginWebSocketResponse, map[string]struct{}, error)
}

type userService struct {
	logger           log.Logger
	cfg              *config.Config
	jwtService       jwt.JWTService
	userRepository   repository.UserRepository
	tokenRepository  repository.TokenRepository
	memberRepository repository.MemberRepository
}

func NewUserService(
	logger log.Logger,
	cfg *config.Config,
	jwtService jwt.JWTService,
	userRepository repository.UserRepository,
	tokenRepository repository.TokenRepository,
	memberRepository repository.MemberRepository,
) UserService {
	return &userService{
		logger:           logger,
		cfg:              cfg,
		jwtService:       jwtService,
		userRepository:   userRepository,
		tokenRepository:  tokenRepository,
		memberRepository: memberRepository,
	}
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.userRepository.GetUserByID(ctx, id)
	if err != nil {
		s.logger.Errorf("failed to get user by id: %s", id, err)
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to get user by id: %s", id)
	}
	if user == nil {
		s.logger.Warnf("user not found with id: %s", id)
		return nil, errors.ErrNotFound.Wrap(err).Reform("user not found with id: %s", id)
	}
	return user, nil
}

func (s *userService) Login(ctx context.Context, request model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, request.Email)
	if err != nil {
		s.logger.Errorf("failed to get user by email: %s, error: %v", request.Email, err)
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to get user by email: %s", request.Email)
	}
	if user == nil {
		s.logger.Errorf("user not found with email: %s, error: %v", request.Email, err)
		return nil, errors.ErrNotFound.Wrap(err).Reform("user not found with email: %s", request.Email)
	}
	if user.Status != model.UserStatusActive {
		s.logger.Errorf("user is not active: %s", request.Email)
		return nil, errors.ErrInvalidArgument.WrapString("user is not active").Reform("user is not active: %s", request.Email)
	}

	if !utils.CompareHashAndPassword(request.Password, user.Password) {
		s.logger.Warnf("invalid password for user: %s", request.Email)
		return nil, errors.ErrInvalidArgument.Wrap(err).Reform("invalid password for user: %s", request.Email)
	}

	tokens, err := s.generateTokens(ctx, user)
	if err != nil {
		s.logger.Errorf("failed to generate tokens: %s", user.ID, err)
		return nil, err
	}

	_, err = s.tokenRepository.CreateToken(ctx, &model.Token{
		UserID:    user.ID,
		Token:     utils.HashToken(tokens.RefreshToken),
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTokenTTL()).Unix(),
	})
	if err != nil {
		s.logger.Errorf("failed to create token: %s", user.ID, err)
		return nil, err
	}
	return tokens, nil
}

func (s *userService) generateTokens(ctx context.Context, user *model.User) (*model.LoginResponse, error) {
	claims := &jwt.Claims{
		ID:    user.ID,
		Name:  user.FullName,
		Email: user.Email,
		Role:  string(user.Role),
	}
	accessToken, err := s.jwtService.GenerateAccessToken(claims)
	if err != nil {
		s.logger.Errorf("failed to generate access token: %s", user.ID, err)
		return nil, errors.ErrInternalServerError.Wrap(err).Reform("failed to generate access token: %s", user.ID)
	}
	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		s.logger.Errorf("failed to generate refresh token: %s", user.ID, err)
		return nil, errors.ErrInternalServerError.Wrap(err).Reform("failed to generate refresh token: %s", user.ID)
	}
	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtService.GetAccessTokenTTL())}, nil
}

func (s *userService) Register(ctx context.Context, request model.RegisterRequest) (*model.User, error) {
	exists, err := s.userRepository.CheckUserExists(ctx, request.Email)
	if err != nil {
		s.logger.Errorf("failed to get user by email: %s", request.Email, err)
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to get user by email: %s", request.Email)
	}
	if exists {
		s.logger.Errorf("user already exists with email: %s", request.Email)
		return nil, errors.ErrAlreadyExists.Wrap(err).Reform("user already exists with email: %s", request.Email)
	}

	password, err := utils.HashPassword(request.Password)
	if err != nil {
		s.logger.Errorf("failed to hash password: %s", request.Email, err)
		return nil, errors.ErrInternalServerError.Wrap(err).Reform("failed to hash password: %s", request.Email)
	}

	user, err := s.userRepository.CreateUser(ctx, &model.User{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		FullName:  request.FullName,
		Email:     request.Email,
		Password:  password,
		Role:      model.UserRoleUser,
		Status:    model.UserStatusActive,
	})
	if err != nil {
		s.logger.Errorf("failed to create user: %s", request.Email, err)
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to create user: %s", request.Email)
	}
	// TODO: Send verification email

	return user, nil
}

func (s *userService) RefreshToken(ctx context.Context, token string) (*model.LoginResponse, error) {
	tokenModel, err := s.tokenRepository.GetTokenByToken(ctx, utils.HashToken(token))
	if err != nil {
		s.logger.Errorf("failed to get token by token: %s", token, err)
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to get token by token: %s", token)
	}
	if tokenModel == nil {
		s.logger.Errorf("token not found with token: %s", token)
		return nil, errors.ErrNotFound.Reform("token not found")
	}
	if tokenModel.RevokedAt > 0 {
		s.logger.Errorf("token already revoked: %s", tokenModel.ID)
		return nil, errors.ErrInvalidToken.Reform("token has been revoked")
	}
	if tokenModel.ExpiresAt > 0 && tokenModel.ExpiresAt < time.Now().Unix() {
		s.logger.Errorf("token expired: %s", tokenModel.ID)
		return nil, errors.ErrTokenExpired.Reform("token has expired")
	}

	user, err := s.userRepository.GetUserByID(ctx, tokenModel.UserID)
	if err != nil {
		s.logger.Errorf("failed to get user by id: %s", tokenModel.UserID, err)
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to get user by id: %s", tokenModel.UserID)
	}
	if user == nil {
		s.logger.Errorf("user not found with id: %s", tokenModel.UserID)
		return nil, errors.ErrNotFound.Wrap(err).Reform("user not found with id: %s", tokenModel.UserID)
	}
	if user.Status != model.UserStatusActive {
		s.logger.Errorf("user is not active: %s", tokenModel.UserID)
		return nil, errors.ErrInvalidArgument.WrapString("user is not active").Reform("user is not active: %s", tokenModel.UserID)
	}

	var newTokens *model.LoginResponse
	err = tx.WithTransaction(ctx, s.logger, func(ctxInTx context.Context) error {
		tokens, genErr := s.generateTokens(ctxInTx, user)
		if genErr != nil {
			return genErr
		}
		newTokens = tokens

		if revokeErr := s.tokenRepository.RevokeToken(ctxInTx, tokenModel.Token); revokeErr != nil {
			return errors.ErrDatabase.Wrap(revokeErr).Reform("failed to revoke token: %s", tokenModel.ID)
		}

		_, createErr := s.tokenRepository.CreateToken(ctxInTx, &model.Token{
			UserID:    user.ID,
			Token:     utils.HashToken(tokens.RefreshToken),
			ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTokenTTL()).Unix(),
		})
		if createErr != nil {
			return errors.ErrDatabase.Wrap(createErr).Reform("failed to create token: %s", user.ID)
		}
		return nil
	})
	if err != nil {
		s.logger.Errorf("failed to refresh token: %s", token, err)
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to refresh token: %s", token)
	}
	return newTokens, nil
}

func (s *userService) Logout(ctx context.Context, token string) error {
	tokenModel, err := s.tokenRepository.GetTokenByToken(ctx, utils.HashToken(token))
	if err != nil {
		s.logger.Errorf("failed to get token by token: %s", token, err)
		return errors.ErrDatabase.Wrap(err).Reform("failed to get token by token: %s", token)
	}
	if tokenModel == nil {
		s.logger.Errorf("token not found with token: %s", token)
		return errors.ErrNotFound.Wrap(err).Reform("token not found with token: %s", token)
	}

	err = s.tokenRepository.RevokeToken(ctx, tokenModel.Token)
	if err != nil {
		s.logger.Errorf("failed to revoke token: %s", token, err)
		return errors.ErrDatabase.Wrap(err).Reform("failed to revoke token: %s", token)
	}
	return nil
}

func (s *userService) ChangePassword(ctx context.Context, userID string, request model.ChangePasswordRequest) error {
	user, err := s.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Errorf("failed to get user by id: %s", userID, err)
		return errors.ErrDatabase.Wrap(err).Reform("failed to get user by id: %s", userID)
	}
	if user == nil {
		s.logger.Errorf("user not found with id: %s", userID)
		return errors.ErrNotFound.Wrap(err).Reform("user not found with id: %s", userID)
	}

	if user.Status != model.UserStatusActive {
		s.logger.Errorf("user is not active: %s", userID)
		return errors.ErrInternalServerError.WrapString("user is not active").Reform("user is not active: %s", userID)
	}

	if !utils.CompareHashAndPassword(request.OldPassword, user.Password) {
		s.logger.Errorf("invalid old password for user: %s", userID)
		return errors.ErrInvalidArgument.WrapString("invalid old password for user").Reform("invalid old password for user: %s", userID)
	}

	newPassword, err := utils.HashPassword(request.NewPassword)
	if err != nil {
		s.logger.Errorf("failed to hash new password: %s", userID, err)
		return errors.ErrInternalServerError.Wrap(err).Reform("failed to hash new password: %s", userID)
	}

	_, err = s.userRepository.UpdateUser(ctx, userID, &model.UpdateUserRequest{
		Password: &newPassword,
	})
	if err != nil {
		s.logger.Errorf("failed to update user password: %s", userID, err)
		return errors.ErrDatabase.Wrap(err).Reform("failed to update user password: %s", userID)
	}
	return nil
}

func (s *userService) UpdateUser(ctx context.Context, userID string, request model.UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Errorf("failed to get user by id: %s", userID, err)
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to get user by id: %s", userID)
	}
	if user == nil {
		s.logger.Errorf("user not found with id: %s", userID)
		return nil, errors.ErrNotFound.Wrap(err).Reform("user not found with id: %s", userID)
	}

	updatedUser, err := s.userRepository.UpdateUser(ctx, userID, &request)
	if err != nil {
		s.logger.Errorf("failed to update user: %s", userID, err)
		return nil, errors.ErrDatabase.Wrap(err).Reform("failed to update user: %s", userID)
	}
	return updatedUser, nil
}

func (s *userService) ListUsers(ctx context.Context, request model.ListUsersRequest) ([]*model.User, int, error) {
	users, total, err := s.userRepository.GetUsers(ctx, &request)
	if err != nil {
		s.logger.Errorf("failed to list users: %v", err)
		return nil, 0, errors.ErrDatabase.Wrap(err).Reform("failed to list users")
	}
	return users, total, nil
}

func (s *userService) LoginWebSocket(ctx context.Context, userID string, token string) (*model.LoginWebSocketResponse, map[string]struct{}, error) {
	user, err := s.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Errorf("failed to get user by id: %s", userID, err)
		return nil, nil, errors.ErrDatabase.Wrap(err).Reform("failed to get user by id: %s", userID)
	}
	if user == nil {
		s.logger.Errorf("user not found with id: %s", userID)
		return nil, nil, errors.ErrNotFound.Wrap(err).Reform("user not found with id: %s", userID)
	}
	if user.Status != model.UserStatusActive {
		s.logger.Errorf("user is not active: %s", userID)
		return nil, nil, errors.ErrInvalidArgument.WrapString("user is not active").Reform("user is not active: %s", userID)
	}

	channelIDs, err := s.memberRepository.GetJoinedChannelIDsByUserID(ctx, userID)
	if err != nil {
		s.logger.Errorf("failed to get joined channels by user id: %s", userID, err)
		return nil, nil, errors.ErrDatabase.Wrap(err).Reform("failed to get joined channels by user id: %s", userID)
	}

	topics := make(map[string]struct{})
	for _, channelID := range channelIDs {
		topics[channelID] = struct{}{}
	}
	topics[userID] = struct{}{}

	return &model.LoginWebSocketResponse{
		UserID:       user.ID,
		LoginTs:      time.Now().Unix(),
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		FullName:     user.FullName,
		Email:        user.Email,
		Role:         user.Role,
		Status:       user.Status,
		LastLoginAt:  user.LastLoginAt,
		Key:          token,
		PingInterval: 15,
		PongTimeout:  5,
	}, topics, nil
}

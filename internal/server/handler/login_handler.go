package handler

import (
	"context"
	"net/url"

	appError "github.com/HieuNguyenVV/book-hive/internal/errors"

	"github.com/HieuNguyenVV/book-hive/internal/log"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/jwt"
	wsclient "github.com/HieuNguyenVV/book-hive/internal/server/handler/client"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
	"github.com/HieuNguyenVV/book-hive/internal/server/service"
)

type WebSocketMiddleware interface {
	Handle(c context.Context, s wsclient.WSClient, next func(s wsclient.WSClient) error) error
}

type WebSocketMiddlewareChain struct {
	middlewares []WebSocketMiddleware
}

func NewWebSocketMiddlewareChain() *WebSocketMiddlewareChain {
	return &WebSocketMiddlewareChain{
		middlewares: make([]WebSocketMiddleware, 0),
	}
}

func (c *WebSocketMiddlewareChain) Use(middleware WebSocketMiddleware) {
	c.middlewares = append(c.middlewares, middleware)
}

func (c *WebSocketMiddlewareChain) Execute(ctx context.Context, s wsclient.WSClient) error {
	return c.runCheck(ctx, s, 0)
}

func (c *WebSocketMiddlewareChain) runCheck(ctx context.Context, s wsclient.WSClient, index int) error {
	if index >= len(c.middlewares) {
		return nil // End of chain
	}

	return c.middlewares[index].Handle(ctx, s, func(s wsclient.WSClient) error {
		return c.runCheck(ctx, s, index+1)
	})
}

type TokenValidationMiddleware struct {
	logger      log.Logger
	jwtService  jwt.JWTService
	userService service.UserService
}

func NewTokenValidationMiddleware(logger log.Logger, jwtService jwt.JWTService, userService service.UserService) *TokenValidationMiddleware {
	return &TokenValidationMiddleware{
		logger:      logger,
		jwtService:  jwtService,
		userService: userService,
	}
}

func (m *TokenValidationMiddleware) ValidateQuery(ctx context.Context, values url.Values) (string, *model.User, error) {
	token := values.Get("token")
	if token == "" {
		return "", nil, appError.ErrInvalidToken
	}

	claims, err := m.jwtService.ValidateAccessToken(token)
	if err != nil {
		return "", nil, appError.ErrInvalidToken.Wrap(err)
	}
	if claims == nil || claims.Subject == "" {
		return "", nil, appError.ErrInvalidToken
	}

	user, err := m.userService.GetUserByID(ctx, claims.Subject)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, appError.ErrUserNotFound
	}
	if user.Status != model.UserStatusActive {
		return "", nil, appError.ErrUnauthorized
	}

	return token, user, nil
}

func (m *TokenValidationMiddleware) Handle(ctx context.Context, s wsclient.WSClient, next func(s wsclient.WSClient) error) error {
	token, user, err := m.ValidateQuery(ctx, s.GetRequestURLQuery())
	if err != nil {
		return err
	}
	s.SetUserInfo(model.WebsocketUserInfo{
		User: user,
	})
	s.SetToken(token)
	s.SetUserID(user.ID)
	return next(s)
}

type UserSetupMiddleware struct {
	logger      log.Logger
	userService service.UserService
}

func NewUserSetupMiddleware(logger log.Logger, userService service.UserService) *UserSetupMiddleware {
	return &UserSetupMiddleware{
		logger:      logger,
		userService: userService,
	}
}

func (m *UserSetupMiddleware) Handle(ctx context.Context, s wsclient.WSClient, next func(s wsclient.WSClient) error) error {
	userID, ok := s.GetUserID()
	if !ok {
		return appError.ErrUserNotFound
	}
	token, ok := s.GetToken()
	if !ok {
		return appError.ErrInvalidToken
	}
	loginResponse, topics, err := m.userService.LoginWebSocket(ctx, userID, token)
	if err != nil {
		return err
	}
	s.SetLoginResp(loginResponse)
	s.SetTopics(topics)
	return next(s)
}

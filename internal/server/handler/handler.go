package handler

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	appErrors "github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/HieuNguyenVV/book-hive/internal/helper/event"
	"github.com/HieuNguyenVV/book-hive/internal/log"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/jwt"
	wsclient "github.com/HieuNguyenVV/book-hive/internal/server/handler/client"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
	"github.com/HieuNguyenVV/book-hive/internal/server/service"
	"github.com/google/uuid"
	"github.com/olahol/melody"
)

func MainWebSocketConnectionHandler(
	logger log.Logger,
	jwtService jwt.JWTService,
	userService service.UserService,
	clientRegistry ClientRegistry,
) func(s *melody.Session) {
	tokenMiddleware := NewTokenValidationMiddleware(logger, jwtService, userService)
	userSetupMiddleware := NewUserSetupMiddleware(logger, userService)
	finalSetupMiddleware := NewFinalSetupMiddleware(logger, clientRegistry)

	return func(s *melody.Session) {
		ctx := s.Request.Context()
		wsClient := wsclient.NewWSClient(s, logger)
		chain := NewWebSocketMiddlewareChain()
		chain.Use(tokenMiddleware)
		chain.Use(userSetupMiddleware)
		chain.Use(finalSetupMiddleware)

		if err := chain.Execute(ctx, wsClient); err != nil {
			logger.WithError(err).Warn("websocket middleware rejected connection")
			_ = wsClient.Close()
			return
		}

		logger.Info("websocket client connected")
	}
}

func MainWebSocketDisconnectHandler(logger log.Logger, clientRegistry ClientRegistry) func(s *melody.Session) {
	return func(s *melody.Session) {
		userIDValue, ok := s.Get("userID")
		if !ok {
			return
		}
		userID, ok := userIDValue.(string)
		if !ok || userID == "" {
			return
		}

		connectionIDValue, ok := s.Get("connectionID")
		if !ok {
			return
		}
		connectionID, ok := connectionIDValue.(string)
		if !ok || connectionID == "" {
			return
		}

		clientRegistry.Unregister(userID, connectionID)

		if userInfoValue, ok := s.Get("userInfo"); ok {
			if userInfo, ok := userInfoValue.(model.WebsocketUserInfo); ok && userInfo.CancelFunc != nil {
				userInfo.CancelFunc()
			}
		}

		reason := "client_initiated"
		if isErrorValue, ok := s.Get("is_error"); ok {
			if isError, ok := isErrorValue.(bool); ok && isError {
				reason = "server_error"
			}
		}

		logger.Infof("websocket disconnected user=%s connection=%s reason=%s", userID, connectionID, reason)
	}
}

func MainWebSocketPongHandler(logger log.Logger) func(s *melody.Session) {
	return func(s *melody.Session) {
		userIDValue, _ := s.Get("userID")
		logger.Debugf("websocket protocol pong received user=%v", userIDValue)
	}
}

func MainWebSocketErrorHandler(logger log.Logger) func(s *melody.Session, err error) {
	return func(s *melody.Session, err error) {
		if err == nil {
			return
		}

		errStr := err.Error()
		if strings.Contains(errStr, "close 1005") || strings.Contains(errStr, "close 1000") || strings.Contains(errStr, "close 1001") {
			logger.Debugf("websocket session closed: %v", err)
			return
		}

		s.Set("is_error", true)
		logger.WithError(err).Warn("websocket session error")

		if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "i/o timeout") {
			_ = s.CloseWithMsg([]byte("heartbeat timeout"))
		}
	}
}

type FinalSetupMiddleware struct {
	logger         log.Logger
	clientRegistry ClientRegistry
}

func NewFinalSetupMiddleware(logger log.Logger, clientRegistry ClientRegistry) *FinalSetupMiddleware {
	return &FinalSetupMiddleware{
		logger:         logger,
		clientRegistry: clientRegistry,
	}
}

func (m *FinalSetupMiddleware) Handle(ctx context.Context, s wsclient.WSClient, next func(wsclient.WSClient) error) error {
	userID, ok := s.GetUserID()
	if !ok {
		return appErrors.ErrInvalidArgument.Reform("user id is required")
	}

	loginResp, ok := s.GetLoginResp()
	if !ok || loginResp == nil {
		return appErrors.ErrInvalidArgument.Reform("login response is required")
	}

	topics, ok := s.GetTopics()
	if !ok {
		topics = make(map[string]struct{})
	}

	userInfoValue, ok := s.GetUserInfo()
	if !ok || userInfoValue.User == nil {
		return appErrors.ErrInvalidArgument.Reform("user info is required")
	}

	connectionID := uuid.NewString()
	wsCtx, cancel := context.WithCancel(context.Background())

	userInfo := model.WebsocketUserInfo{
		User:         userInfoValue.User,
		ConnectionID: connectionID,
		Topics:       topics,
		Ctx:          wsCtx,
		CancelFunc:   cancel,
		ConnectedAt:  time.Now().Unix(),
	}

	s.SetConnectionID(connectionID)
	s.SetUserInfo(userInfo)

	if err := sendLogiEvent(s, loginResp); err != nil {
		cancel()
		return err
	}

	m.clientRegistry.Register(userID, connectionID, s)
	s.Connect()

	return next(s)
}

func sendLogiEvent(client wsclient.WSClient, loginResp *model.LoginWebSocketResponse) error {
	payload, err := json.Marshal(loginResp)
	if err != nil {
		return err
	}
	return client.Write(event.WSEventFormat(model.WSEventLOGI, payload))
}

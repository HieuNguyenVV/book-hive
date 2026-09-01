package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/HieuNguyenVV/book-hive/internal/server/handler"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	clientContainer     *ClientContainer
	repositoryContainer *RepositoryContainer
	serviceContainer    *ServiceContainer
	controllerContainer *ControllerContainer
	clientRegistry      handler.ClientRegistry
	wsPubSubServer      *WSPubSubServer
}

func NewServer(clientContainer *ClientContainer, repositoryContainer *RepositoryContainer, serviceContainer *ServiceContainer, controllerContainer *ControllerContainer) *Server {
	m := melody.New()
	clientRegistry := handler.NewClientRegistry()
	if clientContainer.config.WebSocket.MaxMessageSize > 0 {
		m.Config.MaxMessageSize = clientContainer.config.WebSocket.MaxMessageSize
	} else {
		m.Config.MaxMessageSize = 512 * 1024 // 512KB
	}

	writeWaitSec := clientContainer.config.WebSocket.WriteWaitSec
	pongWaitSec := clientContainer.config.WebSocket.PongWaitSec
	pingPeriodSec := clientContainer.config.WebSocket.PingPeriodSec

	if writeWaitSec > 0 {
		m.Config.WriteWait = time.Duration(writeWaitSec) * time.Second
	}
	if pongWaitSec > 0 {
		m.Config.PongWait = time.Duration(pongWaitSec) * time.Second
	}
	if pingPeriodSec > 0 {
		m.Config.PingPeriod = time.Duration(pingPeriodSec) * time.Second
	}
	if m.Config.PingPeriod >= m.Config.PongWait {
		m.Config.PingPeriod = (m.Config.PongWait * 9) / 10
	}

	m.HandleConnect(func(s *melody.Session) {
		handler.MainWebSocketConnectionHandler(
			clientContainer.logger,
			*serviceContainer.jwtService,
			serviceContainer.userService,
			clientRegistry,
		)(s)
	})
	m.HandleMessage(func(s *melody.Session, msg []byte) {
		handler.MainWebSocketMessageHandler(
			clientContainer.logger,
			serviceContainer.messageService,
			clientRegistry,
		)(s, msg)
	})
	m.HandleDisconnect(func(s *melody.Session) {
		handler.MainWebSocketDisconnectHandler(clientContainer.logger, clientRegistry)(s)
	})
	m.HandlePong(func(s *melody.Session) {
		handler.MainWebSocketPongHandler(clientContainer.logger)(s)
	})
	m.HandleError(func(s *melody.Session, err error) {
		handler.MainWebSocketErrorHandler(clientContainer.logger)(s, err)
	})

	engine := clientContainer.httpClient
	{
		engine.GET("/ws", func(c *gin.Context) {
			if err := m.HandleRequest(c.Writer, c.Request); err != nil {
				clientContainer.logger.WithError(err).Error("failed to handle websocket request")
			}
		})
	}
	if clientContainer.config.App.Debug {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	{
		engine.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})
		engine.GET("/readyz", func(c *gin.Context) {
			err := clientContainer.postgres.Ping()
			if err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"message": "database is not ready"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "OK"})
		})
	}

	userGroup := engine.Group("/api/v1/user")
	{
		userGroup.POST("/login", controllerContainer.userController.HandleLogin)
		userGroup.POST("/register", controllerContainer.userController.HandleRegister)
		userGroup.POST("/refresh-token", serviceContainer.refreshTokenMiddleware, controllerContainer.userController.HandleRefreshToken)
		userGroup.POST("/logout", serviceContainer.authMiddleware, controllerContainer.userController.HandleLogout)
		userGroup.GET("/me", serviceContainer.authMiddleware, controllerContainer.userController.HandleGetUser)
		userGroup.POST("/change-password", serviceContainer.authMiddleware, controllerContainer.userController.HandleChangePassword)
		userGroup.GET("/list", serviceContainer.authMiddleware, controllerContainer.userController.HandleListUsers)
		userGroup.PATCH("/:id", serviceContainer.authMiddleware, controllerContainer.userController.HandleUpdateUser)
	}

	channelGroup := engine.Group("/api/v1/channel", serviceContainer.authMiddleware)
	{
		channelGroup.POST("", controllerContainer.channelController.HandleCreateChannel)
		channelGroup.POST("/direct", controllerContainer.channelController.HandleCreateDirectChannel)
		channelGroup.GET("/my/list", controllerContainer.channelController.HandleListMyChannels)
		channelGroup.GET("/list", controllerContainer.channelController.HandleListChannels)
		channelGroup.GET("/:id/messages", controllerContainer.channelController.HandleListChannelMessages)
		channelGroup.GET("/:id", controllerContainer.channelController.HandleGetChannel)
		channelGroup.PATCH("/:id", controllerContainer.channelController.HandleUpdateChannel)
		channelGroup.DELETE("/:id", controllerContainer.channelController.HandleDeleteChannel)
	}

	wsPubSubServer := NewWSPubSubServer(clientContainer.logger, clientContainer.redis, clientRegistry)

	return &Server{
		clientContainer:     clientContainer,
		repositoryContainer: repositoryContainer,
		serviceContainer:    serviceContainer,
		controllerContainer: controllerContainer,
		clientRegistry:      clientRegistry,
		wsPubSubServer:      wsPubSubServer,
	}
}

func (s *Server) Run(ctx context.Context) {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", s.clientContainer.config.App.Port),
		Handler:           s.clientContainer.httpClient,
		ReadHeaderTimeout: 3 * time.Second,
	}
	go func() {
		s.clientContainer.logger.Infof("starting server on port %s", s.clientContainer.config.App.Port)
		s.clientContainer.logger.Infof("Swagger UI: http://localhost:%s/swagger/index.html", s.clientContainer.config.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.clientContainer.logger.WithError(err).Error("failed to start server")
		}
	}()

	go s.wsPubSubServer.Run(ctx)

	<-ctx.Done()

	s.clientContainer.logger.Info("shutting down server")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		s.clientContainer.logger.WithError(err).Error("failed to shutdown server")
	}
	s.clientContainer.postgres.Shutdown()
	if s.clientContainer.redis != nil {
		if err := s.clientContainer.redis.Shutdown(); err != nil {
			s.clientContainer.logger.WithError(err).Error("failed to shutdown redis")
		}
	}
	s.clientContainer.logger.Info("server shutdown complete")
}

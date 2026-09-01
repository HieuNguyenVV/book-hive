package server

import (
	"github.com/HieuNguyenVV/book-hive/internal/log"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/cache"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/config"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/database"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/jwt"
	"github.com/HieuNguyenVV/book-hive/internal/server/controller"
	"github.com/HieuNguyenVV/book-hive/internal/server/middleware"
	"github.com/HieuNguyenVV/book-hive/internal/server/repository"
	"github.com/HieuNguyenVV/book-hive/internal/server/service"
	"github.com/HieuNguyenVV/book-hive/internal/server/validator"
	"github.com/gin-gonic/gin"
)

type ClientContainer struct {
	config     *config.Config
	logger     log.Logger
	httpClient *gin.Engine
	postgres   *database.Postgres
	redis      *cache.Redis
}

type ControllerContainer struct {
	userController    *controller.UserController
	channelController *controller.ChannelController
}

type ServiceContainer struct {
	authMiddleware         gin.HandlerFunc
	refreshTokenMiddleware gin.HandlerFunc
	userService            service.UserService
	channelService         service.ChannelService
	messageService         service.MessageService
	jwtService             *jwt.JWTService
}

type RepositoryContainer struct {
	userRepository    repository.UserRepository
	tokenRepository   repository.TokenRepository
	channelRepository repository.ChannelRepository
	memberRepository  repository.MemberRepository
	messageRepository repository.MessageRepository
}

func NewClientContainer(config *config.Config) *ClientContainer {
	logger := log.New(&log.Config{
		LogLevel: config.Log.LogLevel,
	})

	postgres, err := database.NewPostgres(config.Postgres)
	if err != nil {
		logger.Error("failed to create postgres client", "error", err)
		panic(err)
	}

	redisClient, err := cache.NewRedis(config.Redis)
	if err != nil {
		logger.Error("failed to create redis client", "error", err)
		panic(err)
	}

	httpClient := NewHTTPServer(config, logger)

	return &ClientContainer{
		config:     config,
		logger:     logger,
		httpClient: httpClient,
		postgres:   postgres,
		redis:      redisClient,
	}
}

func NewHTTPServer(config *config.Config, logger log.Logger) *gin.Engine {
	validator.Register()

	var engine *gin.Engine
	if config.App.Debug {
		gin.SetMode(gin.DebugMode)
		engine = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		engine = gin.New()
	}

	engine.Use(
		corsMiddleware,
		middleware.Cors(),
		middleware.Tx(logger),
		gin.Recovery(),
	)
	return engine
}

func corsMiddleware(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Next()
}

func NewServiceContainer(clientContainer *ClientContainer, repositoryContainer *RepositoryContainer) *ServiceContainer {
	jwtService := jwt.NewJWTService(clientContainer.config)
	sendEventService := service.NewSendEventService(clientContainer.logger, clientContainer.redis)
	userService := service.NewUserService(
		clientContainer.logger,
		clientContainer.config,
		jwtService,
		repositoryContainer.userRepository,
		repositoryContainer.tokenRepository,
		repositoryContainer.memberRepository,
	)
	channelService := service.NewChannelService(
		clientContainer.logger,
		repositoryContainer.channelRepository,
		repositoryContainer.memberRepository,
		repositoryContainer.userRepository,
		repositoryContainer.messageRepository,
		sendEventService,
	)
	messageService := service.NewMessageService(
		clientContainer.logger,
		repositoryContainer.channelRepository,
		repositoryContainer.memberRepository,
		repositoryContainer.messageRepository,
		repositoryContainer.userRepository,
		sendEventService,
	)
	authMiddleware := middleware.NewAuthMiddleware(clientContainer.logger, userService, jwtService)
	refreshTokenMiddleware := middleware.NewRefreshTokenMiddleware(clientContainer.logger, userService, jwtService)
	return &ServiceContainer{
		authMiddleware:         authMiddleware,
		refreshTokenMiddleware: refreshTokenMiddleware,
		jwtService:             &jwtService,
		userService:            userService,
		channelService:         channelService,
		messageService:         messageService,
	}
}

func NewRepositoryContainer(clientContainer *ClientContainer) *RepositoryContainer {
	userRepository := repository.NewUserRepository(clientContainer.postgres)
	tokenRepository := repository.NewTokenRepository(clientContainer.postgres)
	channelRepository := repository.NewChannelRepository(clientContainer.postgres)
	memberRepository := repository.NewMemberRepository(clientContainer.postgres)
	messageRepository := repository.NewMessageRepository(clientContainer.postgres)
	return &RepositoryContainer{
		userRepository:    userRepository,
		tokenRepository:   tokenRepository,
		channelRepository: channelRepository,
		memberRepository:  memberRepository,
		messageRepository: messageRepository,
	}
}

func NewControllerContainer(clientContainer *ClientContainer, serviceContainer *ServiceContainer) *ControllerContainer {
	userController := controller.NewUserController(clientContainer.logger, serviceContainer.userService)
	channelController := controller.NewChannelController(clientContainer.logger, serviceContainer.channelService, serviceContainer.messageService)
	return &ControllerContainer{
		userController:    userController,
		channelController: channelController,
	}
}

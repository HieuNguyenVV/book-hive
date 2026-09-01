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

type ChannelController struct {
	logger         log.Logger
	channelService service.ChannelService
	messageService service.MessageService
}

func NewChannelController(logger log.Logger, channelService service.ChannelService, messageService service.MessageService) *ChannelController {
	return &ChannelController{
		logger:         logger,
		channelService: channelService,
		messageService: messageService,
	}
}

// HandleCreateChannel handles create channel request
// @Summary Create Channel
// @Description Create a new channel
// @Tags Channel
// @Accept json
// @Produce json
// @Param create_channel_request body CreateChannelRequest true "Create channel request"
// @Success 200 {object} ChannelResponse
// @Failure 400 {object} AppError
// @Failure 401 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/channel [post]
// @Security AccessToken
func (c *ChannelController) HandleCreateChannel(ctx *gin.Context) {
	user, ok := getAuthenticatedUser(ctx)
	if !ok {
		return
	}

	request := &dto.CreateChannelRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		c.logger.Errorf("failed to bind create channel request: %v", err)
		handleBindError(ctx, err)
		return
	}

	channel, err := c.channelService.CreateChannel(ctx.Request.Context(), user.ID, request.ToModel())
	if err != nil {
		c.logger.Errorf("failed to create channel: %v", err)
		handleError(ctx, err)
		return
	}

	response := dto.ChannelResponse{}
	response.FromModel(channel)
	ctx.JSON(http.StatusOK, response)
}

// HandleCreateDirectChannel handles create or get direct channel between two users
// @Summary Create Direct Channel
// @Description Create or return existing 1:1 direct channel with target user
// @Tags Channel
// @Accept json
// @Produce json
// @Param create_direct_channel_request body CreateDirectChannelRequest true "Create direct channel request"
// @Success 200 {object} ChannelResponse
// @Failure 400 {object} AppError
// @Failure 401 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/channel/direct [post]
// @Security AccessToken
func (c *ChannelController) HandleCreateDirectChannel(ctx *gin.Context) {
	user, ok := getAuthenticatedUser(ctx)
	if !ok {
		return
	}

	request := &dto.CreateDirectChannelRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		c.logger.Errorf("failed to bind create direct channel request: %v", err)
		handleBindError(ctx, err)
		return
	}

	channel, err := c.channelService.CreateDirectChannel(ctx.Request.Context(), user.ID, request.TargetUserID)
	if err != nil {
		c.logger.Errorf("failed to create direct channel: %v", err)
		handleError(ctx, err)
		return
	}

	response := dto.ChannelResponse{}
	response.FromModel(channel)
	ctx.JSON(http.StatusOK, response)
}

// HandleGetChannel handles get channel detail with members and messages
// @Summary Get Channel Detail
// @Description Get channel by id with members, user info and recent messages
// @Tags Channel
// @Accept json
// @Produce json
// @Param id path string true "Channel ID"
// @Param message_limit query int false "Number of recent messages"
// @Param message_offset query int false "Message offset"
// @Success 200 {object} ChannelDetailResponse
// @Failure 401 {object} AppError
// @Failure 403 {object} AppError
// @Failure 404 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/channel/{id} [get]
// @Security AccessToken
func (c *ChannelController) HandleGetChannel(ctx *gin.Context) {
	user, ok := getAuthenticatedUser(ctx)
	if !ok {
		return
	}

	id := ctx.Param("id")
	if id == "" {
		handleError(ctx, errors.ErrInvalidArgument.Reform("channel id is required"))
		return
	}

	query := &dto.GetChannelDetailQuery{}
	if err := ctx.ShouldBindQuery(query); err != nil {
		c.logger.Errorf("failed to bind get channel detail query: %v", err)
		handleBindError(ctx, err)
		return
	}

	detail, err := c.channelService.GetChannelDetail(ctx.Request.Context(), user.ID, id, model.GetChannelDetailRequest{
		MessageLimit:  query.MessageLimit,
		MessageOffset: query.MessageOffset,
	})
	if err != nil {
		c.logger.Errorf("failed to get channel detail: %v", err)
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.ChannelDetailFromModel(detail))
}

// HandleListMyChannels handles list channels joined by current user
// @Summary List My Channels
// @Description List channels joined by current user with members and last message
// @Tags Channel
// @Accept json
// @Produce json
// @Param token query string false "Pagination token"
// @Param search query string false "Search keyword"
// @Param channel_type query string false "Channel type"
// @Param is_public query bool false "Is public"
// @Param channel_url query string false "Channel URL"
// @Param limit query int false "Page size"
// @Param offset query int false "Page offset"
// @Success 200 {object} ListMyChannelsResponse
// @Failure 401 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/channel/my/list [get]
// @Security AccessToken
func (c *ChannelController) HandleListMyChannels(ctx *gin.Context) {
	c.handleListUserChannels(ctx)
}

// HandleListChannelMessages handles list messages in a channel
// @Summary List Channel Messages
// @Description List messages in a channel with sender info
// @Tags Channel
// @Accept json
// @Produce json
// @Param id path string true "Channel ID"
// @Param token query string false "Pagination token"
// @Param limit query int false "Page size"
// @Param offset query int false "Page offset"
// @Success 200 {object} ListChannelMessagesResponse
// @Failure 401 {object} AppError
// @Failure 403 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/channel/{id}/messages [get]
// @Security AccessToken
func (c *ChannelController) HandleListChannelMessages(ctx *gin.Context) {
	user, ok := getAuthenticatedUser(ctx)
	if !ok {
		return
	}

	id := ctx.Param("id")
	if id == "" {
		handleError(ctx, errors.ErrInvalidArgument.Reform("channel id is required"))
		return
	}

	request := &dto.ListChannelsRequest{}
	if err := ctx.ShouldBindQuery(request); err != nil {
		c.logger.Errorf("failed to bind list channel messages request: %v", err)
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
		if err := pagination.GetLimitOffsetFromToken(*request.Token, &limit, &offset); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	result, err := c.messageService.ListChannelMessages(ctx.Request.Context(), user.ID, id, limit, offset)
	if err != nil {
		c.logger.Errorf("failed to list channel messages: %v", err)
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.ListMessagesFromModel(
		result,
		pagination.GetPrevToken(&limit, &offset, result.Total),
		pagination.GetNextToken(&limit, &offset, result.Total),
	))
}

// HandleListChannels handles list channels joined by current user
// @Summary List Channels
// @Description List channels joined by current user with members, other user and last message
// @Tags Channel
// @Accept json
// @Produce json
// @Param token query string false "Pagination token"
// @Param search query string false "Search keyword"
// @Param channel_type query string false "Channel type"
// @Param is_public query bool false "Is public"
// @Param channel_url query string false "Channel URL"
// @Param limit query int false "Page size"
// @Param offset query int false "Page offset"
// @Success 200 {object} ListMyChannelsResponse
// @Failure 400 {object} AppError
// @Failure 401 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/channel/list [get]
// @Security AccessToken
func (c *ChannelController) HandleListChannels(ctx *gin.Context) {
	c.handleListUserChannels(ctx)
}

func (c *ChannelController) handleListUserChannels(ctx *gin.Context) {
	user, ok := getAuthenticatedUser(ctx)
	if !ok {
		return
	}

	request := &dto.ListChannelsRequest{}
	if err := ctx.ShouldBindQuery(request); err != nil {
		c.logger.Errorf("failed to bind list channels request: %v", err)
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
		if err := pagination.GetLimitOffsetFromToken(*request.Token, &limit, &offset); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	result, err := c.channelService.ListChannels(ctx.Request.Context(), user.ID, model.ListChannelsRequest{
		Search:      request.Search,
		ChannelType: request.ChannelType,
		IsPublic:    request.IsPublic,
		ChannelURL:  request.ChannelURL,
	}, limit, offset)
	if err != nil {
		c.logger.Errorf("failed to list channels: %v", err)
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.ListMyChannelsFromModel(
		result,
		pagination.GetPrevToken(&limit, &offset, result.Total),
		pagination.GetNextToken(&limit, &offset, result.Total),
	))
}

// HandleUpdateChannel handles update channel request
// @Summary Update Channel
// @Description Update channel by id
// @Tags Channel
// @Accept json
// @Produce json
// @Param id path string true "Channel ID"
// @Param update_channel_request body UpdateChannelRequest true "Update channel request"
// @Success 200 {object} ChannelResponse
// @Failure 400 {object} AppError
// @Failure 401 {object} AppError
// @Failure 403 {object} AppError
// @Failure 404 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/channel/{id} [patch]
// @Security AccessToken
func (c *ChannelController) HandleUpdateChannel(ctx *gin.Context) {
	user, ok := getAuthenticatedUser(ctx)
	if !ok {
		return
	}

	id := ctx.Param("id")
	if id == "" {
		handleError(ctx, errors.ErrInvalidArgument.Reform("channel id is required"))
		return
	}

	request := &dto.UpdateChannelRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		c.logger.Errorf("failed to bind update channel request: %v", err)
		handleBindError(ctx, err)
		return
	}

	channel, err := c.channelService.UpdateChannel(ctx.Request.Context(), user.ID, user.Role, id, request.ToModel())
	if err != nil {
		c.logger.Errorf("failed to update channel: %v", err)
		handleError(ctx, err)
		return
	}

	response := dto.ChannelResponse{}
	response.FromModel(channel)
	ctx.JSON(http.StatusOK, response)
}

// HandleDeleteChannel handles delete channel request
// @Summary Delete Channel
// @Description Delete channel by id
// @Tags Channel
// @Accept json
// @Produce json
// @Param id path string true "Channel ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} AppError
// @Failure 403 {object} AppError
// @Failure 404 {object} AppError
// @Failure 500 {object} AppError
// @Router /api/v1/channel/{id} [delete]
// @Security AccessToken
func (c *ChannelController) HandleDeleteChannel(ctx *gin.Context) {
	user, ok := getAuthenticatedUser(ctx)
	if !ok {
		return
	}

	id := ctx.Param("id")
	if id == "" {
		handleError(ctx, errors.ErrInvalidArgument.Reform("channel id is required"))
		return
	}

	if err := c.channelService.DeleteChannel(ctx.Request.Context(), user.ID, user.Role, id); err != nil {
		c.logger.Errorf("failed to delete channel: %v", err)
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Channel deleted successfully"})
}

func getAuthenticatedUser(ctx *gin.Context) (*model.User, bool) {
	userVal, ok := ctx.Get(middleware.CtxUserKey)
	if !ok {
		handleError(ctx, errors.ErrInvalidArgument.Reform("user is required"))
		return nil, false
	}
	user, ok := userVal.(*model.User)
	if !ok || user == nil {
		handleError(ctx, errors.ErrInvalidArgument.Reform("user is required"))
		return nil, false
	}
	return user, true
}

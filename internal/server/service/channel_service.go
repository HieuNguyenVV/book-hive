package service

import (
	"context"
	"time"

	"github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/HieuNguyenVV/book-hive/internal/log"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/database"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
	"github.com/HieuNguyenVV/book-hive/internal/server/repository"
	"github.com/google/uuid"
)

type ChannelService interface {
	CreateChannel(ctx context.Context, userID string, request model.CreateChannelRequest) (*model.Channel, error)
	CreateDirectChannel(ctx context.Context, userID string, targetUserID string) (*model.Channel, error)
	GetChannelByID(ctx context.Context, id string) (*model.Channel, error)
	GetChannelDetail(ctx context.Context, userID, channelID string, request model.GetChannelDetailRequest) (*model.ChannelDetail, error)
	ListMyChannels(ctx context.Context, userID string, request model.ListChannelsRequest, limit, offset int) (*model.ListMyChannelsResponse, error)
	ListChannels(ctx context.Context, userID string, request model.ListChannelsRequest, limit, offset int) (*model.ListMyChannelsResponse, error)
	UpdateChannel(ctx context.Context, userID string, role model.UserRole, id string, request model.UpdateChannelRequest) (*model.Channel, error)
	DeleteChannel(ctx context.Context, userID string, role model.UserRole, id string) error
}

type channelService struct {
	logger            log.Logger
	channelRepository repository.ChannelRepository
	memberRepository  repository.MemberRepository
	userRepository    repository.UserRepository
	messageRepository repository.MessageRepository
	sendEventService  SendEventService
}

func NewChannelService(
	logger log.Logger,
	channelRepository repository.ChannelRepository,
	memberRepository repository.MemberRepository,
	userRepository repository.UserRepository,
	messageRepository repository.MessageRepository,
	sendEventService SendEventService,
) ChannelService {
	return &channelService{
		logger:            logger,
		channelRepository: channelRepository,
		memberRepository:  memberRepository,
		userRepository:    userRepository,
		messageRepository: messageRepository,
		sendEventService:  sendEventService,
	}
}

func (s *channelService) CreateChannel(ctx context.Context, userID string, request model.CreateChannelRequest) (*model.Channel, error) {
	channelURL, err := s.generateUniqueChannelURL(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	channel := &model.Channel{
		ChannelURL:  channelURL,
		CoverURL:    optionalString(request.CoverURL),
		Name:        request.Name,
		ChannelType: request.ChannelType,
		Metadata:    database.JSONMap(request.Metadata),
		IsPublic:    request.IsPublic,
		CreateBy:    userID,
		CreateAt:    now,
		UpdateBy:    userID,
		UpdateAt:    now,
	}

	created, err := s.channelRepository.CreateChannel(ctx, channel)
	if err != nil {
		return nil, err
	}
	if _, err := s.memberRepository.CreateMember(ctx, &model.Member{
		ChannelID: created.ID,
		UserID:    userID,
		Role:      model.MemberRoleOperator,
		State:     model.MemberStateJoined,
	}); err != nil {
		return nil, err
	}
	return created, nil
}

func (s *channelService) CreateDirectChannel(ctx context.Context, userID, targetUserID string) (*model.Channel, error) {
	if userID == targetUserID {
		return nil, errors.ErrInvalidArgument.Reform("cannot create direct channel with yourself")
	}

	existing, err := s.channelRepository.FindDistinctDirectChannel(ctx, userID, targetUserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	channelURL, err := s.generateUniqueChannelURL(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	channel := &model.Channel{
		ChannelURL:  channelURL,
		Name:        "Direct Message",
		ChannelType: model.ChannelTypeGroup,
		IsPublic:    false,
		IsDistinct:  true,
		CreateBy:    userID,
		CreateAt:    now,
		UpdateBy:    userID,
		UpdateAt:    now,
	}

	created, err := s.channelRepository.CreateChannel(ctx, channel)
	if err != nil {
		return nil, err
	}

	for _, memberUserID := range []string{userID, targetUserID} {
		member := &model.Member{
			ChannelID: created.ID,
			UserID:    memberUserID,
			Role:      model.MemberRoleMember,
			State:     model.MemberStateJoined,
		}
		if memberUserID == userID {
			member.InviterID = &userID
		}
		if _, err := s.memberRepository.CreateMember(ctx, member); err != nil {
			return nil, err
		}
	}

	if err := s.publishJoinChannelEvent(ctx, created, userID, targetUserID); err != nil {
		s.logger.WithError(err).Warn("failed to publish join channel event")
	}

	return created, nil
}

func (s *channelService) publishJoinChannelEvent(ctx context.Context, channel *model.Channel, inviterID, targetUserID string) error {
	members := make([]model.JoinChannelMember, 0, 2)
	for _, memberUserID := range []string{inviterID, targetUserID} {
		member := model.JoinChannelMember{
			UserID: memberUserID,
			Role:   model.MemberRoleMember,
		}
		if memberUserID == inviterID {
			member.InviterID = inviterID
		}
		user, err := s.userRepository.GetUserByID(ctx, memberUserID)
		if err == nil && user != nil {
			member.FullName = user.FullName
			member.Email = user.Email
		}
		members = append(members, member)
	}

	return s.sendEventService.SendJoinChannelEvent(ctx, model.JoinChannelEvent{
		ChannelID:   channel.ID,
		ChannelURL:  channel.ChannelURL,
		Name:        channel.Name,
		ChannelType: channel.ChannelType,
		IsDistinct:  channel.IsDistinct,
		Members:     members,
		Ts:          time.Now().UnixMilli(),
	}, []string{inviterID, targetUserID})
}

func (s *channelService) generateUniqueChannelURL(ctx context.Context) (string, error) {
	for i := 0; i < 5; i++ {
		channelURL := uuid.NewString()
		exists, err := s.channelRepository.ChannelURLExists(ctx, channelURL, "")
		if err != nil {
			return "", err
		}
		if !exists {
			return channelURL, nil
		}
	}
	return "", errors.ErrInternalServerError.Reform("failed to generate unique channel url")
}

func (s *channelService) GetChannelByID(ctx context.Context, id string) (*model.Channel, error) {
	channel, err := s.channelRepository.GetChannelByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, errors.ErrNotFound.Reform("channel not found: %s", id)
	}
	return channel, nil
}

func (s *channelService) GetChannelDetail(ctx context.Context, userID, channelID string, request model.GetChannelDetailRequest) (*model.ChannelDetail, error) {
	channel, err := s.GetChannelByID(ctx, channelID)
	if err != nil {
		return nil, err
	}

	if err := s.ensureChannelMember(ctx, channelID, userID); err != nil {
		return nil, err
	}

	members, err := s.memberRepository.GetMembersWithUserByChannelID(ctx, channelID)
	if err != nil {
		return nil, err
	}

	limit := 50
	offset := 0
	if request.MessageLimit != nil && *request.MessageLimit > 0 {
		limit = *request.MessageLimit
	}
	if request.MessageOffset != nil {
		offset = *request.MessageOffset
	}

	messages, _, err := s.messageRepository.GetMessagesByChannelID(ctx, channelID, limit, offset)
	if err != nil {
		return nil, err
	}

	detail := &model.ChannelDetail{
		Channel:  channel,
		Members:  members,
		Messages: messages,
	}

	if channel.LastMessageID != nil && *channel.LastMessageID != "" {
		lastMessage, err := s.messageRepository.GetMessageWithSenderByID(ctx, *channel.LastMessageID)
		if err == nil {
			detail.LastMessage = lastMessage
		}
	}

	return detail, nil
}

func (s *channelService) ListMyChannels(ctx context.Context, userID string, request model.ListChannelsRequest, limit, offset int) (*model.ListMyChannelsResponse, error) {
	return s.listUserChannels(ctx, userID, request, limit, offset)
}

func (s *channelService) ListChannels(ctx context.Context, userID string, request model.ListChannelsRequest, limit, offset int) (*model.ListMyChannelsResponse, error) {
	return s.listUserChannels(ctx, userID, request, limit, offset)
}

func (s *channelService) listUserChannels(ctx context.Context, userID string, request model.ListChannelsRequest, limit, offset int) (*model.ListMyChannelsResponse, error) {
	channels, total, err := s.channelRepository.GetJoinedChannelsByUserID(ctx, userID, &request, limit, offset)
	if err != nil {
		return nil, err
	}

	channelIDs := make([]string, len(channels))
	for i, channel := range channels {
		channelIDs[i] = channel.ID
	}

	membersMap, err := s.memberRepository.GetMembersWithUserByChannelIDs(ctx, channelIDs)
	if err != nil {
		return nil, err
	}

	items := make([]*model.MyChannelItem, len(channels))
	for i, channel := range channels {
		members := membersMap[channel.ID]
		item := &model.MyChannelItem{
			Channel: channel,
			Members: members,
		}
		if channel.IsDistinct {
			item.OtherUser = otherUserFromMembers(userID, members)
		}
		if channel.LastMessageID != nil && *channel.LastMessageID != "" {
			lastMessage, err := s.messageRepository.GetMessageWithSenderByID(ctx, *channel.LastMessageID)
			if err == nil {
				item.LastMessage = lastMessage
			}
		}
		items[i] = item
	}

	return &model.ListMyChannelsResponse{
		Results: items,
		Total:   total,
	}, nil
}

func otherUserFromMembers(currentUserID string, members []*model.ChannelMemberDetail) *model.UserBrief {
	for _, member := range members {
		if member.UserID != currentUserID {
			user := member.User
			return &user
		}
	}
	return nil
}

func (s *channelService) ensureChannelMember(ctx context.Context, channelID, userID string) error {
	isMember, err := s.memberRepository.IsMember(ctx, channelID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.ErrForbidden.Reform("forbidden to access channel: %s", channelID)
	}
	return nil
}

func (s *channelService) UpdateChannel(ctx context.Context, userID string, role model.UserRole, id string, request model.UpdateChannelRequest) (*model.Channel, error) {
	channel, err := s.channelRepository.GetChannelByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canManageChannel(channel, userID, role) {
		return nil, errors.ErrForbidden.Reform("forbidden to update channel: %s", id)
	}

	return s.channelRepository.UpdateChannel(ctx, id, &request, userID)
}

func (s *channelService) DeleteChannel(ctx context.Context, userID string, role model.UserRole, id string) error {
	channel, err := s.channelRepository.GetChannelByID(ctx, id)
	if err != nil {
		return err
	}
	if !canManageChannel(channel, userID, role) {
		return errors.ErrForbidden.Reform("forbidden to delete channel: %s", id)
	}

	return s.channelRepository.DeleteChannel(ctx, id)
}

func canManageChannel(channel *model.Channel, userID string, role model.UserRole) bool {
	return channel.CreateBy == userID || role == model.UserRoleAdmin
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

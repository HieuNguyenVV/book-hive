package service

import (
	"context"

	"github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/HieuNguyenVV/book-hive/internal/log"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
	"github.com/HieuNguyenVV/book-hive/internal/server/repository"
)

type MessageService interface {
	SendMessage(ctx context.Context, userID string, request model.SocketMessageCreateRequest) (*model.MessageCreateResponse, error)
	ListChannelMessages(ctx context.Context, userID, channelID string, limit, offset int) (*model.ListChannelMessagesResponse, error)
}

type messageService struct {
	logger            log.Logger
	channelRepository repository.ChannelRepository
	memberRepository  repository.MemberRepository
	messageRepository repository.MessageRepository
	userRepository    repository.UserRepository
	sendEventService  SendEventService
}

func NewMessageService(
	logger log.Logger,
	channelRepository repository.ChannelRepository,
	memberRepository repository.MemberRepository,
	messageRepository repository.MessageRepository,
	userRepository repository.UserRepository,
	sendEventService SendEventService,
) MessageService {
	return &messageService{
		logger:            logger,
		channelRepository: channelRepository,
		memberRepository:  memberRepository,
		messageRepository: messageRepository,
		userRepository:    userRepository,
		sendEventService:  sendEventService,
	}
}

func (s *messageService) SendMessage(ctx context.Context, userID string, request model.SocketMessageCreateRequest) (*model.MessageCreateResponse, error) {
	if request.ChannelURL == "" {
		return nil, errors.ErrInvalidArgument.Reform("channel_url is required")
	}

	channel, err := s.channelRepository.GetChannelByURL(ctx, request.ChannelURL)
	if err != nil {
		return nil, err
	}

	isMember, err := s.memberRepository.IsMember(ctx, channel.ID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.ErrForbidden.Reform("forbidden to send message to channel: %s", request.ChannelURL)
	}

	messageText := ""
	if request.Message != nil {
		messageText = *request.Message
	}

	var reqID *string
	if request.ReqID != "" {
		reqID = &request.ReqID
	}

	created, err := s.messageRepository.CreateMessage(ctx, &model.Message{
		ChannelID:   channel.ID,
		SenderID:    userID,
		Content:     messageText,
		MessageType: string(model.WSEventMESG),
		CustomType:  request.CustomType,
		ReqID:       reqID,
	})
	if err != nil {
		return nil, err
	}

	if err := s.messageRepository.UpdateChannelLastMessage(ctx, channel.ID, created.ID); err != nil {
		return nil, err
	}

	sender, err := s.userBrief(ctx, userID)
	if err != nil {
		return nil, err
	}

	ack := &model.MessageCreateResponse{
		ChannelID:   channel.ID,
		ChannelURL:  channel.ChannelURL,
		ChannelName: channel.Name,
		IsDistinct:  channel.IsDistinct,
		Message:     messageText,
		MsgID:       created.ID,
		Ts:          created.CreatedAt * 1000,
		UserID:      userID,
		ReqID:       reqID,
		Sender:      sender,
	}

	pushMessage := model.NewMessage{
		ChannelID:   channel.ID,
		ChannelURL:  channel.ChannelURL,
		ChannelName: channel.Name,
		IsDistinct:  channel.IsDistinct,
		ChannelType: channel.ChannelType,
		MsgID:       created.ID,
		Message:     messageText,
		UserID:      userID,
		Sender:      sender,
		Ts:          created.CreatedAt * 1000,
		ReqID:       reqID,
	}

	if err := s.sendEventService.SendChannelEvent(ctx, channel.ID, model.RedisPublishMessage{
		Event:        model.WSEventMESG,
		Message:      pushMessage,
		SenderID:     userID,
		ExceptUserID: userID,
		IsFromAPI:    false,
	}); err != nil {
		s.logger.WithError(err).Error("failed to publish message event")
		return nil, errors.ErrInternalServerError.Wrap(err).Reform("failed to publish message")
	}

	return ack, nil
}

func (s *messageService) ListChannelMessages(ctx context.Context, userID, channelID string, limit, offset int) (*model.ListChannelMessagesResponse, error) {
	isMember, err := s.memberRepository.IsMember(ctx, channelID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.ErrForbidden.Reform("forbidden to list messages in channel: %s", channelID)
	}

	messages, total, err := s.messageRepository.GetMessagesByChannelID(ctx, channelID, limit, offset)
	if err != nil {
		return nil, err
	}

	return &model.ListChannelMessagesResponse{
		Results: messages,
		Total:   total,
	}, nil
}

func (s *messageService) userBrief(ctx context.Context, userID string) (model.UserBrief, error) {
	user, err := s.userRepository.GetUserByID(ctx, userID)
	if err != nil {
		return model.UserBrief{}, err
	}
	if user == nil {
		return model.UserBrief{}, errors.ErrNotFound.Reform("user not found: %s", userID)
	}
	return model.UserBrief{
		ID:        user.ID,
		FullName:  user.FullName,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}

package service

import (
	"context"
	"encoding/json"

	"github.com/HieuNguyenVV/book-hive/internal/log"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/cache"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
)

type SendEventService interface {
	SendChannelEvent(ctx context.Context, channelID string, payload model.RedisPublishMessage) error
	SendUserEvent(ctx context.Context, userID string, payload model.RedisPublishMessage) error
	SendJoinChannelEvent(ctx context.Context, event model.JoinChannelEvent, memberUserIDs []string) error
}

type sendEventService struct {
	logger log.Logger
	redis  *cache.Redis
}

func NewSendEventService(logger log.Logger, redis *cache.Redis) SendEventService {
	return &sendEventService{
		logger: logger,
		redis:  redis,
	}
}

func (s *sendEventService) SendChannelEvent(ctx context.Context, channelID string, payload model.RedisPublishMessage) error {
	payload.Topic = channelID
	payload.ChannelID = channelID
	payload.Category = model.EventCategoryChannel

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.redis.Publish(ctx, model.GlobalPubSubTopic, body)
}

func (s *sendEventService) SendUserEvent(ctx context.Context, userID string, payload model.RedisPublishMessage) error {
	payload.Topic = userID
	payload.Category = model.EventCategoryUser

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.redis.Publish(ctx, model.GlobalPubSubTopic, body)
}

func (s *sendEventService) SendJoinChannelEvent(ctx context.Context, event model.JoinChannelEvent, memberUserIDs []string) error {
	for _, userID := range memberUserIDs {
		if err := s.SendUserEvent(ctx, userID, model.RedisPublishMessage{
			Event:          model.WSEventJOIN,
			Message:        event,
			SpecificUserID: userID,
			IsFromAPI:      true,
		}); err != nil {
			return err
		}
	}
	return nil
}

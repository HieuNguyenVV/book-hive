package server

import (
	"context"

	"github.com/HieuNguyenVV/book-hive/internal/log"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/cache"
	"github.com/HieuNguyenVV/book-hive/internal/server/handler"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
)

type WSPubSubServer struct {
	logger         log.Logger
	redis          *cache.Redis
	pubSubConsumer *handler.PubSubConsumer
}

func NewWSPubSubServer(logger log.Logger, redis *cache.Redis, clientRegistry handler.ClientRegistry) *WSPubSubServer {
	return &WSPubSubServer{
		logger:         logger,
		redis:          redis,
		pubSubConsumer: handler.NewPubSubConsumer(logger, clientRegistry),
	}
}

func (s *WSPubSubServer) Run(ctx context.Context) {
	pubsub := s.redis.Subscribe(ctx, model.GlobalPubSubTopic)
	defer pubsub.Close()

	ch := pubsub.Channel()
	s.logger.Infof("redis pubsub subscribed to %s", model.GlobalPubSubTopic)

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			s.pubSubConsumer.HandleMessage(ctx, []byte(msg.Payload))
		}
	}
}

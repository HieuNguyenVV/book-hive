package handler

import (
	"context"
	"encoding/json"

	"github.com/HieuNguyenVV/book-hive/internal/log"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
)

type PubSubConsumer struct {
	logger         log.Logger
	clientRegistry ClientRegistry
}

func NewPubSubConsumer(logger log.Logger, clientRegistry ClientRegistry) *PubSubConsumer {
	return &PubSubConsumer{
		logger:         logger,
		clientRegistry: clientRegistry,
	}
}

func (c *PubSubConsumer) HandleMessage(ctx context.Context, payload []byte) {
	var msg model.RedisPublishMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		c.logger.WithError(err).Warn("failed to unmarshal pubsub payload")
		return
	}
	c.clientRegistry.BroadcastMessage(ctx, msg)
}

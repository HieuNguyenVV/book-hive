package client

import (
	"context"

	"github.com/HieuNguyenVV/book-hive/internal/server/model"
)

type MessageWithContext struct {
	Ctx     context.Context
	Payload model.RedisPublishMessage
}

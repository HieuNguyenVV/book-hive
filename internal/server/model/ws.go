package model

import "context"

type WebsocketUserInfo struct {
	User         *User
	ConnectionID string
	Topics       map[string]struct{}
	Ctx          context.Context
	CancelFunc   context.CancelFunc
	ConnectedAt  int64
}

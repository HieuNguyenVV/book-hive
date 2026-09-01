package handler

import (
	"context"
	"sync"

	wsclient "github.com/HieuNguyenVV/book-hive/internal/server/handler/client"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
)

type ClientRegistry interface {
	Register(userID, connectionID string, client wsclient.WSClient)
	Unregister(userID, connectionID string)
	GetClient(userID, connectionID string) (wsclient.WSClient, bool)
	BroadcastMessage(ctx context.Context, msg model.RedisPublishMessage)
}

type clientRegistry struct {
	mu      sync.RWMutex
	clients map[string]map[string]wsclient.WSClient
}

func NewClientRegistry() ClientRegistry {
	return &clientRegistry{
		clients: make(map[string]map[string]wsclient.WSClient),
	}
}

func (r *clientRegistry) Register(userID, connectionID string, client wsclient.WSClient) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.clients[userID]; !ok {
		r.clients[userID] = make(map[string]wsclient.WSClient)
	}
	r.clients[userID][connectionID] = client
}

func (r *clientRegistry) Unregister(userID, connectionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	connections, ok := r.clients[userID]
	if !ok {
		return
	}
	delete(connections, connectionID)
	if len(connections) == 0 {
		delete(r.clients, userID)
	}
}

func (r *clientRegistry) GetClient(userID, connectionID string) (wsclient.WSClient, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	connections, ok := r.clients[userID]
	if !ok {
		return nil, false
	}
	client, ok := connections[connectionID]
	return client, ok
}

func (r *clientRegistry) BroadcastMessage(ctx context.Context, msg model.RedisPublishMessage) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for userID, connections := range r.clients {
		if msg.SpecificUserID != "" && userID != msg.SpecificUserID {
			continue
		}
		if msg.ExceptUserID != "" && userID == msg.ExceptUserID {
			continue
		}

		for _, client := range connections {
			if !client.ShouldNotify(msg.Topic) {
				continue
			}
			_ = client.Notify(ctx, msg)
		}
	}
}

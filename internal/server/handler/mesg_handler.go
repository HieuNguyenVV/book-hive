package handler

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	appErrors "github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/HieuNguyenVV/book-hive/internal/helper/event"
	"github.com/HieuNguyenVV/book-hive/internal/log"
	wsclient "github.com/HieuNguyenVV/book-hive/internal/server/handler/client"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
	"github.com/HieuNguyenVV/book-hive/internal/server/service"
	"github.com/olahol/melody"
)

func MainWebSocketMessageHandler(
	logger log.Logger,
	messageService service.MessageService,
	clientRegistry ClientRegistry,
) func(s *melody.Session, msg []byte) {
	return func(s *melody.Session, msg []byte) {
		ProcessWebSocketMessage(s, msg, logger, messageService, clientRegistry)
	}
}

func ProcessWebSocketMessage(
	session *melody.Session,
	msg []byte,
	logger log.Logger,
	messageService service.MessageService,
	clientRegistry ClientRegistry,
) {
	if len(msg) < 4 {
		logger.Warn("websocket message too short")
		return
	}

	userID, ok := session.Get("userID")
	if !ok {
		logger.Warn("websocket message received without authenticated user")
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		logger.Warn("websocket message received with invalid user id")
		return
	}

	connectionIDValue, ok := session.Get("connectionID")
	if !ok {
		logger.Warn("websocket message received without connection id")
		return
	}
	connectionID, ok := connectionIDValue.(string)
	if !ok || connectionID == "" {
		logger.Warn("websocket message received with invalid connection id")
		return
	}

	client, ok := clientRegistry.GetClient(userIDStr, connectionID)
	if !ok {
		logger.Warnf("websocket client not found for user=%s connection=%s", userIDStr, connectionID)
		return
	}

	eventName := string(msg[:4])
	payload := trimMessagePayload(msg[4:])
	if !model.IsValidWSEvent(eventName) {
		client.SendErrorEvent(appErrors.ErrInvalidValue.Reform("invalid websocket event: %s", eventName), "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	routeWSEvent(ctx, client, model.WSEvent(eventName), payload, messageService, logger)
}

func routeWSEvent(
	ctx context.Context,
	client wsclient.WSClient,
	eventName model.WSEvent,
	payload []byte,
	messageService service.MessageService,
	logger log.Logger,
) {
	switch eventName {
	case model.WSEventPING:
		handlePingEvent(client, payload, logger)
	case model.WSEventMESG:
		handleMesgEvent(ctx, client, payload, messageService, logger)
	default:
		client.SendErrorEvent(appErrors.ErrInvalidValue.Reform("unsupported websocket event: %s", eventName), "")
	}
}

func handlePingEvent(client wsclient.WSClient, payload []byte, logger log.Logger) {
	var req model.Ping
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &req); err != nil {
			client.SendErrorEvent(appErrors.ErrInvalidValue.Wrap(err), req.ReqID)
			return
		}
	}

	now := time.Now().UnixMilli()
	resp, err := json.Marshal(model.Pong{Sts: now, Ts: now})
	if err != nil {
		logger.WithError(err).Error("failed to marshal pong response")
		return
	}
	if client.IsClosed() {
		return
	}
	if err := client.Write(event.WSEventFormat(model.WSEventPONG, resp)); err != nil {
		logger.WithError(err).Warn("failed to write pong response")
	}
}

func handleMesgEvent(
	ctx context.Context,
	client wsclient.WSClient,
	payload []byte,
	messageService service.MessageService,
	logger log.Logger,
) {
	var req model.SocketMessageCreateRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		client.SendErrorEvent(appErrors.ErrInvalidValue.Wrap(err), req.ReqID)
		return
	}

	userID, ok := client.GetUserID()
	if !ok {
		client.SendErrorEvent(appErrors.ErrUnauthorized, req.ReqID)
		return
	}

	ack, err := messageService.SendMessage(ctx, userID, req)
	if err != nil {
		logger.WithError(err).Warn("failed to handle MESG event")
		client.SendErrorEvent(err, req.ReqID)
		return
	}

	ackPayload, err := json.Marshal(ack)
	if err != nil {
		client.SendErrorEvent(appErrors.ErrInternalServerError.Wrap(err), req.ReqID)
		return
	}
	if client.IsClosed() {
		return
	}

	if err := client.Write(event.WSEventFormat(model.WSEventMESG, ackPayload)); err != nil {
		logger.WithError(err).Warn("failed to write MESG ack")
		client.SendErrorEvent(err, req.ReqID)
	}
}

func trimMessagePayload(payload []byte) []byte {
	return []byte(strings.TrimSpace(string(payload)))
}

package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"sync"
	"time"

	appErrors "github.com/HieuNguyenVV/book-hive/internal/errors"
	"github.com/HieuNguyenVV/book-hive/internal/helper/event"
	"github.com/HieuNguyenVV/book-hive/internal/log"
	"github.com/HieuNguyenVV/book-hive/internal/server/model"
	"github.com/olahol/melody"
)

const (
	sessionKeyUserInfo     = "userInfo"
	sessionKeyUserID       = "userID"
	sessionKeyLoginResp    = "loginResp"
	sessionKeyTopics       = "topics"
	sessionKeyToken        = "token"
	sessionKeyConnectionID = "connectionID"
	sessionKeyIsError      = "is_error"
)

type WSClient interface {
	Connect()
	GetRequestURLQuery() url.Values
	SetUserInfo(info model.WebsocketUserInfo)
	GetUserInfo() (model.WebsocketUserInfo, bool)
	SetUserID(userID string)
	GetUserID() (string, bool)
	SetLoginResp(loginResp *model.LoginWebSocketResponse)
	GetLoginResp() (*model.LoginWebSocketResponse, bool)
	SetTopics(topics map[string]struct{})
	GetTopics() (map[string]struct{}, bool)
	SetToken(token string)
	GetToken() (string, bool)
	SetConnectionID(connectionID string)
	GetConnectionID() (string, bool)
	SetError(hasError bool)
	IsError() bool
	Write(data []byte) error
	IsClosed() bool
	Close() error
	SendErrorEvent(err error, reqID string)
	ShouldNotify(topic string) bool
	AddTopic(topic string)
	Notify(ctx context.Context, msg model.RedisPublishMessage) error
	GetUnderlyingSession() *melody.Session
}

type wsClient struct {
	Context
	quitChan  chan struct{}
	writeChan chan MessageWithContext
	closeOnce sync.Once
	logger    log.Logger
}

func NewWSClient(session *melody.Session, logger log.Logger) WSClient {
	return &wsClient{
		Context:   NewMelodyClientContext(session),
		quitChan:  make(chan struct{}),
		writeChan: make(chan MessageWithContext, 256),
		logger:    logger,
	}
}

func (m *wsClient) Connect() {
	go m.writePump()
}

func (m *wsClient) writePump() {
	for {
		select {
		case <-m.quitChan:
			return
		case msgWithCtx, ok := <-m.writeChan:
			if !ok {
				return
			}
			if err := m.processMessage(msgWithCtx.Ctx, msgWithCtx.Payload); err != nil {
				m.logger.WithError(err).Warn("failed to process websocket outbound message")
			}
		}
	}
}

func (m *wsClient) processMessage(ctx context.Context, data model.RedisPublishMessage) error {
	if m.IsClosed() {
		return nil
	}
	if !m.ShouldNotify(data.Topic) {
		return nil
	}

	if data.Event == model.WSEventJOIN {
		if channelID := joinChannelIDFromMessage(data.Message); channelID != "" {
			m.AddTopic(channelID)
		}
	}

	userID, ok := m.GetUserID()
	if ok {
		if data.ExceptUserID != "" && userID == data.ExceptUserID {
			return nil
		}
		if data.SpecificUserID != "" && userID != data.SpecificUserID {
			return nil
		}
		if data.SenderID != "" && data.SenderID == userID && !data.IsFromAPI {
			return nil
		}
	}

	payload, err := json.Marshal(data.Message)
	if err != nil {
		return err
	}
	return m.Write(event.WSEventFormat(data.Event, payload))
}

func (m *wsClient) Write(data []byte) error {
	if m.IsClosed() {
		return melody.ErrSessionClosed
	}
	return m.GetUnderlyingSession().Write(data)
}

func (m *wsClient) IsClosed() bool {
	session := m.GetUnderlyingSession()
	return session == nil || session.IsClosed()
}

func (m *wsClient) Close() error {
	var closeErr error
	m.closeOnce.Do(func() {
		close(m.quitChan)
		close(m.writeChan)
		closeErr = m.GetUnderlyingSession().Close()
	})
	return closeErr
}

func (m *wsClient) SendErrorEvent(err error, reqID string) {
	if err == nil || m.IsClosed() {
		return
	}

	code := appErrors.ErrInternalServerError.Code
	message := err.Error()
	var appErr appErrors.AppError
	if errors.As(err, &appErr) {
		code = appErr.Code
		message = appErr.Message
	}

	resp := model.WSErrorResponse{
		Code:    code,
		Message: message,
		ReqID:   reqID,
	}
	payload, marshalErr := json.Marshal(resp)
	if marshalErr != nil {
		m.logger.WithError(marshalErr).Error("failed to marshal websocket error response")
		return
	}
	if writeErr := m.Write(event.WSEventFormat(model.WSEventEROR, payload)); writeErr != nil {
		m.logger.WithError(writeErr).Warn("failed to send websocket error event")
	}
}

func (m *wsClient) ShouldNotify(topic string) bool {
	if topic == "" {
		return false
	}
	topics, ok := m.GetTopics()
	if !ok {
		return false
	}
	_, subscribed := topics[topic]
	return subscribed
}

func (m *wsClient) AddTopic(topic string) {
	if topic == "" {
		return
	}
	topics, ok := m.GetTopics()
	if !ok {
		topics = make(map[string]struct{})
	}
	topics[topic] = struct{}{}
	m.SetTopics(topics)
}

func joinChannelIDFromMessage(message interface{}) string {
	raw, err := json.Marshal(message)
	if err != nil {
		return ""
	}
	var evt model.JoinChannelEvent
	if err := json.Unmarshal(raw, &evt); err != nil {
		return ""
	}
	return evt.ChannelID
}

func (m *wsClient) GetRequestURLQuery() url.Values {
	return m.GetUnderlyingSession().Request.URL.Query()
}

func (m *wsClient) SetUserInfo(info model.WebsocketUserInfo) {
	m.GetUnderlyingSession().Set(sessionKeyUserInfo, info)
}

func (m *wsClient) GetUserInfo() (model.WebsocketUserInfo, bool) {
	value, ok := m.GetUnderlyingSession().Get(sessionKeyUserInfo)
	if !ok {
		return model.WebsocketUserInfo{}, false
	}
	info, ok := value.(model.WebsocketUserInfo)
	return info, ok
}

func (m *wsClient) SetUserID(userID string) {
	m.GetUnderlyingSession().Set(sessionKeyUserID, userID)
}

func (m *wsClient) GetUserID() (string, bool) {
	value, ok := m.GetUnderlyingSession().Get(sessionKeyUserID)
	if !ok {
		return "", false
	}
	userID, ok := value.(string)
	return userID, ok
}

func (m *wsClient) SetLoginResp(loginResp *model.LoginWebSocketResponse) {
	m.GetUnderlyingSession().Set(sessionKeyLoginResp, loginResp)
}

func (m *wsClient) GetLoginResp() (*model.LoginWebSocketResponse, bool) {
	value, ok := m.GetUnderlyingSession().Get(sessionKeyLoginResp)
	if !ok {
		return nil, false
	}
	loginResp, ok := value.(*model.LoginWebSocketResponse)
	return loginResp, ok
}

func (m *wsClient) SetTopics(topics map[string]struct{}) {
	m.GetUnderlyingSession().Set(sessionKeyTopics, topics)
}

func (m *wsClient) GetTopics() (map[string]struct{}, bool) {
	value, ok := m.GetUnderlyingSession().Get(sessionKeyTopics)
	if !ok {
		return nil, false
	}
	topics, ok := value.(map[string]struct{})
	return topics, ok
}

func (m *wsClient) SetToken(token string) {
	m.GetUnderlyingSession().Set(sessionKeyToken, token)
}

func (m *wsClient) GetToken() (string, bool) {
	value, ok := m.GetUnderlyingSession().Get(sessionKeyToken)
	if !ok {
		return "", false
	}
	token, ok := value.(string)
	return token, ok
}

func (m *wsClient) SetConnectionID(connectionID string) {
	m.GetUnderlyingSession().Set(sessionKeyConnectionID, connectionID)
}

func (m *wsClient) GetConnectionID() (string, bool) {
	value, ok := m.GetUnderlyingSession().Get(sessionKeyConnectionID)
	if !ok {
		return "", false
	}
	connectionID, ok := value.(string)
	return connectionID, ok
}

func (m *wsClient) SetError(hasError bool) {
	m.GetUnderlyingSession().Set(sessionKeyIsError, hasError)
}

func (m *wsClient) IsError() bool {
	value, ok := m.GetUnderlyingSession().Get(sessionKeyIsError)
	if !ok {
		return false
	}
	isError, ok := value.(bool)
	return ok && isError
}

func (m *wsClient) Notify(_ context.Context, msg model.RedisPublishMessage) error {
	if m.IsClosed() {
		return melody.ErrSessionClosed
	}
	select {
	case m.writeChan <- MessageWithContext{Ctx: context.Background(), Payload: msg}:
		return nil
	default:
		return appErrors.ErrInternalServerError.Reform("websocket write channel is full")
	}
}

func (m *wsClient) ConnectedDuration() time.Duration {
	info, ok := m.GetUserInfo()
	if !ok || info.ConnectedAt == 0 {
		return 0
	}
	return time.Since(time.Unix(info.ConnectedAt, 0))
}

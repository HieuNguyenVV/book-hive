package model

type Ping struct {
	ID     int64  `json:"id"`
	Active int    `json:"active"`
	ReqID  string `json:"req_id"`
}

type Pong struct {
	Sts int64 `json:"sts"`
	Ts  int64 `json:"ts"`
}

type SocketMessageCreateRequest struct {
	ChannelURL string  `json:"channel_url"`
	Message    *string `json:"message"`
	ReqID      string  `json:"req_id"`
	CustomType string  `json:"custom_type"`
	Data       string  `json:"data"`
}

type MessageCreateResponse struct {
	ChannelID   string    `json:"channel_id"`
	ChannelURL  string    `json:"channel_url"`
	ChannelName string    `json:"channel_name"`
	IsDistinct  bool      `json:"is_distinct"`
	Message     string    `json:"message"`
	MsgID       string    `json:"msg_id"`
	Ts          int64     `json:"ts"`
	ReqID       *string   `json:"req_id,omitempty"`
	UserID      string    `json:"user_id"`
	Sender      UserBrief `json:"sender"`
}

type WSErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	ReqID   string `json:"req_id,omitempty"`
}

package model

type UserBrief struct {
	ID        string `json:"id"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ChannelMemberDetail struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	State     string    `json:"state"`
	JoinedAt  int64     `json:"joined_at"`
	User      UserBrief `json:"user"`
}

type MessageWithSender struct {
	ID          string    `json:"id"`
	ChannelID   string    `json:"channel_id"`
	Content     string    `json:"content"`
	MessageType string    `json:"message_type"`
	CustomType  string    `json:"custom_type"`
	ReqID       *string   `json:"req_id,omitempty"`
	CreatedAt   int64     `json:"created_at"`
	UpdatedAt   int64     `json:"updated_at"`
	Sender      UserBrief `json:"sender"`
}

type ChannelDetail struct {
	Channel     *Channel               `json:"channel"`
	Members     []*ChannelMemberDetail `json:"members"`
	Messages    []*MessageWithSender   `json:"messages"`
	LastMessage *MessageWithSender     `json:"last_message,omitempty"`
}

type MyChannelItem struct {
	Channel     *Channel               `json:"channel"`
	Members     []*ChannelMemberDetail `json:"members"`
	LastMessage *MessageWithSender     `json:"last_message,omitempty"`
	OtherUser   *UserBrief             `json:"other_user,omitempty"`
}

type ListMyChannelsResponse struct {
	Results  []*MyChannelItem `json:"results"`
	Total    int              `json:"total"`
	Previous string           `json:"previous"`
	Next     string           `json:"next"`
}

type ListChannelMessagesRequest struct {
	Limit  *int `json:"limit"`
	Offset *int `json:"offset"`
}

type ListChannelMessagesResponse struct {
	Results  []*MessageWithSender `json:"results"`
	Total    int                  `json:"total"`
	Previous string               `json:"previous"`
	Next     string               `json:"next"`
}

type GetChannelDetailRequest struct {
	MessageLimit  *int `json:"message_limit"`
	MessageOffset *int `json:"message_offset"`
}

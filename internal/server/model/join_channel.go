package model

type JoinChannelMember struct {
	UserID    string `json:"user_id"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	InviterID string `json:"inviter_id,omitempty"`
}

type JoinChannelEvent struct {
	ChannelID   string              `json:"channel_id"`
	ChannelURL  string              `json:"channel_url"`
	Name        string              `json:"name"`
	ChannelType string              `json:"channel_type"`
	IsDistinct  bool                `json:"is_distinct"`
	Members     []JoinChannelMember `json:"members"`
	Ts          int64               `json:"ts"`
}

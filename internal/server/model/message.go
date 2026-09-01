package model

type EventCategory int

const (
	EventCategoryUser    EventCategory = 1
	EventCategoryChannel EventCategory = 2
	EventCategorySocial  EventCategory = 3
)

type RedisPublishMessage struct {
	Category       EventCategory `json:"category"`
	Event          WSEvent       `json:"event"`
	Message        interface{}   `json:"message"`
	SenderID       string        `json:"sender_id"`
	ExceptUserID   string        `json:"except_user_id,omitempty"`
	SpecificUserID string        `json:"specific_user_id,omitempty"`
	ChannelID      string        `json:"channel_id"`
	Topic          string        `json:"topic"`
	IsFromAPI      bool          `json:"is_from_api"`
}

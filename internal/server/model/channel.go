package model

import "github.com/HieuNguyenVV/book-hive/internal/pkg/database"

type Channel struct {
	ID                string           `json:"id" db:"id"`
	ChannelURL        string           `json:"channel_url" db:"channel_url"`
	CoverURL          *string          `json:"cover_url,omitempty" db:"cover_url"`
	Name              string           `json:"name" db:"name"`
	ChannelType       string           `json:"channel_type" db:"channel_type"`
	Metadata          database.JSONMap `json:"metadata" db:"metadata"`
	LastMessageID     *string          `json:"last_message_id,omitempty" db:"last_message_id"`
	LastReadMessageID *string          `json:"last_read_message_id,omitempty" db:"last_read_message_id"`
	IsPublic          bool             `json:"is_public" db:"is_public"`
	IsDistinct        bool             `json:"is_distinct" db:"is_distinct"`
	CreateBy          string           `json:"create_by" db:"create_by"`
	CreateAt          int64            `json:"create_at" db:"create_at"`
	UpdateBy          string           `json:"update_by" db:"update_by"`
	UpdateAt          int64            `json:"update_at" db:"update_at"`
}

type ListChannelsRequest struct {
	Search      *string `json:"search"`
	ChannelType *string `json:"channel_type"`
	IsPublic    *bool   `json:"is_public"`
	ChannelURL  *string `json:"channel_url"`
	Limit       *int    `json:"limit"`
	Offset      *int    `json:"offset"`
}

type CreateChannelRequest struct {
	CoverURL    string                 `json:"cover_url"`
	Name        string                 `json:"name"`
	ChannelType string                 `json:"channel_type"`
	Metadata    map[string]interface{} `json:"metadata"`
	IsPublic    bool                   `json:"is_public"`
}

type UpdateChannelRequest struct {
	CoverURL          *string                `json:"cover_url"`
	Name              *string                `json:"name"`
	ChannelType       *string                `json:"channel_type"`
	Metadata          map[string]interface{} `json:"metadata"`
	LastMessageID     *string                `json:"last_message_id"`
	LastReadMessageID *string                `json:"last_read_message_id"`
	IsPublic          *bool                  `json:"is_public"`
}

type ListChannelsResponse struct {
	Results  []*Channel `json:"results"`
	Total    int        `json:"total"`
	Previous string     `json:"previous"`
	Next     string     `json:"next"`
}

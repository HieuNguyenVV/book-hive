package dto

import "github.com/HieuNguyenVV/book-hive/internal/server/model"

type CreateChannelRequest struct {
	CoverURL    string                 `json:"cover_url"`
	Name        string                 `json:"name" binding:"required"`
	ChannelType string                 `json:"channel_type" binding:"required"`
	Metadata    map[string]interface{} `json:"metadata"`
	IsPublic    bool                   `json:"is_public"`
} // @name CreateChannelRequest

func (r *CreateChannelRequest) ToModel() model.CreateChannelRequest {
	return model.CreateChannelRequest{
		CoverURL:    r.CoverURL,
		Name:        r.Name,
		ChannelType: r.ChannelType,
		Metadata:    r.Metadata,
		IsPublic:    r.IsPublic,
	}
}

type CreateDirectChannelRequest struct {
	TargetUserID string `json:"target_user_id" binding:"required"`
} // @name CreateDirectChannelRequest

type UpdateChannelRequest struct {
	CoverURL          *string                `json:"cover_url"`
	Name              *string                `json:"name"`
	ChannelType       *string                `json:"channel_type"`
	Metadata          map[string]interface{} `json:"metadata"`
	LastMessageID     *string                `json:"last_message_id"`
	LastReadMessageID *string                `json:"last_read_message_id"`
	IsPublic          *bool                  `json:"is_public"`
} // @name UpdateChannelRequest

func (r *UpdateChannelRequest) ToModel() model.UpdateChannelRequest {
	return model.UpdateChannelRequest{
		CoverURL:          r.CoverURL,
		Name:              r.Name,
		ChannelType:       r.ChannelType,
		Metadata:          r.Metadata,
		LastMessageID:     r.LastMessageID,
		LastReadMessageID: r.LastReadMessageID,
		IsPublic:          r.IsPublic,
	}
}

type ChannelResponse struct {
	ID                string                 `json:"id"`
	ChannelURL        string                 `json:"channel_url"`
	CoverURL          string                 `json:"cover_url"`
	Name              string                 `json:"name"`
	ChannelType       string                 `json:"channel_type"`
	Metadata          map[string]interface{} `json:"metadata"`
	LastMessageID     string                 `json:"last_message_id"`
	LastReadMessageID string                 `json:"last_read_message_id"`
	IsPublic          bool                   `json:"is_public"`
	IsDistinct        bool                   `json:"is_distinct"`
	CreateBy          string                 `json:"create_by"`
	CreateAt          int64                  `json:"create_at"`
	UpdateBy          string                 `json:"update_by"`
	UpdateAt          int64                  `json:"update_at"`
} // @name ChannelResponse

func (r *ChannelResponse) FromModel(channel *model.Channel) {
	r.ID = channel.ID
	r.ChannelURL = channel.ChannelURL
	r.CoverURL = derefString(channel.CoverURL)
	r.Name = channel.Name
	r.ChannelType = channel.ChannelType
	r.Metadata = channel.Metadata.ToMap()
	r.LastMessageID = derefString(channel.LastMessageID)
	r.LastReadMessageID = derefString(channel.LastReadMessageID)
	r.IsPublic = channel.IsPublic
	r.IsDistinct = channel.IsDistinct
	r.CreateBy = channel.CreateBy
	r.CreateAt = channel.CreateAt
	r.UpdateBy = channel.UpdateBy
	r.UpdateAt = channel.UpdateAt
}

type ListChannelsRequest struct {
	Token       *string `form:"token"`
	Search      *string `form:"search"`
	ChannelType *string `form:"channel_type"`
	IsPublic    *bool   `form:"is_public"`
	ChannelURL  *string `form:"channel_url"`
	Limit       *int    `form:"limit"`
	Offset      *int    `form:"offset"`
} // @name ListChannelsRequest

type ListChannelsResponse struct {
	Results  []*ChannelResponse `json:"results"`
	Total    int                `json:"total"`
	Previous string             `json:"previous"`
	Next     string             `json:"next"`
} // @name ListChannelsResponse

func (r *ListChannelsResponse) FromModel(response *model.ListChannelsResponse) {
	r.Results = make([]*ChannelResponse, len(response.Results))
	for i, channel := range response.Results {
		item := &ChannelResponse{}
		item.FromModel(channel)
		r.Results[i] = item
	}
	r.Total = response.Total
	r.Previous = response.Previous
	r.Next = response.Next
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

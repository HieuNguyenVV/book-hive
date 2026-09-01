package dto

import "github.com/HieuNguyenVV/book-hive/internal/server/model"

type UserBriefResponse struct {
	ID        string `json:"id"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ChannelMemberResponse struct {
	ID        string            `json:"id"`
	ChannelID string            `json:"channel_id"`
	UserID    string            `json:"user_id"`
	Role      string            `json:"role"`
	State     string            `json:"state"`
	JoinedAt  int64             `json:"joined_at"`
	User      UserBriefResponse `json:"user"`
}

type MessageResponse struct {
	ID          string            `json:"id"`
	ChannelID   string            `json:"channel_id"`
	Content     string            `json:"content"`
	MessageType string            `json:"message_type"`
	CustomType  string            `json:"custom_type"`
	ReqID       *string           `json:"req_id,omitempty"`
	CreatedAt   int64             `json:"created_at"`
	UpdatedAt   int64             `json:"updated_at"`
	Sender      UserBriefResponse `json:"sender"`
}

type ChannelDetailResponse struct {
	Channel     ChannelResponse          `json:"channel"`
	Members     []*ChannelMemberResponse `json:"members"`
	Messages    []*MessageResponse       `json:"messages"`
	LastMessage *MessageResponse         `json:"last_message,omitempty"`
}

type MyChannelItemResponse struct {
	Channel     ChannelResponse          `json:"channel"`
	Members     []*ChannelMemberResponse `json:"members"`
	LastMessage *MessageResponse         `json:"last_message,omitempty"`
	OtherUser   *UserBriefResponse       `json:"other_user,omitempty"`
}

type ListMyChannelsResponse struct {
	Results  []*MyChannelItemResponse `json:"results"`
	Total    int                      `json:"total"`
	Previous string                   `json:"previous"`
	Next     string                   `json:"next"`
}

type ListChannelMessagesResponse struct {
	Results  []*MessageResponse `json:"results"`
	Total    int                `json:"total"`
	Previous string             `json:"previous"`
	Next     string             `json:"next"`
}

type GetChannelDetailQuery struct {
	MessageLimit  *int `form:"message_limit"`
	MessageOffset *int `form:"message_offset"`
}

func userBriefFromModel(user model.UserBrief) UserBriefResponse {
	return UserBriefResponse{
		ID:        user.ID,
		FullName:  user.FullName,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}
}

func memberFromModel(member *model.ChannelMemberDetail) *ChannelMemberResponse {
	if member == nil {
		return nil
	}
	return &ChannelMemberResponse{
		ID:        member.ID,
		ChannelID: member.ChannelID,
		UserID:    member.UserID,
		Role:      member.Role,
		State:     member.State,
		JoinedAt:  member.JoinedAt,
		User:      userBriefFromModel(member.User),
	}
}

func messageFromModel(message *model.MessageWithSender) *MessageResponse {
	if message == nil {
		return nil
	}
	return &MessageResponse{
		ID:          message.ID,
		ChannelID:   message.ChannelID,
		Content:     message.Content,
		MessageType: message.MessageType,
		CustomType:  message.CustomType,
		ReqID:       message.ReqID,
		CreatedAt:   message.CreatedAt,
		UpdatedAt:   message.UpdatedAt,
		Sender:      userBriefFromModel(message.Sender),
	}
}

func ChannelDetailFromModel(detail *model.ChannelDetail) ChannelDetailResponse {
	response := ChannelDetailResponse{
		Members:  make([]*ChannelMemberResponse, len(detail.Members)),
		Messages: make([]*MessageResponse, len(detail.Messages)),
	}
	if detail.Channel != nil {
		channel := ChannelResponse{}
		channel.FromModel(detail.Channel)
		response.Channel = channel
	}
	for i, member := range detail.Members {
		response.Members[i] = memberFromModel(member)
	}
	for i, message := range detail.Messages {
		response.Messages[i] = messageFromModel(message)
	}
	response.LastMessage = messageFromModel(detail.LastMessage)
	return response
}

func ListMyChannelsFromModel(response *model.ListMyChannelsResponse, previous, next string) ListMyChannelsResponse {
	out := ListMyChannelsResponse{
		Total:    response.Total,
		Previous: previous,
		Next:     next,
		Results:  make([]*MyChannelItemResponse, len(response.Results)),
	}
	for i, item := range response.Results {
		channel := ChannelResponse{}
		channel.FromModel(item.Channel)
		members := make([]*ChannelMemberResponse, len(item.Members))
		for j, member := range item.Members {
			members[j] = memberFromModel(member)
		}
		out.Results[i] = &MyChannelItemResponse{
			Channel:     channel,
			Members:     members,
			LastMessage: messageFromModel(item.LastMessage),
		}
		if item.OtherUser != nil {
			otherUser := userBriefFromModel(*item.OtherUser)
			out.Results[i].OtherUser = &otherUser
		}
	}
	return out
}

func ListMessagesFromModel(response *model.ListChannelMessagesResponse, previous, next string) ListChannelMessagesResponse {
	out := ListChannelMessagesResponse{
		Total:    response.Total,
		Previous: previous,
		Next:     next,
		Results:  make([]*MessageResponse, len(response.Results)),
	}
	for i, message := range response.Results {
		out.Results[i] = messageFromModel(message)
	}
	return out
}

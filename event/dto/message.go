package dto

import (
	"github.com/google/uuid"
	"github.com/guregu/null"
	"time"
)

type EmbedMessageResponse struct {
	Id            int64   `json:"id"`
	ChatId        *int64  `json:"chatId"`
	ChatName      *string `json:"chatName"`
	Text          string  `json:"text"`
	Owner         *User   `json:"owner"`
	EmbedType     string  `json:"embedType"`
	IsParticipant bool    `json:"isParticipant"`
}

type Reaction struct {
	Count    int64  `json:"count"`
	Users    []*User `json:"users"`
	Reaction string `json:"reaction"`
}

type DisplayMessageDto struct {
	Id             int64                 `json:"id"`
	Text           string                `json:"text"`
	ChatId         int64                 `json:"chatId"`
	OwnerId        int64                 `json:"ownerId"`
	CreateDateTime time.Time             `json:"createDateTime"`
	EditDateTime   null.Time             `json:"editDateTime"`
	Owner          *User                 `json:"owner"`
	CanEdit        bool                  `json:"canEdit"`
	CanDelete      bool                  `json:"canDelete"`
	FileItemUuid   *uuid.UUID            `json:"fileItemUuid"`
	EmbedMessage   *EmbedMessageResponse `json:"embedMessage"`
	Pinned         bool                  `json:"pinned"`
	BlogPost       bool                  `json:"blogPost"`
	PinnedPromoted *bool                 `json:"pinnedPromoted"`
	Reactions []Reaction				 `json:"reactions"`
}

type MessageDeletedDto struct {
	Id     int64 `json:"id"`
	ChatId int64 `json:"chatId"`
}

type UserTypingNotification struct {
	Login         string `json:"login"`
	ParticipantId int64  `json:"participantId"`
}

type MessageBroadcastNotification struct {
	Login  string `json:"login"`
	UserId int64  `json:"userId"`
	Text   string `json:"text"`
}

type AllUnreadMessages struct {
	MessagesCount int64 `json:"allUnreadMessages"`
}

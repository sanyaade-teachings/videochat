// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"time"

	"github.com/google/uuid"
)

type AllUnreadMessages struct {
	AllUnreadMessages int64 `json:"allUnreadMessages"`
}

type ChatDeletedDto struct {
	ID int64 `json:"id"`
}

type ChatDto struct {
	ID                  int64                   `json:"id"`
	Name                string                  `json:"name"`
	Avatar              *string                 `json:"avatar"`
	AvatarBig           *string                 `json:"avatarBig"`
	ShortInfo           *string                 `json:"shortInfo"`
	LastUpdateDateTime  time.Time               `json:"lastUpdateDateTime"`
	ParticipantIds      []int64                 `json:"participantIds"`
	CanEdit             *bool                   `json:"canEdit"`
	CanDelete           *bool                   `json:"canDelete"`
	CanLeave            *bool                   `json:"canLeave"`
	UnreadMessages      int64                   `json:"unreadMessages"`
	CanBroadcast        bool                    `json:"canBroadcast"`
	CanVideoKick        bool                    `json:"canVideoKick"`
	CanChangeChatAdmins bool                    `json:"canChangeChatAdmins"`
	TetATet             bool                    `json:"tetATet"`
	CanAudioMute        bool                    `json:"canAudioMute"`
	Participants        []*ParticipantWithAdmin `json:"participants"`
	ParticipantsCount   int                     `json:"participantsCount"`
	CanResend           bool                    `json:"canResend"`
	AvailableToSearch   bool                    `json:"availableToSearch"`
	IsResultFromSearch  *bool                   `json:"isResultFromSearch"`
	Pinned              bool                    `json:"pinned"`
	Blog                bool                    `json:"blog"`
}

type ChatEvent struct {
	EventType             string                        `json:"eventType"`
	MessageEvent          *DisplayMessageDto            `json:"messageEvent"`
	MessageDeletedEvent   *MessageDeletedDto            `json:"messageDeletedEvent"`
	UserTypingEvent       *UserTypingDto                `json:"userTypingEvent"`
	MessageBroadcastEvent *MessageBroadcastNotification `json:"messageBroadcastEvent"`
	PreviewCreatedEvent   *PreviewCreatedEvent          `json:"previewCreatedEvent"`
	ParticipantsEvent     []*ParticipantWithAdmin       `json:"participantsEvent"`
	PromoteMessageEvent   *PinnedMessageEvent           `json:"promoteMessageEvent"`
	FileEvent             *WrappedFileInfoDto           `json:"fileEvent"`
}

type ChatUnreadMessageChanged struct {
	ChatID             int64     `json:"chatId"`
	UnreadMessages     int64     `json:"unreadMessages"`
	LastUpdateDateTime time.Time `json:"lastUpdateDateTime"`
}

type DisplayMessageDto struct {
	ID             int64                 `json:"id"`
	Text           string                `json:"text"`
	ChatID         int64                 `json:"chatId"`
	OwnerID        int64                 `json:"ownerId"`
	CreateDateTime time.Time             `json:"createDateTime"`
	EditDateTime   *time.Time            `json:"editDateTime"`
	Owner          *Participant          `json:"owner"`
	CanEdit        bool                  `json:"canEdit"`
	CanDelete      bool                  `json:"canDelete"`
	FileItemUUID   *uuid.UUID            `json:"fileItemUuid"`
	EmbedMessage   *EmbedMessageResponse `json:"embedMessage"`
	Pinned         bool                  `json:"pinned"`
	BlogPost       bool                  `json:"blogPost"`
	PinnedPromoted *bool                 `json:"pinnedPromoted"`
}

type EmbedMessageResponse struct {
	ID            int64        `json:"id"`
	ChatID        *int64       `json:"chatId"`
	ChatName      *string      `json:"chatName"`
	Text          string       `json:"text"`
	Owner         *Participant `json:"owner"`
	EmbedType     string       `json:"embedType"`
	IsParticipant bool         `json:"isParticipant"`
}

type FileInfoDto struct {
	ID             string       `json:"id"`
	Filename       string       `json:"filename"`
	URL            string       `json:"url"`
	PublicURL      *string      `json:"publicUrl"`
	PreviewURL     *string      `json:"previewUrl"`
	Size           int64        `json:"size"`
	CanDelete      bool         `json:"canDelete"`
	CanEdit        bool         `json:"canEdit"`
	CanShare       bool         `json:"canShare"`
	LastModified   time.Time    `json:"lastModified"`
	OwnerID        int64        `json:"ownerId"`
	Owner          *Participant `json:"owner"`
	CanPlayAsVideo bool         `json:"canPlayAsVideo"`
	CanShowAsImage bool         `json:"canShowAsImage"`
	CanPlayAsAudio bool         `json:"canPlayAsAudio"`
}

type GlobalEvent struct {
	EventType                      string                          `json:"eventType"`
	ChatEvent                      *ChatDto                        `json:"chatEvent"`
	ChatDeletedEvent               *ChatDeletedDto                 `json:"chatDeletedEvent"`
	ParticipantEvent               *Participant                    `json:"participantEvent"`
	VideoUserCountChangedEvent     *VideoUserCountChangedDto       `json:"videoUserCountChangedEvent"`
	VideoRecordingChangedEvent     *VideoRecordingChangedDto       `json:"videoRecordingChangedEvent"`
	VideoCallInvitation            *VideoCallInvitationDto         `json:"videoCallInvitation"`
	VideoParticipantDialEvent      *VideoDialChanges               `json:"videoParticipantDialEvent"`
	UnreadMessagesNotification     *ChatUnreadMessageChanged       `json:"unreadMessagesNotification"`
	AllUnreadMessagesNotification  *AllUnreadMessages              `json:"allUnreadMessagesNotification"`
	NotificationEvent              *NotificationDto                `json:"notificationEvent"`
	VideoCallScreenShareChangedDto *VideoCallScreenShareChangedDto `json:"videoCallScreenShareChangedDto"`
}

type MessageBroadcastNotification struct {
	Login  string `json:"login"`
	UserID int64  `json:"userId"`
	Text   string `json:"text"`
}

type MessageDeletedDto struct {
	ID     int64 `json:"id"`
	ChatID int64 `json:"chatId"`
}

type NotificationDto struct {
	ID               int64     `json:"id"`
	ChatID           int64     `json:"chatId"`
	MessageID        *int64    `json:"messageId"`
	NotificationType string    `json:"notificationType"`
	Description      string    `json:"description"`
	CreateDateTime   time.Time `json:"createDateTime"`
	ByUserID         int64     `json:"byUserId"`
	ByLogin          string    `json:"byLogin"`
	ChatTitle        string    `json:"chatTitle"`
}

type Participant struct {
	ID        int64   `json:"id"`
	Login     string  `json:"login"`
	Avatar    *string `json:"avatar"`
	ShortInfo *string `json:"shortInfo"`
}

type ParticipantWithAdmin struct {
	ID        int64   `json:"id"`
	Login     string  `json:"login"`
	Avatar    *string `json:"avatar"`
	Admin     bool    `json:"admin"`
	ShortInfo *string `json:"shortInfo"`
}

type PinnedMessageEvent struct {
	Message    *DisplayMessageDto `json:"message"`
	TotalCount int64              `json:"totalCount"`
}

type PreviewCreatedEvent struct {
	ID            string  `json:"id"`
	URL           string  `json:"url"`
	PreviewURL    *string `json:"previewUrl"`
	AType         *string `json:"aType"`
	CorrelationID *string `json:"correlationId"`
}

type UserStatusEvent struct {
	UserID    int64  `json:"userId"`
	Online    *bool  `json:"online"`
	IsInVideo *bool  `json:"isInVideo"`
	EventType string `json:"eventType"`
}

type UserTypingDto struct {
	Login         string `json:"login"`
	ParticipantID int64  `json:"participantId"`
}

type VideoCallInvitationDto struct {
	ChatID   int64  `json:"chatId"`
	ChatName string `json:"chatName"`
}

type VideoCallScreenShareChangedDto struct {
	ChatID          int64 `json:"chatId"`
	HasScreenShares bool  `json:"hasScreenShares"`
}

type VideoDialChanged struct {
	UserID int64 `json:"userId"`
	Status bool  `json:"status"`
}

type VideoDialChanges struct {
	ChatID int64               `json:"chatId"`
	Dials  []*VideoDialChanged `json:"dials"`
}

type VideoRecordingChangedDto struct {
	RecordInProgress bool  `json:"recordInProgress"`
	ChatID           int64 `json:"chatId"`
}

type VideoUserCountChangedDto struct {
	UsersCount int64 `json:"usersCount"`
	ChatID     int64 `json:"chatId"`
}

type WrappedFileInfoDto struct {
	FileInfoDto  *FileInfoDto `json:"fileInfoDto"`
	Count        int64        `json:"count"`
	FileItemUUID *uuid.UUID   `json:"fileItemUuid"`
}

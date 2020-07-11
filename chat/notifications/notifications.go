package notifications

import (
	"encoding/json"
	"fmt"
	"github.com/centrifugal/centrifuge"
	"github.com/labstack/echo/v4"
	"nkonev.name/chat/auth"
	"nkonev.name/chat/handlers/dto"
	. "nkonev.name/chat/logger"
	"nkonev.name/chat/utils"
)

type Notifications interface {
	NotifyAboutNewChat(c echo.Context, newChatDto *dto.ChatDto, userPrincipalDto *auth.AuthResult)
	NotifyAboutNewMessage(c echo.Context, chatId int64, message *dto.DisplayMessageDto, userPrincipalDto *auth.AuthResult)
}

type notifictionsImpl struct {
	centrifuge *centrifuge.Node
}

func NewNotifications(node *centrifuge.Node) Notifications {
	return &notifictionsImpl{
		centrifuge: node,
	}
}

// created or modified
type CentrifugeNotification struct {
	Payload   interface{} `json:"payload"`
	EventType string      `json:"type"`
}

func (not *notifictionsImpl) NotifyAboutNewChat(c echo.Context, newChatDto *dto.ChatDto, userPrincipalDto *auth.AuthResult) {
	participantsForNotify := newChatDto.ParticipantIds
	for _, participantId := range participantsForNotify {
		participantChannel := not.centrifuge.PersonalChannel(utils.Int64ToString(participantId))
		GetLogEntry(c.Request()).Infof("Sending notification about create the chat to participantChannel: %v", participantChannel)

		notification := CentrifugeNotification{
			Payload:   *newChatDto,
			EventType: "chat_created",
		}
		if marshalledBytes, err2 := json.Marshal(notification); err2 != nil {
			GetLogEntry(c.Request()).Errorf("error during marshalling chat created notification: %s", err2)
		} else {
			_, err := not.centrifuge.Publish(participantChannel, marshalledBytes)
			if err != nil {
				GetLogEntry(c.Request()).Errorf("error publishing to personal channel: %s", err)
			}
		}
	}
}

func (not *notifictionsImpl) NotifyAboutNewMessage(c echo.Context, chatId int64, message *dto.DisplayMessageDto, userPrincipalDto *auth.AuthResult) {
	chatChannel := fmt.Sprintf("%v%v", utils.CHANNEL_PREFIX_CHAT, chatId)
	notification := CentrifugeNotification{
		Payload:   *message,
		EventType: "message_created",
	}
	if marshalledBytes, err2 := json.Marshal(notification); err2 != nil {
		GetLogEntry(c.Request()).Errorf("error during marshalling chat created notification: %s", err2)
	} else {
		_, err := not.centrifuge.Publish(chatChannel, marshalledBytes)
		if err != nil {
			GetLogEntry(c.Request()).Errorf("error publishing to personal channel: %s", err)
		}
	}
}

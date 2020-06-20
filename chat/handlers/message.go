package handlers

import (
	"errors"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/guregu/null"
	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
	"net/http"
	"nkonev.name/chat/auth"
	"nkonev.name/chat/db"
	. "nkonev.name/chat/logger"
	"nkonev.name/chat/utils"
	"time"
)

type NewMessageDto struct {
	Text   string `json:"text"`
	ChatId int64  `json:"chatId"`
}

type EditMessageDto struct {
	Id     int64  `json:"id"`
	Text   string `json:"text"`
	ChatId int64  `json:"chatId"`
}

type DisplayMessageDto struct {
	Id             int64     `json:"id"`
	Text           string    `json:"text"`
	ChatId         int64     `json:"chatId"`
	OwnerId        int64     `json:"ownerId"`
	CreateDateTime time.Time `json:"createDateTime"`
	EditDateTime   null.Time `json:"editDateTime"`
}

func GetMessages(dbR db.DB) func(c echo.Context) error {
	return func(c echo.Context) error {
		var userPrincipalDto, ok = c.Get(utils.USER_PRINCIPAL_DTO).(*auth.AuthResult)
		if !ok {
			GetLogEntry(c.Request()).Errorf("Error during getting auth context")
			return errors.New("Error during getting auth context")
		}

		page := utils.FixPageString(c.QueryParam("page"))
		size := utils.FixSizeString(c.QueryParam("size"))
		offset := utils.GetOffset(page, size)

		chatIdString := c.Param("id")
		chatId, err := utils.ParseInt64(chatIdString)
		if err != nil {
			return err
		}

		if messages, err := dbR.GetMessages(chatId, userPrincipalDto.UserId, size, offset); err != nil {
			GetLogEntry(c.Request()).Errorf("Error get messages from db %v", err)
			return err
		} else {
			messageDtos := make([]*DisplayMessageDto, 0)
			for _, c := range messages {
				messageDtos = append(messageDtos, convertToMessageDto(c))
			}
			GetLogEntry(c.Request()).Infof("Successfully returning %v messages", len(messageDtos))
			return c.JSON(200, messageDtos)
		}
	}
}

func GetMessage(dbR db.DB) func(c echo.Context) error {
	return func(c echo.Context) error {
		var userPrincipalDto, ok = c.Get(utils.USER_PRINCIPAL_DTO).(*auth.AuthResult)
		if !ok {
			GetLogEntry(c.Request()).Errorf("Error during getting auth context")
			return errors.New("Error during getting auth context")
		}

		chatIdString := c.Param("id")
		chatId, err := utils.ParseInt64(chatIdString)
		if err != nil {
			return err
		}

		messageIdString := c.Param("messageId")
		messageId, err := utils.ParseInt64(messageIdString)
		if err != nil {
			return err
		}

		if message, err := dbR.GetMessage(chatId, userPrincipalDto.UserId, messageId); err != nil {
			GetLogEntry(c.Request()).Errorf("Error get messages from db %v", err)
			return err
		} else {
			if message == nil {
				return c.NoContent(http.StatusNotFound)
			}
			messageDto := convertToMessageDto(message)
			GetLogEntry(c.Request()).Infof("Successfully returning message %v", messageDto)
			return c.JSON(200, messageDto)
		}
	}
}

func convertToMessageDto(dbMessage *db.Message) *DisplayMessageDto {
	return &DisplayMessageDto{
		Id:             dbMessage.Id,
		Text:           dbMessage.Text,
		ChatId:         dbMessage.ChatId,
		OwnerId:        dbMessage.OwnerId,
		CreateDateTime: dbMessage.CreateDateTime,
		EditDateTime:   dbMessage.EditDateTime,
	}
}

type CreateMessageDto struct {
	Text string `json:"text"`
}

func (a *CreateMessageDto) Validate() error {
	return validation.ValidateStruct(a, validation.Field(&a.Text, validation.Required, validation.Length(1, 1024*1024)))
}

func PostMessage(dbR db.DB, policy *bluemonday.Policy) func(c echo.Context) error {
	return func(c echo.Context) error {
		var bindTo = new(CreateMessageDto)
		if err := c.Bind(bindTo); err != nil {
			GetLogEntry(c.Request()).Errorf("Error during binding to dto %v", err)
			return err
		}

		if valid, err := ValidateAndRespondError(c, bindTo); err != nil || !valid {
			return err
		}

		var userPrincipalDto, ok = c.Get(utils.USER_PRINCIPAL_DTO).(*auth.AuthResult)
		if !ok {
			GetLogEntry(c.Request()).Errorf("Error during getting auth context")
			return errors.New("Error during getting auth context")
		}

		chatId, err := GetPathParamAsInt64(c, "id")
		if err != nil {
			return err
		}

		errOuter := db.Transact(dbR, func(tx *db.Tx) error {
			if participant, err := tx.IsParticipant(userPrincipalDto.UserId, chatId); err != nil {
				return err
			} else if !participant {
				return c.JSON(http.StatusBadRequest, &utils.H{"message": "You are not allowed to write to this chat"})
			}
			id, err := tx.CreateMessage(convertToCreatableMessage(bindTo, userPrincipalDto, chatId, policy))
			if err != nil {
				return err
			}
			err = tx.AddMessageRead(id, userPrincipalDto.UserId)
			if err != nil {
				return err
			}
			return c.JSON(http.StatusCreated, &utils.H{"id": id})
		})
		if errOuter != nil {
			GetLogEntry(c.Request()).Errorf("Error during act transaction %v", errOuter)
		}
		return errOuter
	}
}

func convertToCreatableMessage(dto *CreateMessageDto, authPrincipal *auth.AuthResult, chatId int64, policy *bluemonday.Policy) *db.Message {
	return &db.Message{
		Text:    policy.Sanitize(dto.Text),
		ChatId:  chatId,
		OwnerId: authPrincipal.UserId,
	}
}

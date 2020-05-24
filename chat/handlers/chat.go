package handlers

import (
	"github.com/labstack/echo/v4"
	"nkonev.name/chat/auth"
	"nkonev.name/chat/db"
	. "nkonev.name/chat/logger"
	"nkonev.name/chat/utils"
)

type ChatDto struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type CreateChatDto struct {
	Name string `json:"name"`
}

func GetChats(db db.DB) func(c echo.Context) error {
	return func(c echo.Context) error {
		if tx, err := db.Begin(); err != nil {
			GetLogEntry(c.Request()).Errorf("Error during open transaction %v", err)
			return err
		} else {
			var userPrincipalDto, ok = c.Get(utils.USER_PRINCIPAL_DTO).(*auth.AuthResult)
			if !ok {
				GetLogEntry(c.Request()).Errorf("Error during getting auth context")
				tx.SafeRollback()
			}

			page := utils.FixPageString(c.QueryParam("page"))
			size := utils.FixSizeString(c.QueryParam("size"))
			offset := utils.GetOffset(page, size)

			if chats, err := tx.GetChats(userPrincipalDto.UserId, size, offset); err != nil {
				GetLogEntry(c.Request()).Errorf("Error get chats from db %v", err)
				tx.SafeRollback()
				return err
			} else {
				chatDtos := make([]*ChatDto, 0)
				for _, c := range chats {
					chatDtos = append(chatDtos, convertToDto(c))
				}
				if err := tx.Commit(); err != nil {
					GetLogEntry(c.Request()).Errorf("Error during commit transaction %v", err)
					return err
				}
				GetLogEntry(c.Request()).Infof("Successfully returning %v chats", len(chatDtos))
				return c.JSON(200, chatDtos)
			}
		}
	}
}

func convertToDto(c *db.Chat) *ChatDto {
	return &ChatDto{
		Id:   c.Id,
		Name: c.Title,
	}
}

func CreateChat(db db.DB) func(c echo.Context) error {
	return func(c echo.Context) error {
		if tx, err := db.Begin(); err != nil {
			GetLogEntry(c.Request()).Errorf("Error during open transaction %v", err)
			return err
		} else {
			var userPrincipalDto, ok = c.Get(utils.USER_PRINCIPAL_DTO).(*auth.AuthResult)
			if !ok {
				GetLogEntry(c.Request()).Errorf("Error during getting auth context")
				tx.SafeRollback()
			}

			var bindTo = new(CreateChatDto)
			if err := c.Bind(bindTo); err != nil {
				GetLogEntry(c.Request()).Errorf("Error during binding to dto %v", err)
				tx.SafeRollback()
				return err
			}

			if err := tx.CreateChat(convertToCreatableChat(bindTo, userPrincipalDto)); err != nil {
				GetLogEntry(c.Request()).Errorf("Error get chats from db %v", err)
				tx.SafeRollback()
				return err
			} else {
				if err := tx.Commit(); err != nil {
					GetLogEntry(c.Request()).Errorf("Error during commit transaction %v", err)
					return err
				}
				GetLogEntry(c.Request()).Infof("Successfully created chat %v", bindTo)
				return c.NoContent(200)
			}
		}
	}
}

func convertToCreatableChat(d *CreateChatDto, a *auth.AuthResult) *db.Chat {
	return &db.Chat{
		Title:   d.Name,
		OwnerId: a.UserId,
	}
}

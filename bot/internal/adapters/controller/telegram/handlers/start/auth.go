package start

import (
	"context"
	"errors"
	"strings"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/redis/go-redis/v9"
	tele "gopkg.in/telebot.v3"
)

func (h Handler) auth(c tele.Context, authCode string) error {
	code, err := h.codesStorage.Get(c.Sender().ID)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			h.logger.Errorf("(user: %d) error while getting auth code from redis: %v", c.Sender().ID, err)
			return c.Send(
				h.layout.Text(c, "technical_issues", err.Error()),
				h.layout.Markup(c, "core:hide"),
			)
		}
		return c.Send(
			h.layout.Text(c, "session_expire"),
			h.layout.Markup(c, "core:hide"),
		)
	}

	if authCode != code.Code {
		return c.Send(
			h.layout.Text(c, "something_went_wrong"),
			h.layout.Markup(c, "core:hide"),
		)
	}

	data := strings.Split(code.CodeContext, ";")
	email, fio := data[0], data[1]

	newUser := entity.User{
		ID:    c.Sender().ID,
		Role:  entity.Student,
		Email: email,
		FIO:   fio,
	}

	_, err = h.userService.Create(context.Background(), newUser)
	if err != nil {
		h.logger.Errorf("(user: %d) error while create new user: %v", c.Sender().ID, err)
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "core:hide"),
		)
	}
	h.logger.Infof("(user: %d) new user created(role: %s)", c.Sender().ID, newUser.Role)

	h.codesStorage.Clear(c.Sender().ID)
	h.codesStorage.Clear(c.Sender().ID)
	h.emailsStorage.Clear(c.Sender().ID)

	return h.menuHandler.SendMenu(c)
}

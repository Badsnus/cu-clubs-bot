package start

import (
	"context"
	"errors"
	"log"

	"github.com/redis/go-redis/v9"
	tele "gopkg.in/telebot.v3"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/banner"
)

func (h Handler) changeRole(c tele.Context, authCode string) error {
	code, err := h.codesStorage.Get(c.Sender().ID)
	if err != nil {
		log.Println(err)
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
			h.layout.Text(c, "wrong_code"),
			h.layout.Markup(c, "core:hide"),
		)
	}

	err = h.userService.ChangeRole(
		context.Background(),
		c.Sender().ID,
		entity.Student,
		code.CodeContext.Email,
	)
	if err != nil {
		h.logger.Errorf("(user: %d) error while change role: %v", c.Sender().ID, err)
		return c.Edit(
			banner.PersonalAccount.Caption(h.layout.Text(c, "technical_issues")),
			h.layout.Markup(c, "core:hide"),
		)
	}

	err = h.codesStorage.Clear(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while clearing auth code from redis: %v", c.Sender().ID, err)
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "core:hide"),
		)
	}
	err = h.emailsStorage.Clear(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while clearing email from redis: %v", c.Sender().ID, err)
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "core:hide"),
		)
	}

	user, err := h.userService.Get(context.Background(), c.Sender().ID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while getting user from db: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "core:hide"),
		)
	}

	fio, err := user.ParseFIO()
	if err != nil {
		h.logger.Errorf("(user: %d) error parsing user fio: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "core:hide"),
		)
	}
	return c.Send(
		banner.PersonalAccount.Caption(h.layout.Text(c, "personal_account_text", struct {
			Name string
			Role string
		}{
			Name: fio.Name,
			Role: user.Role.String(),
		})),
		h.layout.Markup(c, "personalAccount:menu"),
	)
}

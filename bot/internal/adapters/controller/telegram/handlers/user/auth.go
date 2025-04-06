package user

import (
	"context"
	"errors"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/emails"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/banner"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/validator"
	"github.com/nlypage/intele/collector"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
)

func (h Handler) declinePersonalDataAgreement(c tele.Context) error {
	h.logger.Infof("(user: %d) decline personal data agreement", c.Sender().ID)
	return c.Edit(
		banner.Auth.Caption(h.layout.Text(c, "decline_personal_data_agreement_text")),
	)
}

func (h Handler) acceptPersonalDataAgreement(c tele.Context) error {
	h.logger.Infof("(user: %d) accept personal data agreement", c.Sender().ID)
	return c.Edit(
		banner.Auth.Caption(h.layout.Text(c, "auth_menu_text")),
		h.layout.Markup(c, "auth:menu"),
	)
}

func (h Handler) externalUserAuth(c tele.Context) error {
	h.logger.Infof("(user: %d) external user auth", c.Sender().ID)
	inputCollector := collector.New()
	_ = c.Edit(
		banner.Auth.Caption(h.layout.Text(c, "fio_request")),
		h.layout.Markup(c, "auth:backToMenu"),
	)
	inputCollector.Collect(c.Message())

	var (
		fio  string
		done bool
	)
	for {
		response, err := h.input.Get(context.Background(), c.Sender().ID, 0)
		if response.Message != nil {
			inputCollector.Collect(response.Message)
		}
		switch {
		case response.Canceled:
			_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true, ExcludeLast: true})
			return nil
		case err != nil:
			h.logger.Errorf("(user: %d) error while input fio: %v", c.Sender().ID, err)
			_ = inputCollector.Send(c,
				banner.Auth.Caption(h.layout.Text(c, "input_error", h.layout.Text(c, "fio_request"))),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case response.Message == nil:
			h.logger.Errorf("(user: %d) error while input fio: %v", c.Sender().ID, err)
			_ = inputCollector.Send(c,
				banner.Auth.Caption(h.layout.Text(c, "input_error", h.layout.Text(c, "fio_request"))),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case !validator.Fio(response.Message.Text, nil):
			_ = inputCollector.Send(c,
				banner.Auth.Caption(h.layout.Text(c, "invalid_user_fio")),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case validator.Fio(response.Message.Text, nil):
			fio = response.Message.Text
			_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
			done = true
		}
		if done {
			break
		}
	}

	user := entity.User{
		ID:   c.Sender().ID,
		Role: entity.ExternalUser,
		FIO:  fio,
	}
	_, err := h.userService.Create(context.Background(), user)
	if err != nil {
		h.logger.Errorf("(user: %d) error while creating new user: %v", c.Sender().ID, err)
		return c.Send(
			banner.Auth.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "auth:backToMenu"),
		)
	}
	h.logger.Infof("(user: %d) new user created(role: %s)", c.Sender().ID, user.Role)

	return h.menuHandler.SendMenu(c)
}

func (h Handler) grantUserAuth(c tele.Context) error {
	h.logger.Infof("(user: %d) grant user auth", c.Sender().ID)

	grantChatID := int64(viper.GetInt("bot.auth.grant-chat-id"))
	member, err := c.Bot().ChatMemberOf(&tele.Chat{ID: grantChatID}, &tele.User{ID: c.Sender().ID})
	if err != nil {
		h.logger.Errorf("(user: %d) error while verification user's membership in the grant chat: %v", c.Sender().ID, err)
		return c.Send(
			banner.Auth.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "core:hide"),
		)
	}

	if member.Role != tele.Creator && member.Role != tele.Administrator && member.Role != tele.Member {
		return c.Edit(
			banner.Auth.Caption(h.layout.Text(c, "grant_user_required")),
			h.layout.Markup(c, "auth:backToMenu"),
		)
	}

	inputCollector := collector.New()
	_ = c.Edit(
		banner.Auth.Caption(h.layout.Text(c, "fio_request")),
		h.layout.Markup(c, "auth:backToMenu"),
	)
	inputCollector.Collect(c.Message())

	var (
		fio  string
		done bool
	)
	for {
		response, errGet := h.input.Get(context.Background(), c.Sender().ID, 0)
		if response.Message != nil {
			inputCollector.Collect(response.Message)
		}
		switch {
		case response.Canceled:
			_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true, ExcludeLast: true})
			return nil
		case errGet != nil:
			h.logger.Errorf("(user: %d) error while input fio: %v", c.Sender().ID, errGet)
			_ = inputCollector.Send(c,
				banner.Auth.Caption(h.layout.Text(c, "input_error", h.layout.Text(c, "fio_request"))),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case response.Message == nil:
			h.logger.Errorf("(user: %d) error while input fio: %v", c.Sender().ID, errGet)
			_ = inputCollector.Send(c,
				banner.Auth.Caption(h.layout.Text(c, "input_error", h.layout.Text(c, "fio_request"))),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case !validator.Fio(response.Message.Text, nil):
			_ = inputCollector.Send(c,
				banner.Auth.Caption(h.layout.Text(c, "invalid_user_fio")),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case validator.Fio(response.Message.Text, nil):
			fio = response.Message.Text
			_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
			done = true
		}
		if done {
			break
		}
	}

	user := entity.User{
		ID:   c.Sender().ID,
		Role: entity.GrantUser,
		FIO:  fio,
	}
	_, err = h.userService.Create(context.Background(), user)
	if err != nil {
		h.logger.Errorf("(user: %d) error while creating new user: %v", c.Sender().ID, err)
		return c.Send(
			banner.Auth.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "auth:backToMenu"),
		)
	}
	h.logger.Infof("(user: %d) new user created(role: %s)", c.Sender().ID, user.Role)

	return h.menuHandler.SendMenu(c)
}

func (h Handler) studentAuth(c tele.Context) error {
	h.logger.Infof("(user: %d) student auth", c.Sender().ID)

	inputCollector := collector.New()
	_ = c.Edit(
		banner.Auth.Caption(h.layout.Text(c, "email_request")),
		h.layout.Markup(c, "auth:backToMenu"),
	)
	inputCollector.Collect(c.Message())

	var (
		email string
		done  bool
	)
	for {
		response, errGet := h.input.Get(context.Background(), c.Sender().ID, 0)
		if response.Message != nil {
			inputCollector.Collect(response.Message)
		}
		switch {
		case response.Canceled:
			_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true, ExcludeLast: true})
			return nil
		case errGet != nil:
			h.logger.Errorf("(user: %d) error while input email: %v", c.Sender().ID, errGet)
			_ = inputCollector.Send(c,
				banner.Auth.Caption(h.layout.Text(c, "input_error", h.layout.Text(c, "email_request"))),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case response.Message == nil:
			h.logger.Errorf("(user: %d) error while input email: %v", c.Sender().ID, errGet)
			_ = inputCollector.Send(c,
				banner.Auth.Caption(h.layout.Text(c, "input_error", h.layout.Text(c, "email_request"))),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case !validator.Email(response.Message.Text, nil):
			_ = inputCollector.Send(c,
				banner.Auth.Caption(h.layout.Text(c, "invalid_email")),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		case validator.Email(response.Message.Text, nil):
			email = response.Message.Text
			done = true
		}
		if done {
			break
		}
	}

	canResend, err := h.codesStorage.GetCanResend(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while getting auth code from redis: %v", c.Sender().ID, err)
		return c.Send(
			banner.Auth.Caption(h.layout.Text(c, "technical_issues", err.Error())),
		)
	}
	var data, code string
	if canResend {
		data, code, err = h.userService.SendAuthCode(context.Background(), email)
		if err != nil {
			h.logger.Errorf("(user: %d) error while sending auth code: %v", c.Sender().ID, err)
			return c.Send(
				banner.Auth.Caption(h.layout.Text(c, "technical_issues", err.Error())),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		}

		h.emailsStorage.Set(c.Sender().ID, email, "", viper.GetDuration("bot.session.email-ttl"))
		h.codesStorage.Set(c.Sender().ID, code, data, viper.GetDuration("bot.session.auth-ttl"))
		h.codesStorage.SetCanResend(c.Sender().ID, true, viper.GetDuration("bot.session.resend-ttl"))

		h.logger.Infof("(user: %d) auth code sent on %s", c.Sender().ID, email)

		return c.Send(
			banner.Auth.Caption(h.layout.Text(c, "email_auth_link_sent")),
			h.layout.Markup(c, "auth:resendMenu"),
		)
	}

	return c.Send(
		banner.Auth.Caption(h.layout.Text(c, "resend_timeout")),
		h.layout.Markup(c, "auth:resendMenu"),
	)
}

func (h Handler) backToAuthMenu(c tele.Context) error {
	return c.Edit(
		banner.Auth.Caption(h.layout.Text(c, "auth_menu_text")),
		h.layout.Markup(c, "auth:menu"),
	)
}

func (h Handler) resendEmailConfirmationCode(c tele.Context) error {
	h.logger.Infof("(user: %d) resend auth code", c.Sender().ID)

	canResend, err := h.codesStorage.GetCanResend(c.Sender().ID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while getting auth code from redis: %v", c.Sender().ID, err)
		return c.Send(
			banner.Auth.Caption(h.layout.Text(c, "technical_issues", err.Error())),
		)
	}

	var data, code string
	var email emails.Email
	if canResend {
		email, err = h.emailsStorage.Get(c.Sender().ID)
		if err != nil && !errors.Is(err, redis.Nil) {
			h.logger.Errorf("(user: %d) error while getting user email from redis: %v", c.Sender().ID, err)
			return c.Send(
				banner.Auth.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			)
		}

		if errors.Is(err, redis.Nil) {
			return c.Send(
				banner.Auth.Caption(h.layout.Text(c, "session_expire")),
			)
		}

		data, code, err = h.userService.SendAuthCode(context.Background(), email.Email)
		if err != nil {
			h.logger.Errorf("(user: %d) error while sending auth code: %v", c.Sender().ID, err)
			return c.Send(
				banner.Auth.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			)
		}

		h.emailsStorage.Set(c.Sender().ID, email.Email, "", viper.GetDuration("bot.session.email-ttl"))
		h.codesStorage.Set(c.Sender().ID, code, data, viper.GetDuration("bot.session.auth-ttl"))
		h.codesStorage.SetCanResend(c.Sender().ID, true, viper.GetDuration("bot.session.resend-ttl"))

		h.logger.Infof("(user: %d) auth code sent on %s", c.Sender().ID, email.Email)

		return c.Edit(
			banner.Auth.Caption(h.layout.Text(c, "email_auth_link_resent")),
			h.layout.Markup(c, "auth:resendMenu"),
		)
	}

	return c.Edit(
		banner.Auth.Caption(h.layout.Text(c, "resend_timeout")),
		h.layout.Markup(c, "auth:resendMenu"),
	)
}

func (h Handler) AuthSetup(group *tele.Group) {
	group.Handle(h.layout.Callback("auth:personalData:accept"), h.acceptPersonalDataAgreement)
	group.Handle(h.layout.Callback("auth:personalData:decline"), h.declinePersonalDataAgreement)
	group.Handle(h.layout.Callback("auth:external_user"), h.externalUserAuth)
	group.Handle(h.layout.Callback("auth:grant_user"), h.grantUserAuth)
	group.Handle(h.layout.Callback("auth:student"), h.studentAuth)
	group.Handle(h.layout.Callback("auth:resend_email"), h.resendEmailConfirmationCode)
	group.Handle(h.layout.Callback("auth:back_to_menu"), h.backToAuthMenu)
}

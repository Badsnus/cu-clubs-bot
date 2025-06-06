package start

import (
	"context"
	"errors"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/banner"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"
	"gorm.io/gorm"
	"strings"
	"time"

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
	h.emailsStorage.Clear(c.Sender().ID)

	user, err := h.userService.Get(context.Background(), c.Sender().ID)
	if err != nil {
		return c.Send(
			banner.Auth.Caption(h.layout.Text(c, "auth_required")),
			h.layout.Markup(c, "core:hide"),
		)
	}

	eventID, err := h.eventsStorage.GetEventID(c.Sender().ID, "before-reg-event-id")
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			h.logger.Errorf("(user: %d) error while getting event ID: %v", c.Sender().ID, err)
			return c.Send(
				banner.Auth.Caption(h.layout.Text(c, "technical_issues", err.Error())),
				h.layout.Markup(c, "auth:backToMenu"),
			)
		}

		return h.menuHandler.SendMenu(c)
	}

	event, err := h.eventService.Get(context.Background(), eventID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get event: %v", c.Sender().ID, err)
		return c.Send(
			banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	club, err := h.clubService.Get(context.Background(), event.ClubID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get club: %v", c.Sender().ID, err)
		return c.Send(
			banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	participantsCount, err := h.eventParticipantService.CountByEventID(context.Background(), eventID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get participants count: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "mainMenu:back"),
		)
	}

	var registered bool
	_, errGetParticipant := h.eventParticipantService.Get(context.Background(), eventID, c.Sender().ID)
	if errGetParticipant != nil {
		if !errors.Is(errGetParticipant, gorm.ErrRecordNotFound) {
			h.logger.Errorf("(user: %d) error while get participant: %v", c.Sender().ID, errGetParticipant)
			return c.Send(
				banner.Events.Caption(h.layout.Text(c, "technical_issues", errGetParticipant.Error())),
				h.layout.Markup(c, "mainMenu:back"),
			)
		}
	} else {
		registered = true
	}

	endTime := event.EndTime.In(location.Location()).Format("02.01.2006 15:04")
	if event.EndTime.Year() == 1 {
		endTime = ""
	}

	var maxRegistrationEnd time.Time
	if user.Role == entity.Student {
		maxRegistrationEnd = event.RegistrationEnd
	} else {
		if event.RegistrationEnd.Before(utils.GetMaxRegisteredEndTime(event.StartTime)) {
			maxRegistrationEnd = event.RegistrationEnd
		} else {
			maxRegistrationEnd = utils.GetMaxRegisteredEndTime(event.StartTime)
		}
	}

	_ = c.Send(
		banner.Events.Caption(h.layout.Text(c, "event_text", struct {
			Name                  string
			ClubName              string
			Description           string
			Location              string
			StartTime             string
			EndTime               string
			RegistrationEnd       string
			MaxParticipants       int
			ParticipantsCount     int
			AfterRegistrationText string
			IsRegistered          bool
		}{
			Name:                  event.Name,
			ClubName:              club.Name,
			Description:           event.Description,
			Location:              event.Location,
			StartTime:             event.StartTime.In(location.Location()).Format("02.01.2006 15:04"),
			EndTime:               endTime,
			RegistrationEnd:       maxRegistrationEnd.In(location.Location()).Format("02.01.2006 15:04"),
			MaxParticipants:       event.MaxParticipants,
			ParticipantsCount:     participantsCount,
			AfterRegistrationText: event.AfterRegistrationText,
			IsRegistered:          registered,
		})),
		h.layout.Markup(c, "user:url:event", struct {
			ID           string
			IsRegistered bool
			IsOver       bool
		}{
			ID:           eventID,
			IsRegistered: registered,
			IsOver:       event.IsOver(0),
		}))
	return nil
}

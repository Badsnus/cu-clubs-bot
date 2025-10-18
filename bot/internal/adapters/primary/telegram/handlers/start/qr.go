package start

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"
	"gorm.io/gorm"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/banner"
)

func (h Handler) userQR(c tele.Context, qrCodeID string) error {
	_ = c.Delete()
	h.logger.Infof("(user: %d) scan user QR code", c.Sender().ID)

	user, err := h.userService.GetByQRCodeID(context.Background(), qrCodeID)
	if err != nil {
		h.logger.Infof("(user: %d) qr expired: %v", c.Sender().ID, err)
		return c.Send(
			banner.ClubOwner.Caption(h.layout.Text(c, "qr_expired")),
			h.layout.Markup(c, "core:hide"),
		)
	}

	if user.ID == c.Sender().ID {
		h.logger.Infof("(user: %d) user try to scan own qr", c.Sender().ID)
		return c.Send(
			banner.ClubOwner.Caption(h.layout.Text(c, "self_qr_error")),
			h.layout.Markup(c, "core:hide"),
		)
	}

	userClubs, err := h.clubService.GetByOwnerID(context.Background(), c.Sender().ID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while getting user's clubs from db: %v", c.Sender().ID, err)
		return c.Send(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "core:hide"),
		)
	}

	if len(userClubs) == 0 {
		h.logger.Infof("(user: %d) user has no clubs", c.Sender().ID)
		return c.Send(
			banner.ClubOwner.Caption(h.layout.Text(c, "no_clubs")),
			h.layout.Markup(c, "core:hide"),
		)
	}

	var rows []tele.Row
	markup := c.Bot().NewMarkup()
	for _, club := range userClubs {
		callbackID, errSet := h.callbacksStorage.Set(fmt.Sprintf("%s %s", club.ID, qrCodeID), time.Minute*5)
		if errSet != nil {
			h.logger.Errorf("(user: %d) error while setting callback: %v", c.Sender().ID, errSet)
			continue
		}
		rows = append(rows, markup.Row(*h.layout.Button(c, "clubOwner:activateQR:club", struct {
			CallbackID string
			Name       string
		}{
			CallbackID: callbackID,
			Name:       club.Name,
		})))
	}

	rows = append(
		rows,
		markup.Row(*h.layout.Button(c, "core:cancel")),
	)
	markup.Inline(rows...)

	return c.Send(
		banner.ClubOwner.Caption(h.layout.Text(c, "qr_clubs_list", struct {
			FIO      string
			Username string
		}{
			FIO:      user.FIO.String(),
			Username: user.Username,
		})),
		markup,
	)
}

func (h Handler) backToClubsList(c tele.Context) error {
	qrCodeID := c.Callback().Data

	h.logger.Infof("(user: %d) back to activate qr clubs list", c.Sender().ID)

	user, err := h.userService.GetByQRCodeID(context.Background(), qrCodeID)
	if err != nil {
		h.logger.Infof("(user: %d) qr expired: %v", c.Sender().ID, err)
		return c.Edit(
			banner.ClubOwner.Caption(h.layout.Text(c, "qr_expired")),
			h.layout.Markup(c, "core:hide"),
		)
	}

	if user.ID == c.Sender().ID {
		return c.Edit(
			banner.ClubOwner.Caption(h.layout.Text(c, "self_qr_error")),
			h.layout.Markup(c, "core:hide"),
		)
	}

	userClubs, err := h.clubService.GetByOwnerID(context.Background(), c.Sender().ID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while getting user's clubs from db: %v", c.Sender().ID, err)
		return c.Edit(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "core:hide"),
		)
	}

	if len(userClubs) == 0 {
		return c.Edit(
			banner.ClubOwner.Caption(h.layout.Text(c, "no_clubs")),
			h.layout.Markup(c, "core:hide"),
		)
	}

	var rows []tele.Row
	markup := c.Bot().NewMarkup()
	for _, club := range userClubs {
		callbackID, errSet := h.callbacksStorage.Set(fmt.Sprintf("%s %s", club.ID, qrCodeID), time.Minute*5)
		if errSet != nil {
			h.logger.Errorf("(user: %d) error while setting callback: %v", c.Sender().ID, errSet)
			continue
		}
		rows = append(rows, markup.Row(*h.layout.Button(c, "clubOwner:activateQR:club", struct {
			CallbackID string
			Name       string
		}{
			CallbackID: callbackID,
			Name:       club.Name,
		})))
	}

	rows = append(
		rows,
		markup.Row(*h.layout.Button(c, "core:cancel")),
	)
	markup.Inline(rows...)

	return c.Edit(
		banner.ClubOwner.Caption(h.layout.Text(c, "qr_clubs_list", struct {
			FIO      string
			Username string
		}{
			FIO:      user.FIO.String(),
			Username: user.Username,
		})),
		markup,
	)
}

func (h Handler) qrEventsList(c tele.Context) error {
	callbackData, err := h.callbacksStorage.Get(c.Callback().Data)
	if err != nil {
		h.logger.Errorf("(user: %d) error while getting callback from redis: %v", c.Sender().ID, err)
		return c.Edit(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "core:hide"),
		)
	}
	h.callbacksStorage.Delete(c.Callback().Data)
	clubID, qrCodeID := strings.Split(callbackData, " ")[0], strings.Split(callbackData, " ")[1]

	user, err := h.userService.GetByQRCodeID(context.Background(), qrCodeID)
	if err != nil {
		h.logger.Infof("(user: %d) qr expired: %v", c.Sender().ID, err)
		return c.Edit(
			banner.ClubOwner.Caption(h.layout.Text(c, "qr_expired")),
			h.layout.Markup(c, "core:hide"),
		)
	}

	h.logger.Infof("(user: %d) qr events list (club_id=%s, qr_id=%s, qr_owner_id=%d)", c.Sender().ID, clubID, qrCodeID, user.ID)

	events, err := h.eventService.GetFutureByClubID(
		context.Background(),
		-1,
		0,
		"start_time ASC",
		clubID,
		time.Hour*24,
	)
	if err != nil {
		h.logger.Errorf("(user: %d) error while get events: %v", c.Sender().ID, err)
		return c.Edit(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "core:hide"),
		)
	}

	var rows []tele.Row
	markup := c.Bot().NewMarkup()
	for _, event := range events {
		callbackID, errSet := h.callbacksStorage.Set(fmt.Sprintf("%s %s", event.ID, qrCodeID), time.Minute*5)
		if errSet != nil {
			h.logger.Errorf("(user: %d) error while setting callback: %v", c.Sender().ID, errSet)
			continue
		}
		rows = append(rows, markup.Row(*h.layout.Button(c, "clubOwner:activateQR:event", struct {
			CallbackID string
			Name       string
		}{
			CallbackID: callbackID,
			Name:       event.Name,
		})))
	}

	rows = append(
		rows,
		markup.Row(*h.layout.Button(c, "clubOwner:activateQR:clubs:back", struct {
			QrID string
		}{
			QrID: qrCodeID,
		})),
	)
	markup.Inline(rows...)

	h.logger.Infof("(user: %d) edit qr events list", c.Sender().ID)
	return c.Edit(
		banner.ClubOwner.Caption(h.layout.Text(c, "qr_events_list", struct {
			FIO      string
			Username string
		}{
			FIO:      user.FIO.String(),
			Username: user.Username,
		})),
		markup,
	)
}

func (h Handler) activateUserQR(c tele.Context) error {
	callbackData, err := h.callbacksStorage.Get(c.Callback().Data)
	if err != nil {
		h.logger.Errorf("(user: %d) error while getting callback from redis: %v", c.Sender().ID, err)
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "core:hide"),
		)
	}
	h.callbacksStorage.Delete(c.Callback().Data)
	eventID, qrCodeID := strings.Split(callbackData, " ")[0], strings.Split(callbackData, " ")[1]

	h.logger.Infof("(user: %d) activate user qr (event_id=%s, qr_id=%s)", c.Sender().ID, eventID, qrCodeID)
	user, err := h.userService.GetByQRCodeID(context.Background(), qrCodeID)
	if err != nil {
		h.logger.Infof("(user: %d) qr expired: %v", c.Sender().ID, err)
		return c.Edit(
			banner.ClubOwner.Caption(h.layout.Text(c, "qr_expired")),
			h.layout.Markup(c, "core:hide"),
		)
	}

	event, err := h.eventService.Get(context.Background(), eventID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while getting event from db: %v", c.Sender().ID, err)
		return c.Edit(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "core:hide"),
		)
	}

	if event.IsOver(time.Hour * 24) {
		h.logger.Infof("(user: %d) event already started (event_id=%s)", c.Sender().ID, eventID)
		return c.Edit(
			banner.ClubOwner.Caption(h.layout.Text(c, "event_started")),
			h.layout.Markup(c, "core:hide"),
		)
	}

	eventParticipant, err := h.eventParticipantService.Get(context.Background(), eventID, user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.Infof("(user: %d) participant not found (event_id=%s)", c.Sender().ID, eventID)
			eventParticipant, err = h.eventParticipantService.Register(context.Background(), eventID, user.ID)
			if err != nil {
				h.logger.Errorf("(user: %d) error while registering participant: %v", c.Sender().ID, err)
				return c.Edit(
					banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
					h.layout.Markup(c, "core:hide"),
				)
			}
			h.logger.Infof("(user: %d) participant registered (event_id=%s, user_id=%d)", c.Sender().ID, eventID, user.ID)
		} else {
			h.logger.Infof("(user: %d) error while getting event participant from db: %v", c.Sender().ID, err)
			return c.Edit(
				banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
				h.layout.Markup(c, "core:hide"),
			)
		}
	}

	err = h.qrService.RevokeUserQR(context.Background(), user.ID)
	if err != nil {
		h.logger.Errorf("(user: %d) error while revoking user QR code: %v", c.Sender().ID, err)
		return c.Edit(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "core:hide"),
		)
	}

	eventParticipant.IsUserQr = true
	_, err = h.eventParticipantService.Update(context.Background(), eventParticipant)
	if err != nil {
		h.logger.Errorf("(user: %d) error while updating event participant: %v", c.Sender().ID, err)
		return c.Edit(
			banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "core:hide"),
		)
	}

	h.logger.Infof("(user: %d) user qr activated (event_id=%s, user_id=%d)", c.Sender().ID, eventID, user.ID)

	return c.Edit(
		banner.ClubOwner.Caption(h.layout.Text(c, "qr_activated", struct {
			FIO      string
			Username string
		}{
			FIO:      user.FIO.String(),
			Username: user.Username,
		})),
		h.layout.Markup(c, "core:hide"),
	)
}

func (h Handler) SetupUserQR(group *tele.Group) {
	group.Handle(h.layout.Callback("clubOwner:activateQR:clubs:back"), h.backToClubsList)
	group.Handle(h.layout.Callback("clubOwner:activateQR:club"), h.qrEventsList)
	group.Handle(h.layout.Callback("clubOwner:activateQR:event"), h.activateUserQR)
}

func (h Handler) eventQR(c tele.Context, qrCodeID string) error {
	_ = c.Delete()
	h.logger.Infof("(user: %d) scan event QR code", c.Sender().ID)

	event, err := h.eventService.GetByQRCodeID(context.Background(), qrCodeID)
	if err != nil {
		h.logger.Infof("(user: %d) event qr expired: %v", c.Sender().ID, err)
		return c.Send(
			banner.Events.Caption(h.layout.Text(c, "qr_expired")),
			h.layout.Markup(c, "core:hide"),
		)
	}

	if event.IsOver(time.Hour * 24) {
		h.logger.Infof("(user: %d) event already started (event_id=%s)", c.Sender().ID, event.ID)
		return c.Edit(
			banner.Events.Caption(h.layout.Text(c, "event_started")),
			h.layout.Markup(c, "core:hide"),
		)
	}

	eventParticipant, err := h.eventParticipantService.Get(context.Background(), event.ID, c.Sender().ID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.Infof("(user: %d) error while getting event participant from db: %v", c.Sender().ID, err)
			return c.Edit(
				banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
				h.layout.Markup(c, "core:hide"),
			)
		}
		h.logger.Infof("(user: %d) participant not found (event_id=%s)", c.Sender().ID, event.ID)
		eventParticipant, err = h.eventParticipantService.Register(context.Background(), event.ID, c.Sender().ID)
		if err != nil {
			h.logger.Errorf("(user: %d) error while registering participant: %v", c.Sender().ID, err)
			return c.Edit(
				banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
				h.layout.Markup(c, "core:hide"),
			)
		}
		h.logger.Infof("(user: %d) participant registered (event_id=%s, user_id=%d)", c.Sender().ID, event.ID, c.Sender().ID)
	}

	eventParticipant.IsEventQr = true
	_, err = h.eventParticipantService.Update(context.Background(), eventParticipant)
	if err != nil {
		h.logger.Errorf("(user: %d) error while updating event participant: %v", c.Sender().ID, err)
		return c.Edit(
			banner.Events.Caption(h.layout.Text(c, "technical_issues", err.Error())),
			h.layout.Markup(c, "core:hide"),
		)
	}
	h.logger.Infof("(user: %d) event qr activated (event_id=%s, user_id=%d)", c.Sender().ID, event.ID, c.Sender().ID)

	return c.Send(
		banner.Events.Caption(h.layout.Text(c, "event_qr_activated", struct {
			Name string
		}{
			Name: event.Name,
		})),
		h.layout.Markup(c, "core:hide"),
	)
}

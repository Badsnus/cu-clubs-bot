package menu

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/banner"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type clubService interface {
	GetByOwnerID(ctx context.Context, id int64) ([]entity.Club, error)
}

type Handler struct {
	clubService clubService

	layout *layout.Layout
	logger *types.Logger
}

func New(b *bot.Bot) *Handler {
	clubStorage := postgres.NewClubStorage(b.DB)

	return &Handler{
		clubService: service.NewClubService(clubStorage),

		logger: b.Logger,
		layout: b.Layout,
	}
}

func (h Handler) SendMenu(c tele.Context) error {
	isAdmin := utils.IsAdmin(c.Sender().ID)

	menuMarkup := h.layout.Markup(c, "mainMenu:menu")

	//userClubs, err := h.clubService.GetByOwnerID(context.Background(), c.Sender().ID)
	//if err != nil {
	//	h.logger.Errorf("(user: %d) error while getting user's clubs from db: %v", c.Sender().ID, err)
	//	return c.Send(
	//		h.layout.Text(c, "technical_issues", err.Error()),
	//		h.layout.Markup(c, "core:hide"),
	//	)
	//}

	//if len(userClubs) == 1 {
	//	menuMarkup.InlineKeyboard = append(menuMarkup.InlineKeyboard, []tele.InlineButton{*h.layout.Button(c, "clubOwner:myClubs:club", struct {
	//		ID   string
	//		Name string
	//	}{
	//		ID:   userClubs[0].ID,
	//		Name: userClubs[0].Name,
	//	}).Inline()})
	//}
	//if len(userClubs) > 1 {
	//	menuMarkup.InlineKeyboard = append(menuMarkup.InlineKeyboard, []tele.InlineButton{*h.layout.Button(c, "clubOwner:my_clubs").Inline()})
	//}
	menuMarkup.InlineKeyboard = append(menuMarkup.InlineKeyboard, []tele.InlineButton{*h.layout.Button(c, "clubOwner:my_clubs").Inline()})

	if isAdmin {
		menuMarkup.InlineKeyboard = append(menuMarkup.InlineKeyboard, []tele.InlineButton{*h.layout.Button(c, "mainMenu:admin_menu").Inline()})
	}

	h.logger.Infof("(user: %d) send main menu (isAdmin=%t)", c.Sender().ID, isAdmin)
	return c.Send(
		banner.Menu.Caption(h.layout.Text(c, "main_menu_text", c.Sender().Username)),
		menuMarkup,
	)
}

func (h Handler) EditMenu(c tele.Context) error {
	isAdmin := utils.IsAdmin(c.Sender().ID)

	menuMarkup := h.layout.Markup(c, "mainMenu:menu")

	//userClubs, err := h.clubService.GetByOwnerID(context.Background(), c.Sender().ID)
	//if err != nil {
	//	h.logger.Errorf("(user: %d) error while getting user's clubs from db: %v", c.Sender().ID, err)
	//	return c.Send(
	//		h.layout.Text(c, "technical_issues", err.Error()),
	//		h.layout.Markup(c, "core:hide"),
	//	)
	//}

	//if len(userClubs) == 1 {
	//	menuMarkup.InlineKeyboard = append(menuMarkup.InlineKeyboard, []tele.InlineButton{*h.layout.Button(c, "clubOwner:myClubs:club", struct {
	//		ID   string
	//		Name string
	//	}{
	//		ID:   userClubs[0].ID,
	//		Name: userClubs[0].Name,
	//	}).Inline()})
	//}
	//if len(userClubs) > 1 {
	//	menuMarkup.InlineKeyboard = append(menuMarkup.InlineKeyboard, []tele.InlineButton{*h.layout.Button(c, "clubOwner:my_clubs").Inline()})
	//}
	menuMarkup.InlineKeyboard = append(menuMarkup.InlineKeyboard, []tele.InlineButton{*h.layout.Button(c, "clubOwner:my_clubs").Inline()})

	if isAdmin {
		menuMarkup.InlineKeyboard = append(menuMarkup.InlineKeyboard, []tele.InlineButton{*h.layout.Button(c, "mainMenu:admin_menu").Inline()})
	}

	h.logger.Infof("(user: %d) edit main menu (isAdmin=%t)", c.Sender().ID, isAdmin)
	return c.Edit(
		banner.Menu.Caption(h.layout.Text(c, "main_menu_text", c.Sender().Username)),
		menuMarkup,
	)
}

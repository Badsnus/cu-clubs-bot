package setup

import (
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/primary/telegram/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/primary/telegram/handlers/admin"
	clubowner "github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/primary/telegram/handlers/clubOwner"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/primary/telegram/handlers/menu"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/primary/telegram/handlers/middlewares"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/primary/telegram/handlers/start"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/primary/telegram/handlers/user"
)

func Setup(
	bot *bot.Bot,
	middle *middlewares.Handler,
	startHandler *start.Handler,
	userHandler *user.Handler,
	clubOwnerHandler *clubowner.Handler,
	menuHandler *menu.Handler,
	adminHandler *admin.Handler,
	debug bool,
	adminIDs []int64,
) {
	// Pre-setup and global middlewares
	bot.Use(middle.PrivateChatOnly)
	if debug {
		bot.Use(middleware.Logger())
	}
	bot.Use(bot.Layout.Middleware("ru"))
	bot.Use(middleware.AutoRespond())
	bot.Handle(tele.OnText, bot.Input.MessageHandler())
	bot.Handle(tele.OnMedia, bot.Input.MessageHandler())
	bot.Handle(tele.OnCallback, bot.Input.CallbackHandler())
	bot.Handle(tele.OnVideoNote, bot.Input.MessageHandler())
	bot.Use(middle.ResetInputOnBack)
	bot.Handle(bot.Layout.Callback("core:hide"), userHandler.Hide)
	bot.Handle(bot.Layout.Callback("core:cancel"), userHandler.Hide)
	bot.Handle(bot.Layout.Callback("core:back"), userHandler.Hide)

	// Setup handlers
	// Start
	bot.Handle("/start", startHandler.Start)

	// Auth
	userHandler.AuthSetup(bot.Group())
	bot.Use(middle.Authorized)

	// Qr
	startHandler.SetupUserQR(bot.Group())

	// User:
	bot.Handle(bot.Layout.Callback("mainMenu:back"), menuHandler.EditMenu)
	userHandler.UserSetup(bot.Group())
	startHandler.SetupURLEvent(bot.Group())

	// ClubOwner:
	clubOwnerHandler.ClubOwnerSetup(bot.Group(), middle)

	// Admin:
	bot.Use(middleware.Whitelist(adminIDs...))
	adminHandler.AdminSetup(bot.Group())
}

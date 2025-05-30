package setup

import (
	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/admin"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/clubOwner"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/menu"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/middlewares"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/start"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/user"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/smtp"
	"github.com/spf13/viper"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

func Setup(b *bot.Bot) {
	// Start notification scheduler
	notifyLogger, err := logger.Named("notify")
	if err != nil {
		b.Logger.Fatalf("Failed to create notify logger: %v", err)
	}
	notifyService := service.NewNotifyService(
		b.Bot,
		b.Layout,
		notifyLogger,
		service.NewClubOwnerService(postgres.NewClubOwnerStorage(b.DB), postgres.NewUserStorage(b.DB)),
		postgres.NewEventStorage(b.DB),
		postgres.NewNotificationStorage(b.DB),
		nil,
	)
	eventParticipantService := service.NewEventParticipantService(
		b.Bot,
		b.Layout,
		b.Logger,
		postgres.NewEventParticipantStorage(b.DB),
		postgres.NewEventStorage(b.DB),
		postgres.NewUserStorage(b.DB),
		postgres.NewClubStorage(b.DB),
		smtp.NewClient(b.SMTPDialer, viper.GetString("service.smtp.domain"), viper.GetString("service.smtp.email")),
		viper.GetStringSlice("settings.pass-emails"),
		viper.GetInt64("bot.pass.channel-id"),
	)
	notifyService.StartNotifyScheduler()
	eventParticipantService.StartPassScheduler()

	// Pre-setup and global middlewares
	middle := middlewares.New(b)
	startHandler := start.New(b)
	userHandler := user.New(b)
	clubOwnerHandler := clubowner.NewHandler(b)
	menuHandler := menu.New(b)
	adminHandler := admin.New(b)

	b.Use(middle.PrivateChatOnly)
	if viper.GetBool("settings.logging.debug") {
		b.Use(middleware.Logger())
	}
	b.Use(middle.LoadBanners)
	b.Use(b.Layout.Middleware("ru"))
	b.Use(middleware.AutoRespond())
	b.Handle(tele.OnText, b.Input.MessageHandler())
	b.Handle(tele.OnMedia, b.Input.MessageHandler())
	b.Handle(tele.OnCallback, b.Input.CallbackHandler())
	b.Use(middle.ResetInputOnBack)
	b.Handle(b.Layout.Callback("core:hide"), userHandler.Hide)
	b.Handle(b.Layout.Callback("core:cancel"), userHandler.Hide)
	b.Handle(b.Layout.Callback("core:back"), userHandler.Hide)

	// Setup handlers
	//Start
	b.Handle("/start", startHandler.Start)

	//Auth
	userHandler.AuthSetup(b.Group())
	b.Use(middle.Authorized)

	//Qr
	startHandler.SetupUserQR(b.Group())

	//User:
	b.Handle(b.Layout.Callback("mainMenu:back"), menuHandler.EditMenu)
	userHandler.UserSetup(b.Group())
	startHandler.SetupURLEvent(b.Group())

	// ClubOwner:
	clubOwnerHandler.ClubOwnerSetup(b.Group(), middle)

	//Admin:
	admins := viper.GetIntSlice("bot.admin-ids")
	adminsInt64 := make([]int64, len(admins))
	for i, v := range admins {
		adminsInt64[i] = int64(v)
	}
	b.Use(middleware.Whitelist(adminsInt64...))
	adminHandler.AdminSetup(b.Group())

	//adminHandler := handlers.NewAdminHandler(b)
	//b.Handle("/admin", adminHandler.AdminMenu)
	//b.Handle(b.Layout.Callback("backToAdminMenu"), adminHandler.BackToAdminMenu)
	//b.Handle(b.Layout.Callback("manageUsers"), adminHandler.UsersList)
	//b.Handle(b.Layout.Callback("backToUsersList"), adminHandler.UsersList)
	//b.Handle(b.Layout.Callback("user"), adminHandler.ManageUser)
	//b.Handle(b.Layout.Callback("banUser"), adminHandler.BanUser)
	//b.Handle(b.Layout.Callback("mailing"), adminHandler.Mailing)
	//b.Handle(b.Layout.Callback("confirmMailing"), adminHandler.SendMailing)
	//b.Handle(b.Layout.Callback("cancelMailing"), adminHandler.BackToAdminMenu)
}

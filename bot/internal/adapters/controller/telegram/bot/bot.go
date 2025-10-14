package bot

import (
	"github.com/nlypage/intele"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	"gopkg.in/gomail.v2"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"gorm.io/gorm"
)

type Bot struct {
	*tele.Bot
	Layout     *layout.Layout
	DB         *gorm.DB
	Redis      *redis.Client
	SMTPDialer *gomail.Dialer
	Logger     *types.Logger
	Input      *intele.InputManager
}

func New(db *gorm.DB, redisClient *redis.Client, smtpDialer *gomail.Dialer) (*Bot, error) {
	lt, err := layout.New("telegram.yml")
	if err != nil {
		return nil, err
	}

	settings := lt.Settings()
	botLogger, err := logger.Named("bot")
	if err != nil {
		return nil, err
	}
	settings.OnError = func(err error, ctx tele.Context) {
		if ctx.Callback() == nil {
			botLogger.Errorf("(user: %d) | Error: %v", ctx.Sender().ID, err)
		} else {
			botLogger.Errorf("(user: %d) | unique: %s | Error: %v", ctx.Sender().ID, ctx.Callback().Unique, err)
		}
	}

	b, err := tele.NewBot(settings)
	if err != nil {
		return nil, err
	}

	if cmds := lt.Commands(); cmds != nil {
		if err = b.SetCommands(cmds); err != nil {
			return nil, err
		}
	}

	bot := &Bot{
		Bot:    b,
		Layout: lt,
		DB:     db,
		Input: intele.NewInputManager(intele.InputOptions{
			Storage: redisClient.States,
		}),
		SMTPDialer: smtpDialer,
		Logger:     botLogger,
		Redis:      redisClient,
	}

	return bot, nil
}

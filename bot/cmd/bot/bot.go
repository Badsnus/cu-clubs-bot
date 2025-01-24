package bot

import (
	"github.com/nlypage/intele"
	"sync"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"gopkg.in/gomail.v2"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/config"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
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

func New(config *config.Config) (*Bot, error) {
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
		DB:     config.Database,
		Input: intele.NewInputManager(intele.InputOptions{
			Storage: config.Redis.States,
		}),
		SMTPDialer: config.SMTPDialer,
		Logger:     botLogger,
		Redis:      config.Redis,
	}

	return bot, nil
}

func (b *Bot) Start() {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		logger.Log.Info("Bot starting")

		if viper.GetBool("settings.logging.log-to-channel") {
			notifyLogger, err := logger.Named("notify")
			if err != nil {
				logger.Log.Errorf("Failed to create notify logger: %v", err)
			} else {
				notifyService := service.NewNotifyService(b.Bot, b.Layout, notifyLogger)
				logHook, err := notifyService.LogHook(
					viper.GetInt64("settings.logging.channel-id"),
					viper.GetString("settings.logging.locale"),
					zapcore.Level(viper.GetInt("settings.logging.channel-log-level")),
				)
				if err != nil {
					logger.Log.Errorf("Failed to create notify log hook: %v", err)
				} else {
					logger.SetLogHook(logHook)
				}
			}
		}
		b.Bot.Start()
	}()

	wg.Wait()
}

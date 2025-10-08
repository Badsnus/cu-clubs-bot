package bot

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nlypage/intele"

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
	logger.Log.Info("Bot starting")

	// Setup logging hook if enabled
	if viper.GetBool("settings.logging.log-to-channel") {
		notifyLogger, err := logger.Named("notify")
		if err != nil {
			logger.Log.Errorf("Failed to create notify logger: %v", err)
		} else {
			notifyService := service.NewNotifyService(b.Bot, b.Layout, notifyLogger, nil, nil, nil, nil)
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

	// Send version notification if enabled
	if viper.GetBool("settings.version.notify-on-startup") {
		versionLogger, err := logger.Named("version")
		if err != nil {
			logger.Log.Errorf("Failed to create version logger: %v", err)
		} else {
			versionService := service.NewVersionService(b.Bot, b.Layout, versionLogger)
			err := versionService.SendStartupNotification(
				viper.GetInt64("settings.version.channel-id"),
			)
			if err != nil {
				logger.Log.Errorf("Failed to send version notification: %v", err)
			}
		}
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start bot in goroutine
	go func() {
		b.Bot.Start()
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	logger.Log.Infof("Received signal %v, starting graceful shutdown...", sig)
	b.gracefulShutdown()
}

func (b *Bot) gracefulShutdown() {
	logger.Log.Info("Starting graceful shutdown...")

	// Stop the bot
	logger.Log.Info("Stopping bot...")
	b.Bot.Stop()
	logger.Log.Info("Bot stopped")

	// Close database connection
	if b.DB != nil {
		logger.Log.Info("Closing database connection...")
		sqlDB, err := b.DB.DB()
		if err != nil {
			logger.Log.Errorf("Failed to get underlying sql.DB: %v", err)
		} else {
			if err := sqlDB.Close(); err != nil {
				logger.Log.Errorf("Error closing database connection: %v", err)
			} else {
				logger.Log.Info("Database connection closed")
			}
		}
	}

	// Final log and cleanup
	logger.Log.Info("Graceful shutdown completed")
	time.Sleep(100 * time.Millisecond) // Allow final log to be written

	// Close logger resources last
	if err := logger.Cleanup(); err != nil {
		_ = err
	}
}

package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/bot"
	setupBot "github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/setup"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/banner"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger"
)

// App represents the main application structure.
type App struct {
	serviceProvider *serviceProvider
}

// NewApp initializes the application and its dependencies.
func NewApp(ctx context.Context) (*App, error) {
	a := &App{}

	err := a.initDeps(ctx)
	if err != nil {
		return nil, fmt.Errorf("new app: %w", err)
	}

	return a, nil
}

// Run starts the application.
func (a *App) Run() {
	defer a.gracefulShutdown()

	logger.Log.Info("Bot starting")

	// Setup bot handlers
	setupBot.Setup(a.serviceProvider.Bot())

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start bot in goroutine
	go func() {
		a.serviceProvider.Bot().Start()
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	logger.Log.Infof("Received signal %v, starting graceful shutdown...", sig)
}

// gracefulShutdown handles cleanup of all resources
func (a *App) gracefulShutdown() {
	logger.Log.Info("Starting graceful shutdown...")

	// Stop schedulers
	if a.serviceProvider != nil {
		// Stop pass scheduler
		if a.serviceProvider.passService != nil {
			logger.Log.Info("Stopping pass scheduler...")
			a.serviceProvider.passService.StopScheduler()
			logger.Log.Info("Pass scheduler stopped")
		}

		// Stop the bot
		if a.serviceProvider.bot != nil {
			logger.Log.Info("Stopping bot...")
			a.serviceProvider.bot.Stop()
			logger.Log.Info("Bot stopped")
		}

		if a.serviceProvider.db != nil {
			logger.Log.Info("Closing database connection...")
			sqlDB, err := a.serviceProvider.db.DB()
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
	}

	// Final log and cleanup
	logger.Log.Info("Graceful shutdown completed")

	// Close logger resources last
	if err := logger.Cleanup(); err != nil {
		// Can't log this error as logger is closing
		_ = err
	}
}

// initDeps initializes application dependencies
func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initConfig,
		a.initServiceProvider,
		a.initLogger,
		a.initBot,
		a.initBanner,
		a.initPassScheduler,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return fmt.Errorf("init deps: %w", err)
		}
	}

	return nil
}

func (a *App) initConfig(_ context.Context) error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Search config in multiple locations
	viper.AddConfigPath("/opt/config") // docker mounted config
	viper.AddConfigPath(".")           // current directory
	viper.AddConfigPath("../")         // parent directory
	viper.AddConfigPath("../../")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	return nil
}

func (a *App) initServiceProvider(_ context.Context) error {
	a.serviceProvider = newServiceProvider()
	return nil
}

// initBot initializes the bot and sets up hooks and notifications
func (a *App) initBot(_ context.Context) error {
	// Create bot instance
	b, err := bot.New(a.serviceProvider.DB(), a.serviceProvider.RedisClient(), a.serviceProvider.SMTPDialer())
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	// Set bot in service provider
	a.serviceProvider.setBot(b)

	// Setup logging hook if enabled
	if a.serviceProvider.cfg.Logger.LogToChannel() {
		notifyService := a.serviceProvider.NotifyService()
		logHook, err := notifyService.LogHook(
			a.serviceProvider.cfg.Logger.ChannelID(),
			a.serviceProvider.cfg.Logger.ChannelLocale(),
			zapcore.Level(a.serviceProvider.cfg.Logger.ChannelLogLevel()),
		)
		if err != nil {
			logger.Log.Errorf("Failed to create notify log hook: %v", err)
		} else {
			logger.SetLogHook(logHook)
		}
	}

	// Send version notification if enabled
	if a.serviceProvider.cfg.App.VersionNotifyOnStartup() {
		versionService := a.serviceProvider.VersionService()
		err := versionService.SendStartupNotification(
			a.serviceProvider.cfg.App.VersionChannelID(),
		)
		if err != nil {
			logger.Log.Errorf("Failed to send version notification: %v", err)
		}
	}

	return nil
}

func (a *App) initLogger(_ context.Context) error {
	return logger.Init(logger.Config{
		Debug:        a.serviceProvider.cfg.Logger.Debug(),
		TimeLocation: a.serviceProvider.cfg.Logger.TimeLocation(),
		LogToFile:    a.serviceProvider.cfg.Logger.LogToFile(),
		LogsDir:      a.serviceProvider.cfg.Logger.LogsDir(),
	})
}

// initBanner initializes banner files from Telegram (required for app startup)
func (a *App) initBanner(_ context.Context) error {
	err := banner.Load(a.serviceProvider.Bot().Bot, a.serviceProvider.cfg.Banner)
	if err != nil {
		return fmt.Errorf("failed to initialize banners: %w", err)
	}
	return nil
}

// initPassScheduler initializes and starts the pass scheduler
func (a *App) initPassScheduler(_ context.Context) error {
	err := a.serviceProvider.PassService().StartScheduler()
	if err != nil {
		return fmt.Errorf("failed to start pass scheduler: %w", err)
	}

	return nil
}

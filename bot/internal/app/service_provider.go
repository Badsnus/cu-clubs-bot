package app

import (
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/gomail.v2"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/config"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/admin"
	clubowner "github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/clubOwner"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/menu"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/middlewares"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/start"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/user"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger"
	qr "github.com/Badsnus/cu-clubs-bot/bot/pkg/qrcode"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/smtp"
)

type serviceProvider struct {
	// Configuration
	cfg *config.Config

	// Infrastructure
	db          *gorm.DB
	redisClient *redis.Client
	smtpDialer  *gomail.Dialer
	smtpClient  *smtp.Client

	// Bot dependencies
	bot *bot.Bot

	// Storage layer
	userStorage             *postgres.UserStorage
	clubStorage             *postgres.ClubStorage
	eventStorage            *postgres.EventStorage
	eventParticipantStorage *postgres.EventParticipantStorage
	passStorage             *postgres.PassStorage
	clubOwnerStorage        *postgres.ClubOwnerStorage
	notificationStorage     *postgres.NotificationStorage

	// Service layer
	userService             *service.UserService
	clubService             *service.ClubService
	eventService            *service.EventService
	eventParticipantService *service.EventParticipantService
	passService             *service.PassService
	clubOwnerService        *service.ClubOwnerService
	notifyService           *service.NotifyService
	qrService               *service.QrService
	versionService          *service.VersionService

	// Handlers
	adminHandler       *admin.Handler
	userHandler        *user.Handler
	startHandler       *start.Handler
	menuHandler        *menu.Handler
	middlewaresHandler *middlewares.Handler
	clubOwnerHandler   *clubowner.Handler
}

func newServiceProvider() *serviceProvider {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(fmt.Errorf("failed to create config: %w", err))
	}

	// Set BOT_TOKEN environment variable from config
	if err := os.Setenv("BOT_TOKEN", cfg.Bot.Token()); err != nil {
		panic(fmt.Errorf("failed to set BOT_TOKEN environment variable: %w", err))
	}

	return &serviceProvider{
		cfg: cfg,
	}
}

// Infrastructure dependencies

func (s *serviceProvider) DB() *gorm.DB {
	if s.db == nil {
		var gormConfig *gorm.Config
		if s.cfg.Logger.Debug() {
			newLogger := gormLogger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags),
				gormLogger.Config{
					SlowThreshold: time.Second,
					LogLevel:      gormLogger.Info,
					Colorful:      true,
				},
			)
			gormConfig = &gorm.Config{
				TranslateError: true,
				Logger:         newLogger,
			}
		} else {
			gormConfig = &gorm.Config{
				TranslateError: true,
			}
		}

		// Set location
		time.Local = s.cfg.Logger.TimeLocation()

		dsn := s.cfg.PG.DSN()
		logger.Log.Info(dsn)

		database, err := gorm.Open(postgresDriver.Open(dsn), gormConfig)
		if err != nil {
			panic(fmt.Errorf("failed to connect to the database: %w", err))
		}
		logger.Log.Info("Successfully connected to the database")

		// Configure connection pool
		sqlDB, err := database.DB()
		if err != nil {
			panic(fmt.Errorf("failed to get underlying sql.DB: %w", err))
		}

		// Set maximum number of open connections to the database
		sqlDB.SetMaxOpenConns(25)
		// Set maximum number of idle connections in the connection pool
		sqlDB.SetMaxIdleConns(10)
		// Set maximum amount of time a connection may be reused
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
		// Set maximum amount of time a connection may be idle
		sqlDB.SetConnMaxIdleTime(1 * time.Minute)

		logger.Log.Info("Database connection pool configured")

		errMigrate := database.AutoMigrate(postgres.Migrations...)
		if errMigrate != nil {
			panic(fmt.Errorf("failed to migrate database: %w", errMigrate))
		}

		s.db = database
	}

	return s.db
}

func (s *serviceProvider) RedisClient() *redis.Client {
	if s.redisClient == nil {
		r, err := redis.New(redis.Options{
			Host:     s.cfg.RedisConf.Host(),
			Port:     s.cfg.RedisConf.Port(),
			Password: s.cfg.RedisConf.Password(),
		})
		if err != nil {
			panic(fmt.Errorf("failed to connect to redis: %w", err))
		}

		s.redisClient = r
	}

	return s.redisClient
}

func (s *serviceProvider) Redis() *redis.Client {
	return s.redisClient
}

func (s *serviceProvider) SMTPDialer() *gomail.Dialer {
	if s.smtpDialer == nil {
		s.smtpDialer = gomail.NewDialer(
			s.cfg.SMTP.Host(),
			s.cfg.SMTP.Port(),
			s.cfg.SMTP.Login(),
			s.cfg.SMTP.Password(),
		)
	}

	return s.smtpDialer
}

func (s *serviceProvider) SMTPClient() *smtp.Client {
	if s.smtpClient == nil {
		s.smtpClient = smtp.NewClient(s.SMTPDialer(), s.cfg.SMTP.Domain(), s.cfg.SMTP.Email())
	}

	return s.smtpClient
}

// Storage layer

func (s *serviceProvider) UserStorage() *postgres.UserStorage {
	if s.userStorage == nil {
		s.userStorage = postgres.NewUserStorage(s.DB())
	}

	return s.userStorage
}

func (s *serviceProvider) ClubStorage() *postgres.ClubStorage {
	if s.clubStorage == nil {
		s.clubStorage = postgres.NewClubStorage(s.DB())
	}

	return s.clubStorage
}

func (s *serviceProvider) EventStorage() *postgres.EventStorage {
	if s.eventStorage == nil {
		s.eventStorage = postgres.NewEventStorage(s.DB())
	}

	return s.eventStorage
}

func (s *serviceProvider) EventParticipantStorage() *postgres.EventParticipantStorage {
	if s.eventParticipantStorage == nil {
		s.eventParticipantStorage = postgres.NewEventParticipantStorage(s.DB())
	}

	return s.eventParticipantStorage
}

func (s *serviceProvider) PassStorage() *postgres.PassStorage {
	if s.passStorage == nil {
		s.passStorage = postgres.NewPassStorage(s.DB())
	}

	return s.passStorage
}

func (s *serviceProvider) ClubOwnerStorage() *postgres.ClubOwnerStorage {
	if s.clubOwnerStorage == nil {
		s.clubOwnerStorage = postgres.NewClubOwnerStorage(s.DB())
	}

	return s.clubOwnerStorage
}

func (s *serviceProvider) NotificationStorage() *postgres.NotificationStorage {
	if s.notificationStorage == nil {
		s.notificationStorage = postgres.NewNotificationStorage(s.DB())
	}

	return s.notificationStorage
}

// Service layer

func (s *serviceProvider) UserService() *service.UserService {
	if s.userService == nil {
		s.userService = service.NewUserService(
			s.UserStorage(),
			s.EventParticipantStorage(),
			s.SMTPClient(),
			s.cfg.App.EmailConfirmationTemplate(),
		)
	}

	return s.userService
}

func (s *serviceProvider) ClubService() *service.ClubService {
	if s.clubService == nil {
		s.clubService = service.NewClubService(s.Bot().Bot, s.ClubStorage())
	}

	return s.clubService
}

func (s *serviceProvider) EventService() *service.EventService {
	if s.eventService == nil {
		s.eventService = service.NewEventService(s.EventStorage())
	}

	return s.eventService
}

func (s *serviceProvider) EventParticipantService() *service.EventParticipantService {
	if s.eventParticipantService == nil {
		botLogger, err := logger.Named("event-participant")
		if err != nil {
			panic(fmt.Errorf("failed to create event participant logger: %w", err))
		}

		s.eventParticipantService = service.NewEventParticipantService(
			botLogger,
			s.EventParticipantStorage(),
			s.EventStorage(),
			s.PassStorage(),
			s.UserStorage(),
			s.cfg.App.PassExcludedRoles(),
		)
	}

	return s.eventParticipantService
}

func (s *serviceProvider) PassService() *service.PassService {
	if s.passService == nil {
		botLogger, err := logger.Named("pass")
		if err != nil {
			panic(fmt.Errorf("failed to create pass logger: %w", err))
		}

		s.passService = service.NewPassService(
			s.Bot().Bot,
			botLogger,
			s.PassStorage(),
			s.EventStorage(),
			s.UserStorage(),
			s.ClubStorage(),
			s.SMTPClient(),
			s.cfg.App.PassEmails(),
			s.cfg.Bot.PassChannelID(),
		)
	}

	return s.passService
}

func (s *serviceProvider) ClubOwnerService() *service.ClubOwnerService {
	if s.clubOwnerService == nil {
		s.clubOwnerService = service.NewClubOwnerService(
			s.ClubOwnerStorage(),
			s.UserStorage(),
		)
	}

	return s.clubOwnerService
}

func (s *serviceProvider) NotifyService() *service.NotifyService {
	if s.notifyService == nil {
		notifyLogger, err := logger.Named("notify")
		if err != nil {
			panic(fmt.Errorf("failed to create notify logger: %w", err))
		}

		s.notifyService = service.NewNotifyService(
			s.Bot().Bot,
			s.Bot().Layout,
			notifyLogger,
			s.ClubOwnerService(),
			s.EventStorage(),
			s.NotificationStorage(),
			s.EventParticipantStorage(),
		)
	}

	return s.notifyService
}

func (s *serviceProvider) QrService() *service.QrService {
	if s.qrService == nil {
		qrSrvc, err := service.NewQrService(
			s.Bot().Bot,
			qr.CU,
			s.UserService(),
			s.EventService(),
			s.cfg.Bot.QRChannelID(),
			s.cfg.App.QRLogoPath(),
		)
		if err != nil {
			panic(fmt.Errorf("failed to create qr service: %w", err))
		}

		s.qrService = qrSrvc
	}

	return s.qrService
}

func (s *serviceProvider) VersionService() *service.VersionService {
	if s.versionService == nil {
		versionLogger, err := logger.Named("version")
		if err != nil {
			panic(fmt.Errorf("failed to create version logger: %w", err))
		}

		s.versionService = service.NewVersionService(
			s.Bot().Bot,
			s.Bot().Layout,
			versionLogger,
		)
	}

	return s.versionService
}

// Bot dependencies

func (s *serviceProvider) Bot() *bot.Bot {
	return s.bot
}

// setBot sets the bot instance (used by App during initialization)
func (s *serviceProvider) setBot(b *bot.Bot) {
	s.bot = b
}

// Handlers

func (s *serviceProvider) AdminHandler() *admin.Handler {
	if s.adminHandler == nil {
		s.adminHandler = admin.New(
			s.UserService(),
			s.ClubService(),
			s.ClubOwnerService(),
			s.Bot().Bot,
			s.Bot().Layout,
			s.Bot().Logger,
			s.Bot().Input,
		)
	}
	return s.adminHandler
}

func (s *serviceProvider) UserHandler() *user.Handler {
	if s.userHandler == nil {
		s.userHandler = user.New(
			s.UserService(),
			s.EventService(),
			s.ClubService(),
			s.EventParticipantService(),
			s.QrService(),
			s.NotifyService(),
			s.MenuHandler(),
			s.Redis().Codes,
			s.Redis().Emails,
			s.Redis().Events,
			s.Redis().Callbacks,
			s.Bot().Layout,
			s.Bot().Logger,
			s.Bot().Input,
			s.Cfg().Bot.GrantChatID(),
			s.Cfg().App.Timezone(),
			s.Cfg().Bot.ValidEmailDomains(),
			s.Cfg().Session.EmailTTL(),
			s.Cfg().Session.AuthTTL(),
			s.Cfg().Session.ResendTTL(),
		)
	}
	return s.userHandler
}

func (s *serviceProvider) StartHandler() *start.Handler {
	if s.startHandler == nil {
		s.startHandler = start.New(
			s.UserService(),
			s.ClubService(),
			s.EventService(),
			s.EventParticipantService(),
			s.QrService(),
			s.NotifyService(),
			s.Redis().Callbacks,
			s.MenuHandler(),
			s.Redis().Codes,
			s.Redis().Emails,
			s.Redis().Events,
			s.Bot().Layout,
			s.Bot().Logger,
			s.Bot().Input,
			s.Cfg().Session.EventIDTTL(),
		)
	}
	return s.startHandler
}

func (s *serviceProvider) MenuHandler() *menu.Handler {
	if s.menuHandler == nil {
		s.menuHandler = menu.New(
			s.ClubService(),
			s.Bot().Logger,
			s.Bot().Layout,
			s.Cfg().Bot.AdminIDs(),
		)
	}
	return s.menuHandler
}

func (s *serviceProvider) MiddlewaresHandler() *middlewares.Handler {
	if s.middlewaresHandler == nil {
		s.middlewaresHandler = middlewares.New(
			s.UserService(),
			s.ClubService(),
			s.Bot().Bot,
			s.Bot().Layout,
			s.Bot().Logger,
			s.Bot().Input,
		)
	}
	return s.middlewaresHandler
}

func (s *serviceProvider) ClubOwnerHandler() *clubowner.Handler {
	if s.clubOwnerHandler == nil {
		s.clubOwnerHandler = clubowner.NewHandler(
			s.Bot().Bot,
			s.Bot().Layout,
			s.Bot().Logger,
			s.Bot().Input,
			s.Redis().Events,
			s.ClubService(),
			s.ClubOwnerService(),
			s.UserService(),
			s.EventService(),
			s.EventParticipantService(),
			s.QrService(),
			s.NotifyService(),
			s.Redis().Callbacks,
			s.Redis().Codes,
			s.Redis().Emails,
			s.Cfg().Bot.MailingChannelID(),
			s.Cfg().Bot.AvatarChannelID(),
			s.Cfg().Bot.IntroChannelID(),
			s.Cfg().App.PassLocationSubstrings(),
		)
	}
	return s.clubOwnerHandler
}

// Cfg returns the config
func (s *serviceProvider) Cfg() *config.Config {
	return s.cfg
}

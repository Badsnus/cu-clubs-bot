package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"

	"gopkg.in/gomail.v2"

	postgresStorage "github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type Config struct {
	Database   *gorm.DB
	Redis      *redis.Client
	SMTPDialer *gomail.Dialer
}

func initConfig() error {
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

	if err := os.Setenv("BOT_TOKEN", viper.GetString("bot.token")); err != nil {
		return fmt.Errorf("failed to set BOT_TOKEN environment variable: %w", err)
	}

	return nil
}

func Get() (*Config, error) {
	if err := initConfig(); err != nil {
		return nil, err
	}

	err := logger.Init(logger.Config{
		Debug:        viper.GetBool("settings.logging.debug"),
		TimeLocation: location.Location(),
		LogToFile:    viper.GetBool("settings.logging.log-to-file"),
		LogsDir:      viper.GetString("settings.logging.logs-dir"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	var gormConfig *gorm.Config
	if viper.GetBool("settings.logging.debug") {
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
	time.Local = location.Location()

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s TimeZone=%s",
		viper.GetString("service.database.user"),
		viper.GetString("service.database.password"),
		viper.GetString("service.database.name"),
		viper.GetString("service.database.host"),
		viper.GetInt("service.database.port"),
		viper.GetString("service.database.ssl-mode"),
		viper.GetString("settings.timezone"),
	)

	database, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}
	logger.Log.Infof("Successfully connected to the database (%s)", viper.GetString("service.database.name"))

	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	logger.Log.Info("Database connection pool configured")

	errMigrate := database.AutoMigrate(postgresStorage.Migrations...)
	if errMigrate != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", errMigrate)
	}

	r, err := redis.New(redis.Options{
		Host:     viper.GetString("service.redis.host"),
		Port:     viper.GetString("service.redis.port"),
		Password: viper.GetString("service.redis.password"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	dialer := gomail.NewDialer(
		viper.GetString("service.smtp.host"),
		viper.GetInt("service.smtp.port"),
		viper.GetString("service.smtp.login"),
		viper.GetString("service.smtp.pass"),
	)

	return &Config{
		Database:   database,
		Redis:      r,
		SMTPDialer: dialer,
	}, nil
}

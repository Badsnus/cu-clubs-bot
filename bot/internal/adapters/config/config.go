package config

import (
	"context"
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"os"
	"time"

	postgresStorage "github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type Config struct {
	Database   *gorm.DB
	StateRedis *redis.Client
	CodeRedis  *redis.Client
	EmailRedis *redis.Client
	SMTPDialer *gomail.Dialer
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := os.Setenv("BOT_TOKEN", viper.GetString("bot.token")); err != nil {
		panic(err)
	}
}

func Get() *Config {
	initConfig()

	err := logger.Init(logger.Config{
		Debug:     viper.GetBool("settings.logging.debug"),
		TimeZone:  viper.GetString("settings.logging.timezone"),
		LogToFile: viper.GetBool("settings.logging.log-to-file"),
		LogsDir:   viper.GetString("settings.logging.logs-dir"),
	})
	if err != nil {
		panic(err)
	}

	var gormConfig *gorm.Config
	if viper.GetBool("settings.debug") {
		newLogger := gormLogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormLogger.Config{
				SlowThreshold: time.Second,
				LogLevel:      gormLogger.Info,
				Colorful:      true,
			},
		)
		gormConfig = &gorm.Config{
			Logger: newLogger,
		}
	} else {
		gormConfig = &gorm.Config{}
	}

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable TimeZone=GMT+3",
		viper.GetString("service.database.user"),
		viper.GetString("service.database.password"),
		viper.GetString("service.database.name"),
		viper.GetString("service.database.host"),
		viper.GetInt("service.database.port"),
	)

	database, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		logger.Log.Panicf("Failed to connect to the database: %v", err)
	} else {
		logger.Log.Info("Successfully connected to the database")
	}

	errMigrate := database.AutoMigrate(postgresStorage.Migrations...)
	if errMigrate != nil {
		logger.Log.Panicf("Failed to migrate database: %v", errMigrate)
	}

	stateRedisDB := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", viper.GetString("service.redis.host"), viper.GetInt("service.redis.port")),
		Password: viper.GetString("service.redis.password"),
		DB:       0,
	})
	err = stateRedisDB.Ping(context.Background()).Err()
	if err != nil {
		logger.Log.Panicf("Failed to connect to redis: %v", err)
	} else {
		logger.Log.Info("Successfully connected to redis")
	}

	codeRedisDB := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", viper.GetString("service.redis.host"), viper.GetInt("service.redis.port")),
		Password: viper.GetString("service.redis.password"),
		DB:       1,
	})
	err = codeRedisDB.Ping(context.Background()).Err()
	if err != nil {
		logger.Log.Panicf("Failed to connect to redis: %v", err)
	} else {
		logger.Log.Info("Successfully connected to redis")
	}

	emailRedisDB := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", viper.GetString("service.redis.host"), viper.GetInt("service.redis.port")),
		Password: viper.GetString("service.redis.password"),
		DB:       2,
	})
	err = emailRedisDB.Ping(context.Background()).Err()
	if err != nil {
		logger.Log.Panicf("Failed to connect to redis: %v", err)
	} else {
		logger.Log.Info("Successfully connected to redis")
	}

	dialer := gomail.NewDialer(
		viper.GetString("service.smtp.host"),
		viper.GetInt("service.smtp.port"),
		viper.GetString("service.smtp.login"),
		viper.GetString("service.smtp.pass"),
	)

	return &Config{
		Database:   database,
		StateRedis: stateRedisDB,
		CodeRedis:  codeRedisDB,
		EmailRedis: emailRedisDB,
		SMTPDialer: dialer,
	}
}

package service

import (
	"fmt"
	"os"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

// VersionService handles version information and notifications
type VersionService struct {
	bot    *tele.Bot
	layout *layout.Layout
	logger *types.Logger
}

// NewVersionService creates a new version service
func NewVersionService(
	bot *tele.Bot,
	layout *layout.Layout,
	logger *types.Logger,
) *VersionService {
	return &VersionService{
		bot:    bot,
		layout: layout,
		logger: logger,
	}
}

// GetVersionInfo retrieves version information from environment variables
func (s *VersionService) GetVersionInfo() map[string]string {
	info := make(map[string]string)

	// Get PR name from environment variable
	prName := os.Getenv("BOT_PR_NAME")
	if prName == "" {
		prName = "Unknown"
	}
	info["pr_name"] = prName

	// Get PR url from environment variable
	prUrl := os.Getenv("BOT_PR_URL")
	info["pr_url"] = prUrl

	// Get build date from environment variable
	buildDate := os.Getenv("BOT_BUILD_DATE")
	if buildDate == "" {
		buildDate = "Unknown"
	}
	info["build_date"] = buildDate

	return info
}

// SendStartupNotification sends a notification about the bot version to the specified channel
func (s *VersionService) SendStartupNotification(channelID int64) error {
	versionInfo := s.GetVersionInfo()

	prName := versionInfo["pr_name"]
	prURL := versionInfo["pr_url"]
	buildDate := versionInfo["build_date"]

	// Log the values before sending
	s.logger.Infow("Notification details:",
		"pr_name:", prName,
		"pr_url:", prURL,
		"build_date:", buildDate)

	chat, err := s.bot.ChatByID(channelID)
	if err != nil {
		return fmt.Errorf("failed to get chat by ID %d: %w", channelID, err)
	}

	_, err = s.bot.Send(
		chat,
		s.layout.TextLocale("ru", "bot_started", struct {
			Name      string
			URL       string
			BuildDate string
			StartTime string
		}{
			Name:      prName,
			URL:       prURL,
			BuildDate: buildDate,
			StartTime: time.Now().Format("2006-01-02 15:04:05"),
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to send startup notification: %w", err)
	}

	s.logger.Infof("Sent startup notification to channel %d", channelID)
	return nil
}

package service

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"
	"strings"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"

	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	"go.uber.org/zap/zapcore"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type clubOwnerService interface {
	GetByClubID(ctx context.Context, clubID string) ([]dto.ClubOwner, error)
}

type eventStorage interface {
	GetUpcomingEvents(ctx context.Context, before time.Time) ([]entity.Event, error)
}

type notificationStorage interface {
	Create(ctx context.Context, notification *entity.EventNotification) error
	GetUnnotifiedUsers(ctx context.Context, eventID string, notificationType entity.NotificationType) ([]entity.EventParticipant, error)
}

type notifyEventParticipantStorage interface {
	GetByEventID(ctx context.Context, eventID string) ([]entity.EventParticipant, error)
}

type NotifyService struct {
	clubOwnerService              clubOwnerService
	eventStorage                  eventStorage
	notificationStorage           notificationStorage
	notifyEventParticipantStorage notifyEventParticipantStorage

	bot    *tele.Bot
	layout *layout.Layout
	logger *types.Logger
}

func NewNotifyService(
	bot *tele.Bot,
	layout *layout.Layout,
	logger *types.Logger,
	clubOwnerService clubOwnerService,
	eventStorage eventStorage,
	notificationStorage notificationStorage,
	notifyEventParticipantStorage notifyEventParticipantStorage,
) *NotifyService {
	return &NotifyService{
		clubOwnerService:              clubOwnerService,
		eventStorage:                  eventStorage,
		notificationStorage:           notificationStorage,
		notifyEventParticipantStorage: notifyEventParticipantStorage,
		bot:                           bot,
		layout:                        layout,
		logger:                        logger,
	}
}

// LogHook returns a log hook for the specified channel
//
// Parameters:
//   - channelID is the channel to send the log to
//   - locale is the locale to use for the layout
//   - level is the minimum log level to send
func (s *NotifyService) LogHook(channelID int64, locale string, level zapcore.Level) (types.LogHook, error) {
	chat, err := s.bot.ChatByID(channelID)
	if err != nil {
		return nil, err
	}
	return func(log types.Log) {
		if log.Level >= level {
			_, err = s.bot.Send(chat, s.layout.TextLocale(locale, "log", log))
			if err != nil && !strings.Contains(log.Message, "failed to send log to channel") {
				s.logger.Errorf("failed to send log to channel %d: %v\n", channelID, err)
			}
		}
	}, nil
}

// SendClubWarning sends a warning to club owners if they have enabled notifications
func (s *NotifyService) SendClubWarning(clubID string, what interface{}, opts ...interface{}) error {
	clubOwners, err := s.clubOwnerService.GetByClubID(context.Background(), clubID)
	if err != nil {
		return err
	}

	var errors []error
	for _, owner := range clubOwners {
		if owner.Warnings {
			chat, errGetChat := s.bot.ChatByID(owner.UserID)
			if errGetChat != nil {
				errors = append(errors, errGetChat)
			}
			_, errSend := s.bot.Send(chat, what, opts...)
			if errSend != nil {
				errors = append(errors, errSend)
			}
		}
	}

	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

func (s *NotifyService) SendEventUpdate(eventID string, what interface{}, opts ...interface{}) error {
	participants, err := s.notifyEventParticipantStorage.GetByEventID(context.Background(), eventID)
	if err != nil {
		return err
	}

	var errors []error
	for _, participant := range participants {
		chat, errGetChat := s.bot.ChatByID(participant.UserID)
		if errGetChat != nil {
			errors = append(errors, errGetChat)
		}
		_, errSend := s.bot.Send(chat, what, opts...)
		if errSend != nil {
			errors = append(errors, errSend)
		}
	}

	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

// StartNotifyScheduler starts the scheduler for sending notifications
func (s *NotifyService) StartNotifyScheduler() {
	s.logger.Info("Starting notify scheduler")
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			ctx := context.Background()
			s.checkAndNotify(ctx)
		}
	}()
}

// checkAndNotify checks for events starting in the next 25 hours (to cover both day and hour notifications)
//
// NOTE: localisation is hardcoded for now (ru)
func (s *NotifyService) checkAndNotify(ctx context.Context) {
	s.logger.Debugf("Checking for events starting in the next 25 hours")
	now := time.Now().In(location.Location())

	// Get events starting in the next 25 hours (to cover both day and hour notifications)
	events, err := s.eventStorage.GetUpcomingEvents(ctx, now.Add(25*time.Hour))
	if err != nil {
		s.logger.Errorf("failed to get upcoming events: %v", err)
		return
	}

	for _, event := range events {
		// Convert event time to Moscow time zone since it's stored without timezone
		eventStartTime := time.Date(
			event.StartTime.Year(),
			event.StartTime.Month(),
			event.StartTime.Day(),
			event.StartTime.Hour(),
			event.StartTime.Minute(),
			event.StartTime.Second(),
			event.StartTime.Nanosecond(),
			location.Location(),
		)

		timeUntilStart := eventStartTime.Sub(now)
		s.logger.Debugf("Event %s starts in %s", event.ID, timeUntilStart)

		// Check for day notification (between 23-24 hours before start)
		if timeUntilStart >= 23*time.Hour && timeUntilStart <= 24*time.Hour {
			s.logger.Infof("Sending day notification for event (event_id=%s)", event.ID)
			s.sendNotifications(ctx, event, entity.NotificationTypeDay)
		}

		// Check for hour notification (between 55-60 minutes before start)
		if timeUntilStart >= 55*time.Minute && timeUntilStart <= 60*time.Minute {
			s.logger.Infof("Sending hour notification for event (event_id=%s)", event.ID)
			s.sendNotifications(ctx, event, entity.NotificationTypeHour)
		}
	}
}

// sendNotifications sends notifications to users that have not been notified
func (s *NotifyService) sendNotifications(ctx context.Context, event entity.Event, notificationType entity.NotificationType) {
	// Get users who haven't been notified yet
	participants, err := s.notificationStorage.GetUnnotifiedUsers(ctx, event.ID, notificationType)
	if err != nil {
		s.logger.Errorf("failed to get unnotified users for event %s: %v", event.ID, err)
		return
	}

	for _, participant := range participants {
		s.logger.Infof(
			"Sending %s notification to user (user_id=%d, event_id=%s, notification_type=%s)",
			notificationType,
			participant.UserID,
			event.ID,
			notificationType,
		)

		// Send notification
		chat, errGetChat := s.bot.ChatByID(participant.UserID)
		if errGetChat != nil {
			s.logger.Errorf("failed to get chat for user %d: %v", participant.UserID, errGetChat)
			continue
		}

		var messageKey string
		switch notificationType {
		case entity.NotificationTypeDay:
			messageKey = "event_notification_day"
		case entity.NotificationTypeHour:
			messageKey = "event_notification_hour"
		}

		_, errSend := s.bot.Send(chat,
			s.layout.TextLocale("ru", messageKey, event),
			s.layout.MarkupLocale("ru", "core:hide"),
		)
		if errSend != nil {
			s.logger.Errorf("failed to send notification to user %d: %v", participant.UserID, errSend)
			continue
		}

		// Record that notification was sent
		notification := &entity.EventNotification{
			EventID: event.ID,
			UserID:  participant.UserID,
			Type:    notificationType,
		}

		if err := s.notificationStorage.Create(ctx, notification); err != nil {
			s.logger.Errorf("failed to create notification record: %v", err)
		}
	}
}

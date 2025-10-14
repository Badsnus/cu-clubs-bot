package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type NotificationStorage struct {
	db *gorm.DB
}

func NewNotificationStorage(db *gorm.DB) *NotificationStorage {
	return &NotificationStorage{
		db: db,
	}
}

func (s *NotificationStorage) Create(ctx context.Context, notification *entity.EventNotification) error {
	return s.db.WithContext(ctx).Create(notification).Error
}

// GetUnnotifiedUsers returns a list of users who have not been notified about an event for a specific notification type
func (s *NotificationStorage) GetUnnotifiedUsers(ctx context.Context, eventID string, notificationType entity.NotificationType) ([]entity.EventParticipant, error) {
	var participants []entity.EventParticipant

	err := s.db.WithContext(ctx).
		Joins("LEFT JOIN event_notifications ON event_notifications.user_id = event_participants.user_id AND event_notifications.event_id = event_participants.event_id AND event_notifications.type = ?", notificationType).
		Where("event_participants.event_id = ? AND event_notifications.id IS NULL", eventID).
		Find(&participants).Error

	return participants, err
}

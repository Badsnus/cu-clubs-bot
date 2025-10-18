package secondary

import (
	"context"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

// NotificationRepository defines the interface for notification data access
type NotificationRepository interface {
	Create(ctx context.Context, notification *entity.EventNotification) error
	GetUnnotifiedUsers(ctx context.Context, eventID string, notificationType entity.NotificationType) ([]entity.EventParticipant, error)
}

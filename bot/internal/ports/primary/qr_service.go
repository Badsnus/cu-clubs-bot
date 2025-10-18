package primary

import (
	"context"

	tele "gopkg.in/telebot.v3"
)

// QrService defines the interface for QR code-related use cases
type QrService interface {
	GetUserQR(ctx context.Context, userID int64) (qr tele.File, err error)
	RevokeUserQR(ctx context.Context, userID int64) error
	GetEventQR(ctx context.Context, eventID string) (qr tele.File, err error)
}

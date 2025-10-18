package secondary

import (
	"context"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

// PassRepository defines the interface for pass data access
type PassRepository interface {
	CreatePass(ctx context.Context, pass *entity.Pass) (*entity.Pass, error)
	GetPass(ctx context.Context, id string) (*entity.Pass, error)
	UpdatePass(ctx context.Context, pass *entity.Pass) (*entity.Pass, error)
	GetPassesByEventID(ctx context.Context, eventID string) ([]entity.Pass, error)
	GetPassesByUserID(ctx context.Context, userID int64, limit, offset int) ([]entity.Pass, error)
	GetPendingPassesForSchedule(ctx context.Context, before time.Time) ([]entity.Pass, error)
	MarkPassAsSent(ctx context.Context, id string, sentAt time.Time, emailSent, telegramSent bool) error
	MarkPassesAsSent(ctx context.Context, ids []string, sentAt time.Time, emailSent, telegramSent bool) error
	CreateBulkPasses(ctx context.Context, passes []entity.Pass) error
	GetActivePassForUser(ctx context.Context, eventID string, userID int64) (*entity.Pass, error)
	HasActivePass(ctx context.Context, eventID string, userID int64) (bool, error)
	CancelPassesByEventAndUser(ctx context.Context, eventID string, userID int64) error
	GetPassesByRequester(ctx context.Context, requesterType entity.PassRequesterType, requesterID string, limit, offset int) ([]entity.Pass, error)
	CountPassesByRequester(ctx context.Context, requesterType entity.PassRequesterType, requesterID string) (int64, error)
	GetPassesByEventAndRequester(ctx context.Context, eventID string, requesterType entity.PassRequesterType, requesterID string) ([]entity.Pass, error)
}

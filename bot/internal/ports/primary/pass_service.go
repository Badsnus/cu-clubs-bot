package primary

import (
	"context"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

// PassService defines the interface for pass-related use cases
type PassService interface {
	StartScheduler() error
	CreatePassForUser(
		ctx context.Context,
		eventID string,
		userID int64,
		requesterType entity.PassRequesterType,
		requesterID any,
		passType entity.PassType,
		reason string,
		scheduledAt time.Time,
	) (*entity.Pass, error)
	CreatePassByClub(
		ctx context.Context,
		eventID string,
		userID int64,
		clubID string,
		reason string,
		scheduledAt time.Time,
	) (*entity.Pass, error)
	CreatePassesByClub(
		ctx context.Context,
		eventID string,
		userIDs []int64,
		clubID string,
		reason string,
		scheduledAt time.Time,
	) ([]entity.Pass, []error)
	StopScheduler()
}

package primary

import (
	"context"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"

	tele "gopkg.in/telebot.v3"
)

// ClubService defines the interface for club-related use cases
type ClubService interface {
	Create(ctx context.Context, club *entity.Club) (*entity.Club, error)
	GetWithPagination(ctx context.Context, limit, offset int, order string) ([]entity.Club, error)
	Get(ctx context.Context, id string) (*entity.Club, error)
	GetByOwnerID(ctx context.Context, id int64) ([]entity.Club, error)
	CountByShouldShow(ctx context.Context, shouldShow bool) (int64, error)
	GetByShouldShowWithPagination(ctx context.Context, shouldShow bool, limit, offset int, order string) ([]entity.Club, error)
	Update(ctx context.Context, club *entity.Club) (*entity.Club, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
	GetAvatar(ctx context.Context, id string) (*tele.File, error)
	GetIntro(ctx context.Context, id string) (*tele.File, error)
}

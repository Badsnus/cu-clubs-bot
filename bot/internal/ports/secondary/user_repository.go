package secondary

import (
	"context"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/valueobject"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	Get(ctx context.Context, id uint) (*entity.User, error)
	GetUserByID(ctx context.Context, userID int64) (*entity.User, error)
	GetByQRCodeID(ctx context.Context, qrCodeID string) (*entity.User, error)
	GetMany(ctx context.Context, ids []int64) ([]entity.User, error)
	GetByEmail(ctx context.Context, email valueobject.Email) (*entity.User, error)
	GetAll(ctx context.Context) ([]entity.User, error)
	GetEventUsers(ctx context.Context, eventID string) ([]dto.EventUser, error)
	GetUsersByEventID(ctx context.Context, eventID string) ([]entity.User, error)
	GetUsersByClubID(ctx context.Context, clubID string) ([]entity.User, error)
	GetManyUsersByEventIDs(ctx context.Context, eventIDs []string) ([]entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	Count(ctx context.Context) (int64, error)
	GetWithPagination(ctx context.Context, limit, offset int, order string) ([]entity.User, error)
	IgnoreMailing(ctx context.Context, userID int64, clubID string) (bool, error)
}

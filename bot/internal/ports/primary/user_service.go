package primary

import (
	"context"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/valueobject"

	tele "gopkg.in/telebot.v3"
)

// UserService defines the interface for user-related use cases
type UserService interface {
	Create(ctx context.Context, user entity.User) (*entity.User, error)
	Get(ctx context.Context, userID int64) (*entity.User, error)
	GetByEmail(ctx context.Context, email valueobject.Email) (*entity.User, error)
	GetByQRCodeID(ctx context.Context, qrCodeID string) (*entity.User, error)
	GetAll(ctx context.Context) ([]entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	UpdateData(ctx context.Context, c tele.Context) (*entity.User, error)
	Count(ctx context.Context) (int64, error)
	GetWithPagination(ctx context.Context, limit int, offset int, order string) ([]entity.User, error)
	Ban(ctx context.Context, userID int64) (*entity.User, error)
	GetUsersByEventID(ctx context.Context, eventID string) ([]entity.User, error)
	GetEventUsers(ctx context.Context, eventID string) ([]dto.EventUser, error)
	GetUsersByClubID(ctx context.Context, clubID string) ([]entity.User, error)
	GetUserEvents(ctx context.Context, userID int64, limit, offset int) ([]dto.UserEvent, error)
	CountUserEvents(ctx context.Context, userID int64) (int64, error)
	SendAuthCode(ctx context.Context, email valueobject.Email, botUserName string) (string, error)
	IgnoreMailing(ctx context.Context, userID int64, clubID string) (bool, error)
	ChangeRole(ctx context.Context, userID int64, role valueobject.Role, email valueobject.Email) error
}

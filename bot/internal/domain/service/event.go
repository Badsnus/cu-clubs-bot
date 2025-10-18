package service

import (
	"context"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/ports/secondary"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type EventService struct {
	repo secondary.EventRepository
}

func NewEventService(storage secondary.EventRepository) *EventService {
	return &EventService{
		repo: storage,
	}
}

func (s *EventService) Create(ctx context.Context, event *entity.Event) (*entity.Event, error) {
	return s.repo.Create(ctx, event)
}

func (s *EventService) Get(ctx context.Context, id string) (*entity.Event, error) {
	return s.repo.Get(ctx, id)
}

func (s *EventService) GetByQRCodeID(ctx context.Context, qrCodeID string) (*entity.Event, error) {
	return s.repo.GetByQRCodeID(ctx, qrCodeID)
}

func (s *EventService) GetMany(ctx context.Context, ids []string) ([]entity.Event, error) {
	return s.repo.GetMany(ctx, ids)
}

func (s *EventService) GetAll(ctx context.Context) ([]entity.Event, error) {
	return s.repo.GetAll(ctx)
}

func (s *EventService) GetByClubID(ctx context.Context, limit, offset int, clubID string) ([]entity.Event, error) {
	return s.repo.GetByClubID(ctx, limit, offset, clubID)
}

func (s *EventService) CountByClubID(ctx context.Context, clubID string) (int64, error) {
	return s.repo.CountByClubID(ctx, clubID)
}

func (s *EventService) GetFutureByClubID(
	ctx context.Context,
	limit,
	offset int,
	order string,
	clubID string,
	additionalTime time.Duration,
) ([]entity.Event, error) {
	return s.repo.GetFutureByClubID(ctx, limit, offset, order, clubID, additionalTime)
}

// func (s *EventService) CountFutureByClubID(ctx context.Context, clubID string) (int64, error) {
//	return s.repo.CountFutureByClubID(ctx, clubID)
//}

func (s *EventService) Update(ctx context.Context, event *entity.Event) (*entity.Event, error) {
	return s.repo.Update(ctx, event)
}

func (s *EventService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *EventService) Count(ctx context.Context, role entity.Role) (int64, error) {
	return s.repo.Count(ctx, string(role))
}

func (s *EventService) GetWithPagination(ctx context.Context, limit, offset int, order string, role entity.Role, userID int64) ([]dto.Event, error) {
	return s.repo.GetWithPagination(ctx, limit, offset, order, string(role), userID)
}

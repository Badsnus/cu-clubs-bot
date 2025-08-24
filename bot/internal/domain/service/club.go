package service

import (
	"context"
	tele "gopkg.in/telebot.v3"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type ClubStorage interface {
	Create(ctx context.Context, club *entity.Club) (*entity.Club, error)
	GetWithPagination(ctx context.Context, limit, offset int, order string) ([]entity.Club, error)
	Get(ctx context.Context, id string) (*entity.Club, error)
	GetByOwnerID(ctx context.Context, id int64) ([]entity.Club, error)
	Update(ctx context.Context, club *entity.Club) (*entity.Club, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
	CountByShouldShow(ctx context.Context, shouldShow bool) (int64, error)
	GetByShouldShowWithPagination(ctx context.Context, shouldShow bool, limit, offset int, order string) ([]entity.Club, error)
}

type ClubService struct {
	bot     *tele.Bot
	storage ClubStorage
}

func NewClubService(bot *tele.Bot, storage ClubStorage) *ClubService {
	return &ClubService{
		bot:     bot,
		storage: storage,
	}
}

func (s *ClubService) Create(ctx context.Context, club *entity.Club) (*entity.Club, error) {
	return s.storage.Create(ctx, club)
}

func (s *ClubService) GetWithPagination(ctx context.Context, limit, offset int, order string) ([]entity.Club, error) {
	return s.storage.GetWithPagination(ctx, limit, offset, order)
}

func (s *ClubService) Get(ctx context.Context, id string) (*entity.Club, error) {
	return s.storage.Get(ctx, id)
}

func (s *ClubService) GetByOwnerID(ctx context.Context, id int64) ([]entity.Club, error) {
	return s.storage.GetByOwnerID(ctx, id)
}

func (s *ClubService) CountByShouldShow(ctx context.Context, shouldShow bool) (int64, error) {
	return s.storage.CountByShouldShow(ctx, shouldShow)
}

func (s *ClubService) GetByShouldShowWithPagination(ctx context.Context, shouldShow bool, limit, offset int, order string) ([]entity.Club, error) {
	return s.storage.GetByShouldShowWithPagination(ctx, shouldShow, limit, offset, order)
}

func (s *ClubService) Update(ctx context.Context, club *entity.Club) (*entity.Club, error) {
	return s.storage.Update(ctx, club)
}

func (s *ClubService) Delete(ctx context.Context, id string) error {
	return s.storage.Delete(ctx, id)
}

func (s *ClubService) Count(ctx context.Context) (int64, error) {
	return s.storage.Count(ctx)
}

func (s *ClubService) GetAvatar(ctx context.Context, id string) (*tele.File, error) {
	club, err := s.storage.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if club.AvatarID == "" {
		return nil, nil
	}

	file, err := s.bot.FileByID(club.AvatarID)
	if err != nil {
		return nil, err
	}

	return &file, nil
}

func (s *ClubService) GetIntro(ctx context.Context, id string) (*tele.File, error) {
	club, err := s.storage.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if club.IntroID == "" {
		return nil, nil
	}

	file, err := s.bot.FileByID(club.IntroID)
	if err != nil {
		return nil, err
	}

	return &file, nil
}

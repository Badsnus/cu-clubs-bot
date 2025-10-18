package service

import (
	"context"

	tele "gopkg.in/telebot.v3"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/ports/secondary"
)

type ClubService struct {
	bot  *tele.Bot
	repo secondary.ClubRepository
}

func NewClubService(bot *tele.Bot, storage secondary.ClubRepository) *ClubService {
	return &ClubService{
		bot:  bot,
		repo: storage,
	}
}

func (s *ClubService) Create(ctx context.Context, club *entity.Club) (*entity.Club, error) {
	return s.repo.Create(ctx, club)
}

func (s *ClubService) GetWithPagination(ctx context.Context, limit, offset int, order string) ([]entity.Club, error) {
	return s.repo.GetWithPagination(ctx, limit, offset, order)
}

func (s *ClubService) Get(ctx context.Context, id string) (*entity.Club, error) {
	return s.repo.Get(ctx, id)
}

func (s *ClubService) GetByOwnerID(ctx context.Context, id int64) ([]entity.Club, error) {
	return s.repo.GetByOwnerID(ctx, id)
}

func (s *ClubService) CountByShouldShow(ctx context.Context, shouldShow bool) (int64, error) {
	return s.repo.CountByShouldShow(ctx, shouldShow)
}

func (s *ClubService) GetByShouldShowWithPagination(ctx context.Context, shouldShow bool, limit, offset int, order string) ([]entity.Club, error) {
	return s.repo.GetByShouldShowWithPagination(ctx, shouldShow, limit, offset, order)
}

func (s *ClubService) Update(ctx context.Context, club *entity.Club) (*entity.Club, error) {
	return s.repo.Update(ctx, club)
}

func (s *ClubService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *ClubService) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}

func (s *ClubService) GetAvatar(ctx context.Context, id string) (*tele.File, error) {
	club, err := s.repo.Get(ctx, id)
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
	club, err := s.repo.Get(ctx, id)
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

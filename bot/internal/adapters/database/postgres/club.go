package postgres

import (
	"context"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"gorm.io/gorm"
)

type ClubStorage struct {
	db *gorm.DB
}

func NewClubStorage(db *gorm.DB) *ClubStorage {
	return &ClubStorage{
		db: db,
	}
}

func (s *ClubStorage) Create(ctx context.Context, club *entity.Club) (*entity.Club, error) {
	err := s.db.WithContext(ctx).Create(&club).Error
	return club, err
}

func (s *ClubStorage) Get(ctx context.Context, id string) (*entity.Club, error) {
	var club entity.Club
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&club).Error
	return &club, err
}

func (s *ClubStorage) Update(ctx context.Context, club *entity.Club) (*entity.Club, error) {
	err := s.db.WithContext(ctx).Save(&club).Error
	return club, err
}

// Delete is a function that deletes a club and all its events from the database.
func (s *ClubStorage) Delete(ctx context.Context, id string) error {
	err := s.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Club{}).Error
	if err != nil {
		return err
	}
	err = s.db.WithContext(ctx).Where("club_id = ?", id).Delete(&entity.Event{}).Error
	return err
}

func (s *ClubStorage) Count(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&entity.Club{}).Count(&count).Error
	return count, err
}

func (s *ClubStorage) GetWithPagination(ctx context.Context, offset, limit int, order string) ([]entity.Club, error) {
	var clubs []entity.Club
	err := s.db.WithContext(ctx).Order(order).Offset(offset).Limit(limit).Find(&clubs).Error
	return clubs, err
}

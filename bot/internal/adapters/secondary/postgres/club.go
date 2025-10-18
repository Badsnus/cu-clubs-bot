package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type ClubRepository struct {
	db *gorm.DB
}

func NewClubRepository(db *gorm.DB) *ClubRepository {
	return &ClubRepository{
		db: db,
	}
}

func (s *ClubRepository) Create(ctx context.Context, club *entity.Club) (*entity.Club, error) {
	err := s.db.WithContext(ctx).Create(&club).Error
	return club, err
}

func (s *ClubRepository) Get(ctx context.Context, id string) (*entity.Club, error) {
	var club entity.Club
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&club).Error
	return &club, err
}

func (s *ClubRepository) GetByOwnerID(ctx context.Context, id int64) ([]entity.Club, error) {
	var user entity.User
	err := s.db.WithContext(ctx).Preload("Clubs").First(&user, "id = ?", id).Error
	return user.Clubs, err
}

// GetManyByIDs is a function that get clubs by ids.
func (s *ClubRepository) GetManyByIDs(ctx context.Context, clubIDs []string) ([]entity.Club, error) {
	var clubs []entity.Club

	err := s.db.
		WithContext(ctx).
		Table("clubs").
		Select("DISTINCT ON (id) *").
		Where("id IN ?", clubIDs).
		Find(&clubs).Error
	return clubs, err
}

func (s *ClubRepository) Update(ctx context.Context, club *entity.Club) (*entity.Club, error) {
	err := s.db.WithContext(ctx).Save(&club).Error
	return club, err
}

// Delete is a function that deletes a club and all its events from the database.
func (s *ClubRepository) Delete(ctx context.Context, id string) error {
	err := s.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Club{}).Error
	if err != nil {
		return err
	}
	err = s.db.WithContext(ctx).Where("club_id = ?", id).Delete(&entity.Event{}).Error
	return err
}

func (s *ClubRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&entity.Club{}).Count(&count).Error
	return count, err
}

func (s *ClubRepository) GetWithPagination(ctx context.Context, limit, offset int, order string) ([]entity.Club, error) {
	var clubs []entity.Club
	err := s.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Find(&clubs).Error
	return clubs, err
}

func (s *ClubRepository) CountByShouldShow(ctx context.Context, shouldShow bool) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&entity.Club{}).Where("should_show = ?", shouldShow).Count(&count).Error
	return count, err
}

func (s *ClubRepository) GetByShouldShowWithPagination(ctx context.Context, shouldShow bool, limit, offset int, order string) ([]entity.Club, error) {
	var clubs []entity.Club
	err := s.db.WithContext(ctx).Where("should_show = ?", shouldShow).Order(order).Limit(limit).Offset(offset).Find(&clubs).Error
	return clubs, err
}

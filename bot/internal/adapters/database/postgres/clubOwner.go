package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type ClubOwnerStorage struct {
	db *gorm.DB
}

func NewClubOwnerStorage(db *gorm.DB) *ClubOwnerStorage {
	return &ClubOwnerStorage{
		db: db,
	}
}

func (s *ClubOwnerStorage) Create(ctx context.Context, clubOwner *entity.ClubOwner) (*entity.ClubOwner, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if club exists
		var clubExists int64
		if err := tx.Model(&entity.Club{}).Where("id = ?", clubOwner.ClubID).Count(&clubExists).Error; err != nil {
			return err
		}
		if clubExists == 0 {
			return fmt.Errorf("club with id %s not found", clubOwner.ClubID)
		}

		// Check if user exists
		var userExists int64
		if err := tx.Model(&entity.User{}).Where("id = ?", clubOwner.UserID).Count(&userExists).Error; err != nil {
			return err
		}
		if userExists == 0 {
			return fmt.Errorf("user with id %d not found", clubOwner.UserID)
		}

		// Create club owner
		return tx.Create(&clubOwner).Error
	})

	return clubOwner, err
}

func (s *ClubOwnerStorage) Delete(ctx context.Context, userID int64, clubID string) error {
	err := s.db.WithContext(ctx).Where("club_id = ? AND user_id = ?", clubID, userID).Delete(&entity.ClubOwner{}).Error
	return err
}

func (s *ClubOwnerStorage) Get(ctx context.Context, clubID string, userID int64) (*entity.ClubOwner, error) {
	var clubOwner entity.ClubOwner
	err := s.db.WithContext(ctx).Where("club_id = ? AND user_id = ?", clubID, userID).First(&clubOwner).Error
	return &clubOwner, err
}

func (s *ClubOwnerStorage) Update(ctx context.Context, clubOwner *entity.ClubOwner) (*entity.ClubOwner, error) {
	err := s.db.WithContext(ctx).Save(&clubOwner).Error
	return clubOwner, err
}

func (s *ClubOwnerStorage) GetByClubID(ctx context.Context, clubID string) ([]dto.ClubOwner, error) {
	var result []dto.ClubOwner
	err := s.db.WithContext(ctx).
		Table("club_owners").
		Select("club_owners.club_id, club_owners.user_id, users.username, club_owners.warnings, users.fio, users.email, users.role, users.is_banned").
		Joins("LEFT JOIN users ON users.id = club_owners.user_id").
		Where("club_owners.club_id = ?", clubID).
		Scan(&result).Error
	return result, err
}

func (s *ClubOwnerStorage) GetByUserID(ctx context.Context, userID int64) ([]dto.ClubOwner, error) {
	var result []dto.ClubOwner
	err := s.db.WithContext(ctx).
		Table("club_owners").
		Select("club_owners.club_id, club_owners.user_id, club_owners.warnings, users.fio, users.email, users.role, users.is_banned").
		Joins("LEFT JOIN users ON users.id = club_owners.user_id").
		Where("club_owners.user_id = ?", userID).
		Scan(&result).Error
	return result, err
}

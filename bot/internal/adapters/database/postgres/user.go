package postgres

import (
	"context"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"gorm.io/gorm"
)

type UserStorage struct {
	db *gorm.DB
}

func NewUserStorage(db *gorm.DB) *UserStorage {
	return &UserStorage{
		db: db,
	}
}

// Create is a function that creates a new user in the database.
func (s *UserStorage) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	err := s.db.WithContext(ctx).Create(&user).Error
	return user, err
}

// Get is a function that gets a user from the database by id.
func (s *UserStorage) Get(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return &user, err
}

// GetByQRCodeID is a function that gets a user from the database by qr code id.
func (s *UserStorage) GetByQRCodeID(ctx context.Context, qrCodeID string) (*entity.User, error) {
	var user entity.User
	err := s.db.WithContext(ctx).Where("qr_code_id = ?", qrCodeID).First(&user).Error
	return &user, err
}

func (s *UserStorage) GetMany(ctx context.Context, ids []int64) ([]entity.User, error) {
	var users []entity.User
	err := s.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error
	return users, err
}

func (s *UserStorage) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return &user, err
}

// GetAll is a function that gets all users from the database.
func (s *UserStorage) GetAll(ctx context.Context) ([]entity.User, error) {
	var users []entity.User
	err := s.db.WithContext(ctx).Find(&users).Error
	return users, err
}

// GetEventUsers is a function that get users with visit field by event id.
func (s *UserStorage) GetEventUsers(ctx context.Context, eventID string) ([]dto.EventUser, error) {
	type userWithQR struct {
		entity.User
		IsUserQr  bool
		IsEventQr bool
	}

	var users []userWithQR

	err := s.db.
		WithContext(ctx).
		Table("event_participants").
		Select("users.*, event_participants.is_user_qr, event_participants.is_event_qr").
		Joins("inner join users on event_participants.user_id = users.id").
		Where("event_participants.event_id = ?", eventID).
		Preload("IgnoreMailing").
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	result := make([]dto.EventUser, len(users))
	for i, user := range users {
		result[i] = dto.NewEventUserFromEntity(user.User, user.IsUserQr || user.IsEventQr)
	}

	return result, nil
}

// GetUsersByEventID is a function that get users that registered on event by event id.
func (s *UserStorage) GetUsersByEventID(ctx context.Context, eventID string) ([]entity.User, error) {
	var users []entity.User

	err := s.db.
		WithContext(ctx).
		Table("event_participants").
		Select("users.*").
		Joins("inner join users on event_participants.user_id = users.id").
		Where("event_participants.event_id = ?", eventID).
		Preload("IgnoreMailing").
		Find(&users).Error
	return users, err
}

// GetUsersByClubID is a function that returns all user that registered to club event at least once
func (s *UserStorage) GetUsersByClubID(ctx context.Context, clubID string) ([]entity.User, error) {
	var users []entity.User

	err := s.db.
		WithContext(ctx).
		Table("event_participants").
		Select("DISTINCT users.*").
		Joins("inner join users on event_participants.user_id = users.id").
		Joins("inner join events on event_participants.event_id = events.id").
		Where("events.club_id = ?", clubID).
		Preload("IgnoreMailing").
		Find(&users).Error
	return users, err
}

// GetManyUsersByEventIDs is a function that get users that registered on event by event ids without duplicates.
func (s *UserStorage) GetManyUsersByEventIDs(ctx context.Context, eventIDs []string) ([]entity.User, error) {
	var users []entity.User

	err := s.db.
		WithContext(ctx).
		Table("event_participants").
		Select("DISTINCT ON (users.id) users.*").
		Joins("inner join users on event_participants.user_id = users.id").
		Where("event_participants.event_id IN ?", eventIDs).
		Find(&users).Error
	return users, err
}

// Update is a function that updates a user in the database.
func (s *UserStorage) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	err := s.db.WithContext(ctx).Save(user).Error
	return user, err
}

// Count is a function that gets the count of users from the database.
func (s *UserStorage) Count(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&entity.User{}).Count(&count).Error
	return count, err
}

// GetWithPagination is a function that gets a list of users from the database with pagination.
func (s *UserStorage) GetWithPagination(ctx context.Context, limit, offset int, order string) ([]entity.User, error) {
	var users []entity.User
	err := s.db.WithContext(ctx).Order(order).Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

// IgnoreMailing is a function that allows or disallows mailing for a user (returns error and new state)
func (s *UserStorage) IgnoreMailing(ctx context.Context, userID int64, clubID string) (bool, error) {
	var user *entity.User
	err := s.db.WithContext(ctx).Where("id = ?", userID).Preload("IgnoreMailing").First(&user).Error
	if err != nil {
		return false, err
	}

	if user.IsMailingAllowed(clubID) {
		err = s.db.WithContext(ctx).
			Create(&entity.IgnoreMailing{
				UserID: userID,
				ClubID: clubID,
			}).Error
		return false, err
	}

	err = s.db.
		WithContext(ctx).
		Where("user_id = ? AND club_id = ?", userID, clubID).
		Delete(&entity.IgnoreMailing{}).
		Error
	return true, err
}

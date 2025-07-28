package repository

import (
	"context"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"gorm.io/gorm"
)

type StudentDataStorage struct {
	db *gorm.DB
}

func NewStudentDataStorage(db *gorm.DB) *StudentDataStorage {
	return &StudentDataStorage{
		db: db,
	}
}

// GetByLogin is a function that gets a studentData from the database by login.
func (s *StudentDataStorage) GetByLogin(ctx context.Context, login string) (*entity.StudentData, error) {
	var user entity.StudentData
	err := s.db.WithContext(ctx).Where("login = ?", login).First(&user).Error
	return &user, err
}

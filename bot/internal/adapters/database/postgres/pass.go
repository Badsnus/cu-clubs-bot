package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type PassStorage struct {
	db *gorm.DB
}

func NewPassStorage(db *gorm.DB) *PassStorage {
	return &PassStorage{
		db: db,
	}
}

// CreatePass is a function that creates a new pass in the database.
func (s *PassStorage) CreatePass(ctx context.Context, pass *entity.Pass) (*entity.Pass, error) {
	err := s.db.WithContext(ctx).Create(&pass).Error
	return pass, err
}

// GetPass is a function that gets a pass from the database by id.
func (s *PassStorage) GetPass(ctx context.Context, id string) (*entity.Pass, error) {
	var pass entity.Pass
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&pass).Error
	return &pass, err
}

// UpdatePass is a function that updates a pass in the database.
func (s *PassStorage) UpdatePass(ctx context.Context, pass *entity.Pass) (*entity.Pass, error) {
	err := s.db.WithContext(ctx).Save(pass).Error
	return pass, err
}

// GetPassesByEventID is a function that gets passes by event id.
func (s *PassStorage) GetPassesByEventID(ctx context.Context, eventID string) ([]entity.Pass, error) {
	var passes []entity.Pass
	err := s.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&passes).Error
	return passes, err
}

// GetPassesByUserID is a function that gets passes by user id with pagination.
func (s *PassStorage) GetPassesByUserID(ctx context.Context, userID int64, limit, offset int) ([]entity.Pass, error) {
	var passes []entity.Pass
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Offset(offset).Find(&passes).Error
	return passes, err
}

// GetPendingPassesForSchedule is a function that gets pending passes scheduled before the given time.
func (s *PassStorage) GetPendingPassesForSchedule(ctx context.Context, before time.Time) ([]entity.Pass, error) {
	var passes []entity.Pass
	err := s.db.WithContext(ctx).Where("status = ? AND scheduled_at <= ?", entity.PassStatusPending, before).Find(&passes).Error
	return passes, err
}

// MarkPassAsSent is a function that marks a pass as sent.
func (s *PassStorage) MarkPassAsSent(ctx context.Context, id string, sentAt time.Time, emailSent, telegramSent bool) error {
	err := s.db.WithContext(ctx).Model(&entity.Pass{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        entity.PassStatusSent,
		"sent_at":       sentAt,
		"email_sent":    emailSent,
		"telegram_sent": telegramSent,
		"updated_at":    time.Now(),
	}).Error
	return err
}

// CreateBulkPasses is a function that creates multiple passes in bulk.
func (s *PassStorage) CreateBulkPasses(ctx context.Context, passes []entity.Pass) error {
	err := s.db.WithContext(ctx).Create(&passes).Error
	return err
}

// GetActivePassForUser is a function that gets the active pass for a user and event.
func (s *PassStorage) GetActivePassForUser(ctx context.Context, eventID string, userID int64) (*entity.Pass, error) {
	var pass entity.Pass
	err := s.db.WithContext(ctx).Where("event_id = ? AND user_id = ? AND status != ?", eventID, userID, entity.PassStatusCancelled).Order("created_at desc").First(&pass).Error
	if err != nil {
		return nil, err
	}
	return &pass, nil
}

// HasActivePass is a function that checks if a user has an active pass for an event.
func (s *PassStorage) HasActivePass(ctx context.Context, eventID string, userID int64) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&entity.Pass{}).Where("event_id = ? AND user_id = ? AND status != ?", eventID, userID, entity.PassStatusCancelled).Count(&count).Error
	return count > 0, err
}

// CancelPassesByEventAndUser is a function that cancels all active passes for a user and event.
func (s *PassStorage) CancelPassesByEventAndUser(ctx context.Context, eventID string, userID int64) error {
	err := s.db.WithContext(ctx).Model(&entity.Pass{}).Where("event_id = ? AND user_id = ? AND status != ?", eventID, userID, entity.PassStatusCancelled).Updates(map[string]interface{}{
		"status":     entity.PassStatusCancelled,
		"updated_at": time.Now(),
	}).Error
	return err
}

// GetPassesByRequester is a function that gets passes by requester type and ID with pagination.
func (s *PassStorage) GetPassesByRequester(ctx context.Context, requesterType entity.PassRequesterType, requesterID string, limit, offset int) ([]entity.Pass, error) {
	var passes []entity.Pass
	err := s.db.WithContext(ctx).Where("requester_type = ? AND requester_id = ?", requesterType, requesterID).Order("created_at desc").Limit(limit).Offset(offset).Find(&passes).Error
	return passes, err
}

// CountPassesByRequester is a function that counts passes by requester type and ID.
func (s *PassStorage) CountPassesByRequester(ctx context.Context, requesterType entity.PassRequesterType, requesterID string) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&entity.Pass{}).Where("requester_type = ? AND requester_id = ?", requesterType, requesterID).Count(&count).Error
	return count, err
}

// GetPassesByEventAndRequester is a function that gets passes by event ID and requester.
func (s *PassStorage) GetPassesByEventAndRequester(ctx context.Context, eventID string, requesterType entity.PassRequesterType, requesterID string) ([]entity.Pass, error) {
	var passes []entity.Pass
	err := s.db.WithContext(ctx).Where("event_id = ? AND requester_type = ? AND requester_id = ?", eventID, requesterType, requesterID).Find(&passes).Error
	return passes, err
}

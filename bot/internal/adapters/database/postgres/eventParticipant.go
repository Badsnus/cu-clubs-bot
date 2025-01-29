package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"gorm.io/gorm"
)

type EventParticipantStorage struct {
	db *gorm.DB
}

func NewEventParticipantStorage(db *gorm.DB) *EventParticipantStorage {
	return &EventParticipantStorage{
		db: db,
	}
}

func (s *EventParticipantStorage) Create(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if event exists
		var eventExists int64
		if err := tx.Model(&entity.Event{}).Where("id = ?", eventParticipant.EventID).Count(&eventExists).Error; err != nil {
			return err
		}
		if eventExists == 0 {
			return fmt.Errorf("event with id %s not found", eventParticipant.EventID)
		}

		// Check if user exists
		var userExists int64
		if err := tx.Model(&entity.User{}).Where("id = ?", eventParticipant.UserID).Count(&userExists).Error; err != nil {
			return err
		}
		if userExists == 0 {
			return fmt.Errorf("user with id %d not found", eventParticipant.UserID)
		}

		// Create participant
		return tx.Create(&eventParticipant).Error
	})

	return eventParticipant, err
}

func (s *EventParticipantStorage) Get(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error) {
	var eventParticipant entity.EventParticipant
	err := s.db.WithContext(ctx).Where("event_id = ? AND user_id = ?", eventID, userID).First(&eventParticipant).Error
	return &eventParticipant, err
}

func (s *EventParticipantStorage) Update(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error) {
	err := s.db.WithContext(ctx).Save(&eventParticipant).Error
	return eventParticipant, err
}

func (s *EventParticipantStorage) Delete(ctx context.Context, eventID string, userID int64) error {
	err := s.db.WithContext(ctx).Where("event_id = ? AND user_id = ?", eventID, userID).Delete(&entity.EventParticipant{}).Error
	return err
}

func (s *EventParticipantStorage) GetByEventID(ctx context.Context, eventID string) ([]entity.EventParticipant, error) {
	var eventParticipants []entity.EventParticipant
	err := s.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&eventParticipants).Error
	return eventParticipants, err
}

func (s *EventParticipantStorage) CountByEventID(ctx context.Context, eventID string) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&entity.EventParticipant{}).Where("event_id = ?", eventID).Count(&count).Error
	return count, err
}

func (s *EventParticipantStorage) CountVisitedByEventID(ctx context.Context, eventID string) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&entity.EventParticipant{}).Where("event_id = ? AND (is_event_qr = true OR is_user_qr = true)", eventID).Count(&count).Error
	return count, err
}

func (s *EventParticipantStorage) GetUserEvents(ctx context.Context, userID int64, limit, offset int) ([]entity.Event, error) {
	var events []entity.Event
	currentTime := time.Now()

	// Count total upcoming events for this user
	var upcomingCount int64
	if err := s.db.WithContext(ctx).
		Model(&entity.Event{}).
		Joins("JOIN event_participants ON events.id = event_participants.event_id").
		Where("event_participants.user_id = ? AND events.start_time > ?", userID, currentTime).
		Count(&upcomingCount).Error; err != nil {
		return nil, err
	}

	// If offset is within upcoming events, get upcoming events
	if offset < int(upcomingCount) {
		if err := s.db.WithContext(ctx).
			Joins("JOIN event_participants ON events.id = event_participants.event_id").
			Where("event_participants.user_id = ? AND events.start_time > ?", userID, currentTime).
			Order("events.start_time ASC").
			Limit(limit).
			Offset(offset).
			Find(&events).Error; err != nil {
			return nil, err
		}
	}

	// If we haven't filled the limit, and there might be past events to show
	remainingLimit := limit - len(events)
	if remainingLimit > 0 {
		pastOffset := max(0, offset-int(upcomingCount)) // Adjust offset for past events
		var pastEvents []entity.Event
		if err := s.db.WithContext(ctx).
			Joins("JOIN event_participants ON events.id = event_participants.event_id").
			Where("event_participants.user_id = ? AND events.start_time <= ?", userID, currentTime).
			Order("events.start_time DESC").
			Limit(remainingLimit).
			Offset(pastOffset).
			Find(&pastEvents).Error; err != nil {
			return nil, err
		}
		events = append(events, pastEvents...)
	}

	return events, nil
}

func (s *EventParticipantStorage) CountUserEvents(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&entity.EventParticipant{}).
		Joins("JOIN events ON event_participants.event_id = events.id").
		Where("event_participants.user_id = ? AND events.deleted_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

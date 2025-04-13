package postgres

import (
	"context"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"gorm.io/gorm"
)

type EventStorage struct {
	db *gorm.DB
}

func NewEventStorage(db *gorm.DB) *EventStorage {
	return &EventStorage{
		db: db,
	}
}

// Create is a function that creates a new event in the database.
func (s *EventStorage) Create(ctx context.Context, event *entity.Event) (*entity.Event, error) {
	err := s.db.WithContext(ctx).Create(&event).Error
	return event, err
}

// Get is a function that gets an event from the database by id.
func (s *EventStorage) Get(ctx context.Context, id string) (*entity.Event, error) {
	var event entity.Event
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&event).Error
	return &event, err
}

func (s *EventStorage) GetByQRCodeID(ctx context.Context, qrCodeID string) (*entity.Event, error) {
	var event entity.Event
	err := s.db.WithContext(ctx).Where("qr_code_id = ?", qrCodeID).First(&event).Error
	return &event, err
}

func (s *EventStorage) GetMany(ctx context.Context, ids []string) ([]entity.Event, error) {
	var events []entity.Event
	err := s.db.WithContext(ctx).Where("id IN ?", ids).Find(&events).Error
	return events, err
}

// GetAll is a function that gets all events from the database.
func (s *EventStorage) GetAll(ctx context.Context) ([]entity.Event, error) {
	var events []entity.Event
	err := s.db.WithContext(ctx).Find(&events).Error
	return events, err
}

// GetByClubID is a function that gets events of a club from the database.
// It returns events in the order of start_time (upcoming first [start_time ASC], then past [start_time DESC]).
// If club has more events than limit, it returns only first limit events.
// If club has fewer events than limit, it returns all events.
// If club has no events, it returns empty list.
// If error occurs during the process, it returns error.
func (s *EventStorage) GetByClubID(ctx context.Context, limit, offset int, clubID string) ([]entity.Event, error) {
	var events []entity.Event

	currentTime := time.Now()

	// Count total upcoming events for this club.
	var upcomingCount int64
	if err := s.db.WithContext(ctx).
		Model(&entity.Event{}).
		Where("club_id = ? AND start_time > ?", clubID, currentTime).
		Count(&upcomingCount).Error; err != nil {
		return nil, err
	}

	// If offset is within upcoming events, get upcoming events
	if offset < int(upcomingCount) {
		if err := s.db.WithContext(ctx).
			Where("club_id = ? AND start_time > ?", clubID, currentTime).
			Order("start_time asc").
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
			Where("club_id = ? AND start_time <= ?", clubID, currentTime).
			Order("start_time desc").
			Limit(remainingLimit).
			Offset(pastOffset).
			Find(&pastEvents).Error; err != nil {
			return nil, err
		}
		events = append(events, pastEvents...)
	}

	return events, nil
}

// GetFutureByClubID retrieves future events for a specific club from the database.
// The events are filtered by club ID and a start time greater than the current time
// minus the additional time parameter. The results are ordered and paginated
// according to the provided parameters.
//
// Parameters:
//
//	ctx - the context for managing request-scoped values, cancellation, and timeouts
//	limit - the maximum number of events to retrieve
//	offset - the number of events to skip before starting to collect the result set
//	order - the order in which to return the events (e.g., "start_time ASC")
//	clubID - the unique identifier of the club for which to retrieve events
//	additionalTime - a duration to subtract from the current time to adjust the start time filter
//
// Returns:
//
//	A slice of entity.Event containing the future events that match the criteria, or an error if any occurs during the query.
func (s *EventStorage) GetFutureByClubID(
	ctx context.Context,
	limit, offset int,
	order string,
	clubID string,
	additionalTime time.Duration,
) ([]entity.Event, error) {
	var events []entity.Event
	err := s.db.WithContext(ctx).
		Where("club_id = ? AND start_time > ?", clubID, time.Now().In(location.Location()).Add(-additionalTime)).
		Order(order).
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	return events, err
}

// GetUpcomingEvents returns all events that start before the given time
func (s *EventStorage) GetUpcomingEvents(ctx context.Context, before time.Time) ([]entity.Event, error) {
	var events []entity.Event
	err := s.db.WithContext(ctx).
		Where("start_time <= ? AND start_time > ?", before.In(location.Location()), time.Now().In(location.Location())).
		Find(&events).Error
	return events, err
}

//func (s *EventStorage) CountFutureByClubID(ctx context.Context, clubID string) (int64, error) {
//	var count int64
//	err := s.db.WithContext(ctx).
//		Where("club_id = ? AND start_time > ?", clubID, time.Now().In(location.Location)).
//		Count(&count).Error
//	return count, err
//}

// Update is a function that updates an event in the database.
func (s *EventStorage) Update(ctx context.Context, event *entity.Event) (*entity.Event, error) {
	err := s.db.WithContext(ctx).Save(&event).Error
	return event, err
}

// Delete is a function that deletes an event from the database.
func (s *EventStorage) Delete(ctx context.Context, id string) error {
	err := s.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Event{}).Error
	return err
}

// Count is a function that gets the count of events from the database.
func (s *EventStorage) Count(ctx context.Context, role string) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&entity.Event{}).
		Where("registration_end > ?", time.Now()).
		Where("? = ANY(allowed_roles)", role)

	err := query.Count(&count).Error
	return count, err
}

func (s *EventStorage) CountByClubID(ctx context.Context, clubID string) (int64, error) {
	var count int64

	err := s.db.WithContext(ctx).Model(&entity.Event{}).
		Where("club_id = ? AND deleted_at IS NULL", clubID).
		Count(&count).Error
	return count, err
}

// GetWithPagination is a function that gets a list of events from the database with pagination. (if role is empty, it will return all events)
// If role is empty, it will return events with any role.
func (s *EventStorage) GetWithPagination(ctx context.Context, limit, offset int, order string, role string, userID int64) ([]dto.Event, error) {
	// Create the base query with all conditions
	baseQuery := s.db.WithContext(ctx).
		Model(&entity.Event{}). // Use Model() to ensure deleted_at IS NULL filter
		Where("registration_end > ?", time.Now())

	// Apply role filtering if specified
	if role != "" {
		baseQuery = baseQuery.Where("? = ANY(allowed_roles)", role)
	}

	// Apply ordering and pagination to get the correct subset of events
	var eventIDs []string
	if err := baseQuery.Order(order).Limit(limit).Offset(offset).Pluck("id", &eventIDs).Error; err != nil {
		return nil, err
	}

	// If no events match the criteria, return an empty result
	if len(eventIDs) == 0 {
		return []dto.Event{}, nil
	}

	// Get the full event data including registration status
	var events []struct {
		entity.Event
		IsRegistered bool
	}

	err := s.db.WithContext(ctx).
		Model(&entity.Event{}). // Use Model() here too for consistency
		Select("events.*, CASE WHEN ep.user_id IS NOT NULL THEN true ELSE false END as is_registered").
		Joins("LEFT JOIN event_participants ep ON events.id = ep.event_id AND ep.user_id = ?", userID).
		Where("events.id IN ?", eventIDs).
		Order(order). // Maintain the same order
		Find(&events).Error
	if err != nil {
		return nil, err
	}

	// Convert to DTOs
	result := make([]dto.Event, len(events))
	for i, event := range events {
		result[i] = dto.NewEventFromEntity(event.Event, event.IsRegistered)
	}

	return result, nil
}

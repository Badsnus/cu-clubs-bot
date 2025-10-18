package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/ports/secondary"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
)

/*
EventParticipantService - сервис для управления участниками событий.
Основные функции:
- Регистрация пользователей на события
- Автоматическое создание пропусков при регистрации на события, требующие пропуск
- Управление статусом посещения через QR-коды
- Статистика участников и их активности
*/
type EventParticipantService struct {
	logger        *types.Logger
	storage       secondary.EventParticipantRepository
	eventStorage  secondary.EventRepository
	passStorage   secondary.PassRepository
	userStorage   secondary.UserRepository
	excludedRoles []string
}

func NewEventParticipantService(
	logger *types.Logger,
	repo secondary.EventParticipantRepository,
	eventRepo secondary.EventRepository,
	passRepo secondary.PassRepository,
	userRepo secondary.UserRepository,
	excludedRoles []string,
) *EventParticipantService {
	return &EventParticipantService{
		logger:        logger,
		storage:       repo,
		eventStorage:  eventRepo,
		passStorage:   passRepo,
		userStorage:   userRepo,
		excludedRoles: excludedRoles,
	}
}

func (s *EventParticipantService) Register(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error) {
	s.logger.Debugf("Registering user %d for event %s", userID, eventID)

	participant, err := s.storage.Create(ctx, &entity.EventParticipant{
		UserID:  userID,
		EventID: eventID,
	})
	if err != nil {
		s.logger.Errorf("Failed to register user %d for event %s: %v", userID, eventID, err)
		return nil, err
	}

	if err := s.createPassIfRequired(ctx, eventID, userID); err != nil {
		s.logger.Errorf("Failed to create pass for user %d, event %s: %v", userID, eventID, err)
	}

	s.logger.Debugf("Successfully registered user %d for event %s", userID, eventID)
	return participant, nil
}

func (s *EventParticipantService) createPassIfRequired(ctx context.Context, eventID string, userID int64) error {
	event, err := s.eventStorage.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if !event.IsPassRequiredForUser(user, s.excludedRoles) {
		s.logger.Debugf("Pass not required for user %d, event %s", userID, eventID)
		return nil
	}

	hasActive, err := s.passStorage.HasActivePass(ctx, eventID, userID)
	if err != nil {
		s.logger.Errorf("Failed to check active pass for user %d, event %s: %v", userID, eventID, err)
		return err
	}

	if hasActive {
		s.logger.Debugf("Active pass already exists for user %d, event %s - skipping creation", userID, eventID)
		return nil
	}

	scheduledAt := event.CalculateScheduledAt()

	pass := &entity.Pass{
		EventID:     eventID,
		UserID:      userID,
		Type:        entity.PassTypeEvent,
		Status:      entity.PassStatusPending,
		ScheduledAt: scheduledAt,
		Reason:      "registration",
	}
	pass.SetRequester(entity.PassRequesterTypeUser, userID)

	_, err = s.passStorage.CreatePass(ctx, pass)
	if err != nil {
		s.logger.Errorf("Failed to create pass for user %d, event %s: %v", userID, eventID, err)
		return err
	}

	s.logger.Debugf("Created pass for user %d, event %s (scheduled at: %s)", userID, eventID, scheduledAt.Format("2006-01-02 15:04:05"))
	return nil
}

func (s *EventParticipantService) Get(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error) {
	return s.storage.Get(ctx, eventID, userID)
}

func (s *EventParticipantService) Update(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error) {
	participant, err := s.storage.Update(ctx, eventParticipant)
	if err != nil {
		s.logger.Errorf("Failed to update participant: %v", err)
		return nil, err
	}
	return participant, nil
}

func (s *EventParticipantService) Delete(ctx context.Context, eventID string, userID int64) error {
	if err := s.passStorage.CancelPassesByEventAndUser(ctx, eventID, userID); err != nil {
		s.logger.Errorf("Failed to cancel passes for user %d, event %s: %v", userID, eventID, err)
	}

	err := s.storage.Delete(ctx, eventID, userID)
	if err != nil {
		s.logger.Errorf("Failed to remove user %d from event %s: %v", userID, eventID, err)
		return err
	}

	return nil
}

func (s *EventParticipantService) GetByEventID(ctx context.Context, eventID string) ([]entity.EventParticipant, error) {
	return s.storage.GetByEventID(ctx, eventID)
}

func (s *EventParticipantService) CountByEventID(ctx context.Context, eventID string) (int, error) {
	count, err := s.storage.CountByEventID(ctx, eventID)
	return int(count), err
}

func (s *EventParticipantService) CountVisitedByEventID(ctx context.Context, eventID string) (int, error) {
	count, err := s.storage.CountVisitedByEventID(ctx, eventID)
	return int(count), err
}

func (s *EventParticipantService) GetUserEvents(ctx context.Context, userID int64, limit, offset int) ([]dto.UserEvent, error) {
	return s.storage.GetUserEvents(ctx, userID, limit, offset)
}

func (s *EventParticipantService) CountUserEvents(ctx context.Context, userID int64) (int64, error) {
	return s.storage.CountUserEvents(ctx, userID)
}

func (s *EventParticipantService) MarkAsVisited(ctx context.Context, eventID string, userID int64, isUserQR, isEventQR bool) error {
	participant, err := s.storage.Get(ctx, eventID, userID)
	if err != nil {
		return err
	}

	participant.IsUserQr = isUserQR
	participant.IsEventQr = isEventQR

	_, err = s.storage.Update(ctx, participant)
	return err
}

func (s *EventParticipantService) IsUserRegistered(ctx context.Context, eventID string, userID int64) (bool, error) {
	_, err := s.storage.Get(ctx, eventID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		s.logger.Errorf("failed to check user registration for event %s, user %d: %v", eventID, userID, err)
		return false, err
	}
	return true, nil
}

func (s *EventParticipantService) BulkRegister(ctx context.Context, eventID string, userIDs []int64) ([]entity.EventParticipant, error) {
	var participants []entity.EventParticipant

	for _, userID := range userIDs {
		participant, err := s.Register(ctx, eventID, userID)
		if err != nil {
			s.logger.Errorf("Failed to register user %d for event %s: %v", userID, eventID, err)
			continue
		}
		participants = append(participants, *participant)
	}

	return participants, nil
}

func (s *EventParticipantService) GetVisitedParticipants(ctx context.Context, eventID string) ([]entity.EventParticipant, error) {
	allParticipants, err := s.storage.GetByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	var visitedParticipants []entity.EventParticipant
	for _, participant := range allParticipants {
		if participant.IsUserQr || participant.IsEventQr {
			visitedParticipants = append(visitedParticipants, participant)
		}
	}

	return visitedParticipants, nil
}

func (s *EventParticipantService) GetNotVisitedParticipants(ctx context.Context, eventID string) ([]entity.EventParticipant, error) {
	allParticipants, err := s.storage.GetByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	var notVisitedParticipants []entity.EventParticipant
	for _, participant := range allParticipants {
		if !participant.IsUserQr && !participant.IsEventQr {
			notVisitedParticipants = append(notVisitedParticipants, participant)
		}
	}

	return notVisitedParticipants, nil
}

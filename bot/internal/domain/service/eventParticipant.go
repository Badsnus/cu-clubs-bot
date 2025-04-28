package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	"github.com/robfig/cron/v3"
	"github.com/xuri/excelize/v2"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"strconv"
	"strings"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
)

type EventParticipantStorage interface {
	Create(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error)
	Get(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error)
	Update(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error)
	Delete(ctx context.Context, eventID string, userID int64) error
	GetByEventID(ctx context.Context, eventID string) ([]entity.EventParticipant, error)
	CountByEventID(ctx context.Context, eventID string) (int64, error)
	CountVisitedByEventID(ctx context.Context, eventID string) (int64, error)
	GetUserEvents(ctx context.Context, userID int64, limit, offset int) ([]dto.UserEvent, error)
	CountUserEvents(ctx context.Context, userID int64) (int64, error)
}

type eventParticipantEventStorage interface {
	GetUpcomingEvents(ctx context.Context, before time.Time) ([]entity.Event, error)
}

type userStorage interface {
	GetManyUsersByEventIDs(ctx context.Context, eventIDs []string) ([]entity.User, error)
}

type eventParticipantSMTPClient interface {
	Send(to string, body, message string, subject string, file *bytes.Buffer)
}

type clubStorage interface {
	GetManyByIDs(ctx context.Context, clubIDs []string) ([]entity.Club, error)
}

type EventParticipantService struct {
	bot    *tele.Bot
	layout *layout.Layout
	logger *types.Logger

	storage                    EventParticipantStorage
	eventStorage               eventParticipantEventStorage
	userStorage                userStorage
	clubStorage                clubStorage
	eventParticipantSMTPClient eventParticipantSMTPClient

	passEmails []string
	passChatID int64
}

func NewEventParticipantService(
	bot *tele.Bot,
	layout *layout.Layout,
	logger *types.Logger,
	storage EventParticipantStorage,
	eventStorage eventParticipantEventStorage,
	userStorage userStorage,
	clubStorage clubStorage,
	eventParticipantSMTPClient eventParticipantSMTPClient,
	passEmails []string,
	passChatID int64,
) *EventParticipantService {
	return &EventParticipantService{
		bot:    bot,
		layout: layout,
		logger: logger,

		storage:                    storage,
		eventStorage:               eventStorage,
		userStorage:                userStorage,
		clubStorage:                clubStorage,
		eventParticipantSMTPClient: eventParticipantSMTPClient,

		passEmails: passEmails,
		passChatID: passChatID,
	}
}

func (s *EventParticipantService) Register(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error) {
	return s.storage.Create(ctx, &entity.EventParticipant{
		UserID:  userID,
		EventID: eventID,
	})
}

func (s *EventParticipantService) Get(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error) {
	return s.storage.Get(ctx, eventID, userID)
}

func (s *EventParticipantService) Update(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error) {
	return s.storage.Update(ctx, eventParticipant)
}

func (s *EventParticipantService) Delete(ctx context.Context, eventID string, userID int64) error {
	return s.storage.Delete(ctx, eventID, userID)
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

func (s *EventParticipantService) StartPassScheduler() {
	s.logger.Info("Starting pass scheduler")
	go func() {
		c := cron.New(cron.WithLocation(location.Location()))
		if _, err := c.AddFunc("1 16 * * 1-5", func() {
			ctx := context.Background()
			s.checkAndSend(ctx)
		}); err != nil {
			s.logger.Error("Failed to start pass scheduler:", err)
			return
		}

		if _, err := c.AddFunc("1 12 * * 6", func() {
			ctx := context.Background()
			s.checkAndSend(ctx)
		}); err != nil {
			s.logger.Error("Failed to start pass scheduler:", err)
			return
		}
		c.Start()

		select {}
	}()
}

func (s *EventParticipantService) checkAndSend(ctx context.Context) {
	s.logger.Debugf("Checking for events starting in the next 25 hours")
	now := time.Now().In(location.Location())

	events, err := s.eventStorage.GetUpcomingEvents(ctx, now.Add(72*time.Hour))
	if err != nil {
		s.logger.Errorf("failed to get upcoming events: %v", err)
		return
	}

	var eventIDs []string
	var clubsIDs []string
	var eventStartDate time.Time

	for _, event := range events {
		eventStartTime := event.StartTime.In(location.Location())
		weekday := eventStartTime.Weekday()

		// Determine notification time
		var notificationTime time.Time
		if weekday == time.Sunday {
			notificationTime = time.Date(eventStartTime.Year(), eventStartTime.Month(), eventStartTime.Day()-1, 12, 0, 0, 0, location.Location())
		} else if weekday == time.Monday {
			notificationTime = time.Date(eventStartTime.Year(), eventStartTime.Month(), eventStartTime.Day()-2, 12, 0, 0, 0, location.Location())
		} else {
			notificationTime = time.Date(eventStartTime.Year(), eventStartTime.Month(), eventStartTime.Day(), 0, 0, 0, 0, eventStartTime.Location()).Add(-24 * time.Hour).Add(16 * time.Hour)
		}

		// Check if it's valid event for notification and it's time to send the notification
		if (strings.Contains(event.Location, "Гашека 7")) && !(now.Before(notificationTime) || now.After(notificationTime.Add(1*time.Hour))) {
			eventIDs = append(eventIDs, event.ID)
			clubsIDs = append(clubsIDs, event.ClubID)
			eventStartTime = event.StartTime
		}
	}

	var participants []entity.User
	participants, err = s.userStorage.GetManyUsersByEventIDs(ctx, eventIDs)
	if err != nil {
		s.logger.Errorf("failed to get participants for events %s: %v", eventIDs, err)
		return
	}

	var participantsWithoutStudents []entity.User
	for _, participant := range participants {
		if !(participant.Role == entity.Student) {
			participantsWithoutStudents = append(participantsWithoutStudents, participant)
		}
	}
	if len(participantsWithoutStudents) == 0 {
		return
	}

	clubs, err := s.clubStorage.GetManyByIDs(context.Background(), clubsIDs)
	if err != nil {
		s.logger.Errorf("failed to get clubs: %v", err)
		return
	}
	clubsName := make([]string, len(clubs))
	for i, club := range clubs {
		clubsName[i] = club.Name
	}
	clubsNameStr := strings.Join(clubsName, ", ")

	var buf *bytes.Buffer
	buf, err = participantsToXLSX(participantsWithoutStudents)
	if err != nil {
		s.logger.Errorf("failed to form xlsx with participants %s: %v", eventIDs, err)
		return
	}

	message := fmt.Sprintf("Внешние гости_%s_%s", clubsNameStr, eventStartDate.Format("02.01.2006"))

	for _, passEmail := range s.passEmails {
		s.eventParticipantSMTPClient.Send(passEmail, message, message, message, buf)
	}

	chat, errGetChat := s.bot.ChatByID(s.passChatID)
	if errGetChat != nil {
		s.logger.Errorf("failed to get chat %d: %v", s.passChatID, errGetChat)
		return
	}

	file := &tele.Document{
		File:     tele.FromReader(buf),
		Caption:  s.layout.TextLocale("ru", "pass_users"),
		FileName: "users.xlsx",
	}
	_, errSend := s.bot.Send(chat, file)
	if errSend != nil {
		s.logger.Errorf("failed to send notification to chat %d: %v", s.passChatID, errSend)
		return
	}
}

func participantsToXLSX(users []entity.User) (*bytes.Buffer, error) {
	f := excelize.NewFile()

	sheet := "Sheet1"
	_ = f.SetCellValue(sheet, "A1", "Фамилия")
	_ = f.SetCellValue(sheet, "B1", "Имя")
	_ = f.SetCellValue(sheet, "C1", "Отчество")
	for i, user := range users {
		fio := strings.Split(user.FIO, " ")
		if len(fio) != 3 {
			continue
		}

		row := i + 2
		_ = f.SetCellValue(sheet, "A"+strconv.Itoa(row), fio[0])
		_ = f.SetCellValue(sheet, "B"+strconv.Itoa(row), fio[1])
		_ = f.SetCellValue(sheet, "C"+strconv.Itoa(row), fio[2])
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return &buf, nil
}

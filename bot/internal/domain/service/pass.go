package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/xuri/excelize/v2"
	tele "gopkg.in/telebot.v3"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
)

/*
PassService - сервис управления пропусками для событий.

Основные принципы работы:
- Пропуски создаются автоматически при регистрации пользователя на событие, которое требует пропуск
- Планировщик отправляет уже созданные пропуски согласно расписанию
- Поддерживается отправка через email и Telegram
- Гибкая система конфигураций для разных типов отправки (будни/выходные)
*/

type PassConfig struct {
	Name            string
	EmailRecipients []string
	TelegramChatID  int64
	TimeBeforeEvent time.Duration
	IsActive        bool
	CronSchedule    string
}

type PassStorage interface {
	CreatePass(ctx context.Context, pass *entity.Pass) (*entity.Pass, error)
	GetPass(ctx context.Context, id string) (*entity.Pass, error)
	UpdatePass(ctx context.Context, pass *entity.Pass) (*entity.Pass, error)
	GetPassesByEventID(ctx context.Context, eventID string) ([]entity.Pass, error)
	GetPassesByUserID(ctx context.Context, userID int64, limit, offset int) ([]entity.Pass, error)
	GetPendingPassesForSchedule(ctx context.Context, before time.Time) ([]entity.Pass, error)
	MarkPassAsSent(ctx context.Context, id string, sentAt time.Time, emailSent, telegramSent bool) error
	MarkPassesAsSent(ctx context.Context, ids []string, sentAt time.Time, emailSent, telegramSent bool) error
	CreateBulkPasses(ctx context.Context, passes []entity.Pass) error

	GetActivePassForUser(ctx context.Context, eventID string, userID int64) (*entity.Pass, error)
	HasActivePass(ctx context.Context, eventID string, userID int64) (bool, error)

	GetPassesByRequester(ctx context.Context, requesterType entity.PassRequesterType, requesterID string, limit, offset int) ([]entity.Pass, error)
	CountPassesByRequester(ctx context.Context, requesterType entity.PassRequesterType, requesterID string) (int64, error)
	GetPassesByEventAndRequester(ctx context.Context, eventID string, requesterType entity.PassRequesterType, requesterID string) ([]entity.Pass, error)
}

type PassEventStorage interface {
	GetEventByID(ctx context.Context, eventID string) (*entity.Event, error)
}

type PassUserStorage interface {
	GetManyUsersByEventIDs(ctx context.Context, eventIDs []string) ([]entity.User, error)
	GetUserByID(ctx context.Context, userID int64) (*entity.User, error)
	GetMany(ctx context.Context, ids []int64) ([]entity.User, error)
}

type PassClubStorage interface {
	GetManyByIDs(ctx context.Context, clubIDs []string) ([]entity.Club, error)
}

type PassSMTPClient interface {
	Send(to string, body, message string, subject string, file *bytes.Buffer) error
}

type EventWithPasses struct {
	Event  entity.Event
	Passes []entity.Pass
}

type PassService struct {
	bot    *tele.Bot
	logger *types.Logger

	passStorage  PassStorage
	eventStorage PassEventStorage
	userStorage  PassUserStorage
	clubStorage  PassClubStorage
	smtpClient   PassSMTPClient

	cron             *cron.Cron
	configs          map[string]*PassConfig
	schedulerStarted bool
}

func NewPassService(
	bot *tele.Bot,
	logger *types.Logger,
	passStorage PassStorage,
	eventStorage PassEventStorage,
	userStorage PassUserStorage,
	clubStorage PassClubStorage,
	smtpClient PassSMTPClient,
	passEmails []string,
	telegramChatID int64,
) *PassService {
	ps := &PassService{
		bot:              bot,
		logger:           logger,
		passStorage:      passStorage,
		eventStorage:     eventStorage,
		userStorage:      userStorage,
		clubStorage:      clubStorage,
		smtpClient:       smtpClient,
		cron:             cron.New(cron.WithLocation(location.Location())),
		configs:          make(map[string]*PassConfig),
		schedulerStarted: false,
	}

	weekdayConfig := &PassConfig{
		Name:            "weekday",
		EmailRecipients: passEmails,
		TelegramChatID:  telegramChatID,
		TimeBeforeEvent: 24 * time.Hour,
		IsActive:        true,
		CronSchedule:    "0 16 * * 1-5",
	}
	ps.configs["weekday"] = weekdayConfig

	weekendConfig := &PassConfig{
		Name:            "weekend",
		EmailRecipients: passEmails,
		TelegramChatID:  telegramChatID,
		TimeBeforeEvent: 48 * time.Hour,
		IsActive:        true,
		CronSchedule:    "0 12 * * 6,0",
	}
	ps.configs["weekend"] = weekendConfig

	//testConfig := &PassConfig{
	//	Name:            "test",
	//	EmailRecipients: passEmails,
	//	TelegramChatID:  telegramChatID,
	//	TimeBeforeEvent: 1 * time.Hour,
	//	IsActive:        true,
	//	CronSchedule:    "* * * * *",
	//}
	//ps.configs["test"] = testConfig

	return ps
}

func (s *PassService) GetConfig(name string) *PassConfig {
	if config, exists := s.configs[name]; exists {
		return config
	}
	return nil
}

func (s *PassService) GetConfigs() map[string]*PassConfig {
	return s.configs
}

func (s *PassService) GetActiveConfigs() map[string]*PassConfig {
	activeConfigs := make(map[string]*PassConfig)
	for name, config := range s.configs {
		if config.IsActive {
			activeConfigs[name] = config
		}
	}
	return activeConfigs
}

// CreatePassForUser создает пропуск для пользователя с проверкой на дублирование
func (s *PassService) CreatePassForUser(
	ctx context.Context,
	eventID string,
	userID int64,
	requesterType entity.PassRequesterType,
	requesterID any,
	passType entity.PassType,
	reason string,
	scheduledAt time.Time,
) (*entity.Pass, error) {
	hasActive, err := s.passStorage.HasActivePass(ctx, eventID, userID)
	if err != nil {
		s.logger.Errorf("Failed to check active pass for user %d, event %s: %v", userID, eventID, err)
		return nil, fmt.Errorf("failed to check active pass: %w", err)
	}

	if hasActive {
		s.logger.Debugf("Active pass already exists for user %d, event %s", userID, eventID)
		return nil, fmt.Errorf("active pass already exists for this user and event")
	}

	pass := &entity.Pass{
		EventID:     eventID,
		UserID:      userID,
		Type:        passType,
		Status:      entity.PassStatusPending,
		Reason:      reason,
		ScheduledAt: scheduledAt,
	}
	pass.SetRequester(requesterType, requesterID)

	created, err := s.passStorage.CreatePass(ctx, pass)
	if err != nil {
		s.logger.Errorf("Failed to create pass for user %d, event %s: %v", userID, eventID, err)
		return nil, fmt.Errorf("failed to create pass: %w", err)
	}

	s.logger.Debugf("Created pass %s for user %d, event %s (type: %s, requester: %s)", created.ID, userID, eventID, passType, requesterType)
	return created, nil
}

// CreatePassByAdmin создает пропуск вручную от имени администратора
func (s *PassService) CreatePassByAdmin(
	ctx context.Context,
	eventID string,
	userID int64,
	adminID int64,
	reason string,
	scheduledAt time.Time,
) (*entity.Pass, error) {
	return s.CreatePassForUser(
		ctx,
		eventID,
		userID,
		entity.PassRequesterTypeAdmin,
		adminID,
		entity.PassTypeManual,
		reason,
		scheduledAt,
	)
}

// CreatePassesByAdmin создает пропуски для нескольких пользователей от имени администратора
func (s *PassService) CreatePassesByAdmin(
	ctx context.Context,
	eventID string,
	userIDs []int64,
	adminID int64,
	reason string,
	scheduledAt time.Time,
) ([]entity.Pass, []error) {
	var passes []entity.Pass
	var errors []error

	for _, userID := range userIDs {
		pass, err := s.CreatePassByAdmin(ctx, eventID, userID, adminID, reason, scheduledAt)
		if err != nil {
			errors = append(errors, fmt.Errorf("user %d: %w", userID, err))
			continue
		}
		passes = append(passes, *pass)
	}

	return passes, errors
}

// CreatePassByClub создает пропуск от имени клуба через API
func (s *PassService) CreatePassByClub(
	ctx context.Context,
	eventID string,
	userID int64,
	clubID string,
	reason string,
	scheduledAt time.Time,
) (*entity.Pass, error) {
	return s.CreatePassForUser(
		ctx,
		eventID,
		userID,
		entity.PassRequesterTypeClub,
		clubID,
		entity.PassTypeApi,
		reason,
		scheduledAt,
	)
}

// CreatePassesByClub создает пропуски для нескольких пользователей от имени клуба
func (s *PassService) CreatePassesByClub(
	ctx context.Context,
	eventID string,
	userIDs []int64,
	clubID string,
	reason string,
	scheduledAt time.Time,
) ([]entity.Pass, []error) {
	var passes []entity.Pass
	var errors []error

	for _, userID := range userIDs {
		pass, err := s.CreatePassByClub(ctx, eventID, userID, clubID, reason, scheduledAt)
		if err != nil {
			errors = append(errors, fmt.Errorf("user %d: %w", userID, err))
			continue
		}
		passes = append(passes, *pass)
	}

	return passes, errors
}

func (s *PassService) GetPassesByEventID(ctx context.Context, eventID string) ([]entity.Pass, error) {
	return s.passStorage.GetPassesByEventID(ctx, eventID)
}

func (s *PassService) GetPassesByUserID(ctx context.Context, userID int64, limit, offset int) ([]entity.Pass, error) {
	return s.passStorage.GetPassesByUserID(ctx, userID, limit, offset)
}

// GetPassesByAdmin получает пропуски, созданные конкретным администратором
func (s *PassService) GetPassesByAdmin(ctx context.Context, adminID int64, limit, offset int) ([]entity.Pass, error) {
	requesterID := fmt.Sprintf("%d", adminID)
	return s.passStorage.GetPassesByRequester(ctx, entity.PassRequesterTypeAdmin, requesterID, limit, offset)
}

// GetPassesByClub получает пропуски, созданные конкретным клубом
func (s *PassService) GetPassesByClub(ctx context.Context, clubID string, limit, offset int) ([]entity.Pass, error) {
	return s.passStorage.GetPassesByRequester(ctx, entity.PassRequesterTypeClub, clubID, limit, offset)
}

// GetActivePassForUser получает активный пропуск для пользователя и события
func (s *PassService) GetActivePassForUser(ctx context.Context, eventID string, userID int64) (*entity.Pass, error) {
	return s.passStorage.GetActivePassForUser(ctx, eventID, userID)
}

// HasActivePass проверяет наличие активного пропуска
func (s *PassService) HasActivePass(ctx context.Context, eventID string, userID int64) (bool, error) {
	return s.passStorage.HasActivePass(ctx, eventID, userID)
}

func (s *PassService) StartScheduler() error {
	s.logger.Debug("Initializing pass scheduler...")

	for _, config := range s.configs {
		if !config.IsActive || config.CronSchedule == "" {
			s.logger.Debugf("Skipping config %s (active: %v, schedule: %s)", config.Name, config.IsActive, config.CronSchedule)
			continue
		}

		configName := config.Name
		s.logger.Debugf("Adding cron job for config %s with schedule: %s", configName, config.CronSchedule)

		_, err := s.cron.AddFunc(config.CronSchedule, func() {
			s.logger.Debugf("=== CRON TRIGGERED for %s ===", configName)
			s.processPendingPasses(context.Background(), configName)
		})
		if err != nil {
			return fmt.Errorf("failed to add cron job for config %s: %w", config.Name, err)
		}
		s.logger.Debugf("Successfully added cron job for config %s", configName)
	}

	s.cron.Start()
	s.schedulerStarted = true
	entries := s.cron.Entries()
	s.logger.Infof("Pass scheduler started with %d jobs", len(entries))
	for i, entry := range entries {
		s.logger.Debugf("Job #%d: next run at %s", i+1, entry.Next.Format("2006-01-02 15:04:05"))
	}
	return nil
}

func (s *PassService) StopScheduler() {
	if s.cron != nil {
		s.cron.Stop()
		s.schedulerStarted = false
		s.logger.Info("Pass scheduler stopped")
	}
}

func (s *PassService) processPendingPasses(ctx context.Context, configName string) {
	s.logger.Debugf("Processing pending passes for config: %s", configName)

	config := s.GetConfig(configName)
	if config == nil || !config.IsActive {
		s.logger.Debugf("Config %s not found or inactive", configName)
		return
	}

	cutoffTime := time.Now().Add(config.TimeBeforeEvent)
	s.logger.Debugf("Looking for pending passes before: %s", cutoffTime.Format("2006-01-02 15:04:05"))

	pendingPasses, err := s.passStorage.GetPendingPassesForSchedule(ctx, cutoffTime)
	if err != nil {
		s.logger.Error("Failed to get pending passes", "error", err)
		return
	}

	s.logger.Debugf("Found %d pending passes", len(pendingPasses))

	var eventsWithPasses []EventWithPasses
	if len(pendingPasses) > 0 {
		eventsWithPasses = s.groupPassesByEvent(ctx, pendingPasses)
	}

	telegramSent, emailSent, err := s.sendConsolidatedPassNotification(ctx, eventsWithPasses, config)
	if err != nil {
		s.logger.Error("Failed to send consolidated notification", "error", err)
		return
	}

	if len(pendingPasses) > 0 {
		sentAt := time.Now()
		var passIDs []string
		for _, eventWithPasses := range eventsWithPasses {
			for _, pass := range eventWithPasses.Passes {
				passIDs = append(passIDs, pass.ID)
			}
		}
		if len(passIDs) > 0 {
			if err := s.passStorage.MarkPassesAsSent(ctx, passIDs, sentAt, emailSent, telegramSent); err != nil {
				s.logger.Error("Failed to mark passes as sent", "error", err)
			}
		}
	}

	s.logger.Infow("Processed pending passes",
		"events", len(eventsWithPasses),
		"totalPasses", len(pendingPasses),
		"config", configName)
}

func (s *PassService) groupPassesByEvent(ctx context.Context, passes []entity.Pass) []EventWithPasses {
	eventPassesMap := make(map[string][]entity.Pass)
	eventMap := make(map[string]entity.Event)

	for _, pass := range passes {
		eventPassesMap[pass.EventID] = append(eventPassesMap[pass.EventID], pass)

		if _, exists := eventMap[pass.EventID]; !exists {
			event, err := s.eventStorage.GetEventByID(ctx, pass.EventID)
			if err != nil {
				s.logger.Error("Failed to get event", "eventID", pass.EventID, "error", err)
				continue
			}
			eventMap[pass.EventID] = *event
		}
	}

	var result []EventWithPasses
	for eventID, eventPasses := range eventPassesMap {
		if event, exists := eventMap[eventID]; exists {
			result = append(result, EventWithPasses{
				Event:  event,
				Passes: eventPasses,
			})
		}
	}

	return result
}

func (s *PassService) sendConsolidatedPassNotification(ctx context.Context, eventsWithPasses []EventWithPasses, config *PassConfig) (telegramSent bool, emailSent bool, err error) {
	totalPasses := 0
	for _, eventWithPasses := range eventsWithPasses {
		totalPasses += len(eventWithPasses.Passes)
	}

	message := s.formatConsolidatedPassMessage(eventsWithPasses, totalPasses)

	var consolidatedExcel *bytes.Buffer
	if totalPasses > 0 {
		consolidatedExcel, err = s.generateConsolidatedPassExcel(ctx, eventsWithPasses)
		if err != nil {
			s.logger.Error("Failed to generate consolidated Excel file", "error", err)
			return false, false, err
		}
	} else {
		consolidatedExcel, err = s.generateEmptyPassExcel()
		if err != nil {
			s.logger.Error("Failed to generate empty Excel file", "error", err)
			return false, false, err
		}
	}

	if config.TelegramChatID != 0 {
		buf := bytes.NewBuffer(consolidatedExcel.Bytes())
		if sendErr := s.sendTelegramNotification(config.TelegramChatID, message, buf); sendErr != nil {
			s.logger.Error("Failed to send consolidated Telegram notification", "error", sendErr)
			telegramSent = false
		} else {
			telegramSent = true
		}
	}

	if len(config.EmailRecipients) > 0 {
		subject := fmt.Sprintf("Сводка пропусков - %d событий (%d пропусков)",
			len(eventsWithPasses), totalPasses)

		emailSent = true
		for _, email := range config.EmailRecipients {
			buf := bytes.NewBuffer(consolidatedExcel.Bytes())
			if sendErr := s.smtpClient.Send(email, message, message, subject, buf); sendErr != nil {
				s.logger.Error("Failed to send email", "email", email, "error", sendErr)
				emailSent = false
			}
		}
	}

	return telegramSent, emailSent, nil
}

func (s *PassService) formatConsolidatedPassMessage(eventsWithPasses []EventWithPasses, totalPasses int) string {
	var message strings.Builder

	message.WriteString("📋 <b>Сводка пропусков</b>\n\n")

	if totalPasses == 0 {
		message.WriteString("✅ <b>Нет пропусков для отправки</b>\n\n")
		return message.String()
	}

	message.WriteString(fmt.Sprintf("📊 <b>Всего событий:</b> %d\n", len(eventsWithPasses)))
	message.WriteString(fmt.Sprintf("👥 <b>Всего пропусков:</b> %d\n\n", totalPasses))

	for i, eventWithPasses := range eventsWithPasses {
		event := eventWithPasses.Event
		passes := eventWithPasses.Passes

		message.WriteString(fmt.Sprintf("<b>%d. %s</b>\n", i+1, event.Name))
		message.WriteString(fmt.Sprintf("📅 %s\n", event.StartTime.In(location.Location()).Format("02.01.2006 15:04")))
		message.WriteString(fmt.Sprintf("📍 %s\n", event.Location))
		message.WriteString(fmt.Sprintf("👥 Пропусков: %d\n\n", len(passes)))
	}

	return message.String()
}

func (s *PassService) generateConsolidatedPassExcel(ctx context.Context, eventsWithPasses []EventWithPasses) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			s.logger.Errorf("Failed to close Excel file: %v", err)
		}
	}()

	sheetName := "Пропуски"
	if err := f.SetSheetName("Sheet1", sheetName); err != nil {
		return nil, fmt.Errorf("failed to set sheet name: %w", err)
	}

	headers := []string{"Событие", "Дата", "Время", "Место", "ФИО", "Роль"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, fmt.Errorf("failed to set header cell: %w", err)
		}
	}

	row := 2
	for _, eventWithPasses := range eventsWithPasses {
		event := eventWithPasses.Event
		passes := eventWithPasses.Passes

		var userIDs []int64
		for _, pass := range passes {
			userIDs = append(userIDs, pass.UserID)
		}

		users, err := s.userStorage.GetMany(ctx, userIDs)
		if err != nil {
			s.logger.Error("Failed to get users for Excel", "error", err)
			continue
		}

		userMap := make(map[int64]entity.User)
		for _, user := range users {
			userMap[user.ID] = user
		}

		for _, pass := range passes {
			user, exists := userMap[pass.UserID]
			if !exists {
				continue
			}

			data := []any{
				event.Name,
				event.StartTime.In(location.Location()).Format("02.01.2006"),
				event.StartTime.In(location.Location()).Format("15:04"),
				event.Location,
				user.FIO,
				user.Role,
			}

			for i, value := range data {
				cell, _ := excelize.CoordinatesToCellName(i+1, row)
				if err := f.SetCellValue(sheetName, cell, value); err != nil {
					s.logger.Errorf("Failed to set cell value: %v", err)
					continue
				}
			}
			row++
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return &buf, nil
}

func (s *PassService) sendTelegramNotification(chatID int64, message string, file *bytes.Buffer) error {
	if file != nil && file.Len() > 0 {
		document := &tele.Document{
			File:     tele.FromReader(file),
			FileName: fmt.Sprintf("passes_%s.xlsx", time.Now().Format("2006-01-02")),
		}

		document.Caption = message
		_, err := s.bot.Send(&tele.Chat{ID: chatID}, document, &tele.SendOptions{ParseMode: tele.ModeHTML})
		return err
	}

	_, err := s.bot.Send(&tele.Chat{ID: chatID}, message, &tele.SendOptions{ParseMode: tele.ModeHTML})
	return err
}

func (s *PassService) generateEmptyPassExcel() (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			s.logger.Errorf("Failed to close Excel file: %v", err)
		}
	}()

	sheetName := "Пропуски"
	if err := f.SetSheetName("Sheet1", sheetName); err != nil {
		return nil, fmt.Errorf("failed to set sheet name: %w", err)
	}

	headers := []string{"Событие", "Дата", "Время", "Место", "ФИО", "Роль"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, fmt.Errorf("failed to set header cell: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return &buf, nil
}

func (s *PassService) GetPassStatisticsForEvent(ctx context.Context, eventID string) (map[string]int, error) {
	passes, err := s.passStorage.GetPassesByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	stats := map[string]int{
		"total":     len(passes),
		"pending":   0,
		"sent":      0,
		"cancelled": 0,
	}

	for _, pass := range passes {
		switch pass.Status {
		case entity.PassStatusPending:
			stats["pending"]++
		case entity.PassStatusSent:
			stats["sent"]++
		case entity.PassStatusCancelled:
			stats["cancelled"]++
		}
	}

	return stats, nil
}

func (s *PassService) GetSchedulerStatus() bool {
	return s.schedulerStarted
}

func (s *PassService) GetSchedulerInfo() map[string]any {
	info := map[string]any{
		"status":        s.GetSchedulerStatus(),
		"configs":       len(s.configs),
		"activeConfigs": len(s.GetActiveConfigs()),
	}

	if s.cron != nil {
		entries := s.cron.Entries()
		info["scheduledJobs"] = len(entries)

		var nextRuns []string
		for _, entry := range entries {
			nextRuns = append(nextRuns, entry.Next.Format("2006-01-02 15:04:05"))
		}
		info["nextRuns"] = nextRuns
	}

	configDetails := make(map[string]any)
	for name, config := range s.configs {
		configDetails[name] = map[string]any{
			"active":          config.IsActive,
			"schedule":        config.CronSchedule,
			"timeBeforeEvent": config.TimeBeforeEvent.String(),
			"emailRecipients": len(config.EmailRecipients),
			"telegramChatID":  config.TelegramChatID,
		}
	}
	info["configDetails"] = configDetails

	return info
}

package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/smtp"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/dto"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"

	tele "gopkg.in/telebot.v3"
)

type UserStorage interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	Get(ctx context.Context, id uint) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByQRCodeID(ctx context.Context, qrCodeID string) (*entity.User, error)
	GetMany(ctx context.Context, ids []int64) ([]entity.User, error)
	GetAll(ctx context.Context) ([]entity.User, error)
	GetEventUsers(ctx context.Context, eventID string) ([]dto.EventUser, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	Count(ctx context.Context) (int64, error)
	GetWithPagination(ctx context.Context, limit int, offset int, order string) ([]entity.User, error)
	GetUsersByEventID(ctx context.Context, eventID string) ([]entity.User, error)
	GetUsersByClubID(ctx context.Context, clubID string) ([]entity.User, error)
	IgnoreMailing(ctx context.Context, userID int64, clubID string) (bool, error)
}

type smtpClient interface {
	Send(to string, body, message string, subject string, file *bytes.Buffer)
}

type eventParticipantStorage interface {
	GetUserEvents(ctx context.Context, userID int64, limit, offset int) ([]dto.UserEvent, error)
	CountUserEvents(ctx context.Context, userID int64) (int64, error)
}

type UserService struct {
	userStorage             UserStorage
	eventParticipantStorage eventParticipantStorage
	smtpClient              smtpClient

	emailHTMLFilePath string
}

func NewUserService(userStorage UserStorage, eventParticipantStorage eventParticipantStorage, smtpClient smtpClient, emailHTMLFilePath string) *UserService {
	return &UserService{
		userStorage:             userStorage,
		eventParticipantStorage: eventParticipantStorage,
		smtpClient:              smtpClient,

		emailHTMLFilePath: emailHTMLFilePath,
	}
}

func (s *UserService) Create(ctx context.Context, user entity.User) (*entity.User, error) {
	return s.userStorage.Create(ctx, &user)
}

func (s *UserService) Get(ctx context.Context, userID int64) (*entity.User, error) {
	return s.userStorage.Get(ctx, uint(userID))
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	return s.userStorage.GetByEmail(ctx, email)
}

func (s *UserService) GetByQRCodeID(ctx context.Context, qrCodeID string) (*entity.User, error) {
	return s.userStorage.GetByQRCodeID(ctx, qrCodeID)
}

func (s *UserService) GetAll(ctx context.Context) ([]entity.User, error) {
	return s.userStorage.GetAll(ctx)
}

func (s *UserService) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	return s.userStorage.Update(ctx, user)
}

func (s *UserService) UpdateData(ctx context.Context, c tele.Context) (*entity.User, error) {
	user, err := s.Get(ctx, c.Sender().ID)
	if err != nil {
		return nil, err
	}
	user.ID = c.Sender().ID

	return s.userStorage.Update(ctx, user)
}

func (s *UserService) Count(ctx context.Context) (int64, error) {
	return s.userStorage.Count(ctx)
}

func (s *UserService) GetWithPagination(ctx context.Context, limit int, offset int, order string) ([]entity.User, error) {
	return s.userStorage.GetWithPagination(ctx, limit, offset, order)
}

func (s *UserService) Ban(ctx context.Context, userID int64) (*entity.User, error) {
	user, err := s.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.IsBanned = true
	return s.Update(ctx, user)
}

func (s *UserService) GetUsersByEventID(ctx context.Context, eventID string) ([]entity.User, error) {
	return s.userStorage.GetUsersByEventID(ctx, eventID)
}

func (s *UserService) GetEventUsers(ctx context.Context, eventID string) ([]dto.EventUser, error) {
	return s.userStorage.GetEventUsers(ctx, eventID)
}

func (s *UserService) GetUsersByClubID(ctx context.Context, clubID string) ([]entity.User, error) {
	return s.userStorage.GetUsersByClubID(ctx, clubID)
}

func (s *UserService) GetUserEvents(ctx context.Context, userID int64, limit, offset int) ([]dto.UserEvent, error) {
	return s.eventParticipantStorage.GetUserEvents(ctx, userID, limit, offset)
}

func (s *UserService) CountUserEvents(ctx context.Context, userID int64) (int64, error) {
	return s.eventParticipantStorage.CountUserEvents(ctx, userID)
}

func (s *UserService) SendAuthCode(_ context.Context, email string, botUserName string) (code string, err error) {
	link, code, err := generateAuthLink(12, botUserName)
	if err != nil {
		return "", err
	}

	message, err := smtp.GenerateEmailConfirmationMessage(s.emailHTMLFilePath, map[string]string{
		"AuthLink": link,
	})
	if err != nil {
		return "", err
	}

	s.smtpClient.Send(email, "Email confirmation", message, "Email confirmation", nil)

	return code, nil
}

// IgnoreMailing is a function that allows or disallows mailing for a user (returns error and new state)
func (s *UserService) IgnoreMailing(ctx context.Context, userID int64, clubID string) (bool, error) {
	return s.userStorage.IgnoreMailing(ctx, userID, clubID)
}

func generateAuthLink(codeLength int, botUserName string) (link string, code string, err error) {
	code, err = generateRandomCode(codeLength)
	if err != nil {
		return link, code, err
	}

	return fmt.Sprintf("https://t.me/%s?start=auth_%s", botUserName, code), code, err
}

func generateRandomCode(length int) (string, error) {
	bts := make([]byte, length)
	if _, err := rand.Read(bts); err != nil {
		return "", err
	}
	return hex.EncodeToString(bts)[:length], nil
}

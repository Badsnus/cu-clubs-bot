package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"

	tele "gopkg.in/telebot.v3"
)

type UserStorage interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	Get(ctx context.Context, id uint) (*entity.User, error)
	GetByQRCodeID(ctx context.Context, qrCodeID string) (*entity.User, error)
	GetMany(ctx context.Context, ids []int64) ([]entity.User, error)
	GetAll(ctx context.Context) ([]entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	Count(ctx context.Context) (int64, error)
	GetWithPagination(ctx context.Context, limit int, offset int, order string) ([]entity.User, error)
}

type StudentDataStorage interface {
	GetByLogin(ctx context.Context, login string) (*entity.StudentData, error)
}

type smtpClient interface {
	SendConfirmationEmail(to string, code string)
}

type eventParticipantService interface {
	GetUserEvents(ctx context.Context, userID int64, limit, offset int) ([]entity.Event, error)
	CountUserEvents(ctx context.Context, userID int64) (int64, error)
}

type UserService struct {
	userStorage             UserStorage
	studentDataStorage      StudentDataStorage
	eventParticipantService eventParticipantService
	smtpClient              smtpClient
}

func NewUserService(userStorage UserStorage, studentDataStorage StudentDataStorage, eventParticipantService eventParticipantService, smtpClient smtpClient) *UserService {
	return &UserService{
		userStorage:             userStorage,
		studentDataStorage:      studentDataStorage,
		eventParticipantService: eventParticipantService,
		smtpClient:              smtpClient,
	}
}

func (s *UserService) Create(ctx context.Context, user entity.User) (*entity.User, error) {
	return s.userStorage.Create(ctx, &user)
}

func (s *UserService) Get(ctx context.Context, userID int64) (*entity.User, error) {
	return s.userStorage.Get(ctx, uint(userID))
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

func (s *UserService) GetUserEvents(ctx context.Context, userID int64, limit, offset int) ([]entity.Event, error) {
	return s.eventParticipantService.GetUserEvents(ctx, userID, limit, offset)
}

func (s *UserService) CountUserEvents(ctx context.Context, userID int64) (int64, error) {
	return s.eventParticipantService.CountUserEvents(ctx, userID)
}

func (s *UserService) SendAuthCode(_ context.Context, email string) (string, string, error) {
	code, err := generateRandomCode(12)
	if err != nil {
		return "", "", err
	}

	login := strings.Split(email, "@")[0]

	var data string
	studentData, err := s.studentDataStorage.GetByLogin(context.Background(), login)
	if err == nil {
		s.smtpClient.SendConfirmationEmail(email, code)
		data = fmt.Sprintf("%s;%s", email, studentData.Fio)
	}

	return data, code, nil
}

func generateRandomCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

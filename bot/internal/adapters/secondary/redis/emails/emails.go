package emails

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/valueobject"
)

type Storage struct {
	redis *redis.Client
}

func NewStorage(client *redis.Client) *Storage {
	return &Storage{
		redis: client,
	}
}

type Email struct {
	Email        string       `json:"email"`
	EmailContext EmailContext `json:"email_context"`
}

type EmailContext struct {
	FIO valueobject.FIO `json:"fio"`
}

func (s *Storage) Get(userID int64) (Email, error) {
	emailData, err := s.redis.Get(context.Background(), fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		return Email{}, err
	}

	var email Email
	if err := json.Unmarshal([]byte(emailData), &email); err != nil {
		return Email{}, fmt.Errorf("failed to unmarshal email data: %w", err)
	}

	return email, nil
}

func (s *Storage) Set(userID int64, email string, emailContext EmailContext, expiration time.Duration) error {
	emailData := Email{
		Email:        email,
		EmailContext: emailContext,
	}

	jsonData, err := json.Marshal(emailData)
	if err != nil {
		return fmt.Errorf("failed to marshal email data: %w", err)
	}

	return s.redis.Set(context.Background(), fmt.Sprintf("%d", userID), jsonData, expiration).Err()
}

func (s *Storage) Clear(userID int64) error {
	return s.redis.Del(context.Background(), fmt.Sprintf("%d", userID)).Err()
}

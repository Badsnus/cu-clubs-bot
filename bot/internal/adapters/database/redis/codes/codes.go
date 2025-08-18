package codes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Storage struct {
	redis *redis.Client
}

func NewStorage(client *redis.Client) *Storage {
	return &Storage{
		redis: client,
	}
}

type Code struct {
	Code        string      `json:"code"`
	CodeContext CodeContext `json:"code_context"`
	NextResend  time.Time   `json:"next_resend"`
}

type CodeContext struct {
	Email string `json:"email"`
	FIO   string `json:"fio"`
}

func (s *Storage) Get(userID int64) (Code, error) {
	codeData, err := s.redis.Get(context.Background(), fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		return Code{}, err
	}

	var code Code
	if err := json.Unmarshal([]byte(codeData), &code); err != nil {
		return Code{}, fmt.Errorf("failed to unmarshal code data: %w", err)
	}

	return code, nil
}

func (s *Storage) GetCanResend(userID int64) (bool, time.Duration, error) {
	code, err := s.Get(userID)
	if errors.Is(err, redis.Nil) {
		return true, 0, nil
	}
	if err != nil {
		return false, 0, err
	}

	now := time.Now()
	log.Println(now.UTC(), code.NextResend.UTC())
	canResend := now.UTC().After(code.NextResend.UTC())

	if canResend {
		return true, 0, nil
	}

	timeLeft := code.NextResend.UTC().Sub(now.UTC())
	return false, timeLeft, nil
}

func (s *Storage) Set(
	userID int64,
	code string,
	codeContext CodeContext,
	expiration time.Duration,
	resendCooldown time.Duration,
) error {
	codeData := Code{
		Code:        code,
		CodeContext: codeContext,
		NextResend:  time.Now().UTC().Add(resendCooldown),
	}

	jsonData, err := json.Marshal(codeData)
	if err != nil {
		return fmt.Errorf("failed to marshal code data: %w", err)
	}

	return s.redis.Set(context.Background(), fmt.Sprintf("%d", userID), jsonData, expiration).Err()
}

func (s *Storage) Clear(userID int64) error {
	return s.redis.Del(context.Background(), fmt.Sprintf("%d", userID)).Err()
}

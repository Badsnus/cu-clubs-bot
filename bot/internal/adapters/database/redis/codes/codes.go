package codes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/common/errorz"
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
	Code        string
	CodeContext string
}

func (s *Storage) Get(userID int64) (Code, error) {
	codeData, err := s.redis.Get(context.Background(), fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		return Code{}, err
	}
	codeSlice := strings.Split(codeData, ":")
	if len(codeSlice) == 1 {
		return Code{
			Code:        codeSlice[0],
			CodeContext: "",
		}, nil
	}

	if len(codeSlice) == 2 {
		return Code{
			Code:        codeSlice[0],
			CodeContext: codeSlice[1],
		}, nil
	}

	return Code{}, errorz.ErrInvalidCode
}

func (s *Storage) GetCanResend(userID int64) (bool, error) {
	_, err := s.redis.Get(context.Background(), fmt.Sprintf("can_resend:%d", userID)).Result()
	if errors.Is(err, redis.Nil) {
		return true, nil
	}
	return false, err
}

func (s *Storage) Set(userID int64, code string, codeContext string, expiration time.Duration) {
	s.redis.Set(context.Background(), fmt.Sprintf("%d", userID), fmt.Sprintf("%s:%s", code, codeContext), expiration)
}

func (s *Storage) SetCanResend(userID int64, canResend bool, expiration time.Duration) {
	s.redis.Set(context.Background(), fmt.Sprintf("can_resend:%d", userID), canResend, expiration)
}

func (s *Storage) Clear(userID int64) {
	s.redis.Del(context.Background(), fmt.Sprintf("%d", userID))
}

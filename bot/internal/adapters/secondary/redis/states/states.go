package states

import (
	"context"
	"fmt"
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

func (s *Storage) Get(userID int64) (string, error) {
	data, err := s.redis.Get(context.Background(), fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		return "", err
	}
	return data, nil
}

func (s *Storage) Set(userID int64, state string, expiration time.Duration) error {
	return s.redis.Set(context.Background(), fmt.Sprintf("%d", userID), state, expiration).Err()
}

func (s *Storage) Delete(userID int64) {
	s.redis.Del(context.Background(), fmt.Sprintf("%d", userID))
}

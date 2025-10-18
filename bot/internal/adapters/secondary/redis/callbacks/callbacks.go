package callbacks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// CallbackStorage - is a storage for callbacks
//
// NOTE:
// Use this if you need to store sensitive data in callback or if your callback is too long
type CallbackStorage interface {
	Get(callbackID string) (string, error)
	Set(data string, expiration time.Duration) (string, error)
	Delete(callbackID string)
}

// Storage is a realization of CallbackStorage
type Storage struct {
	redis *redis.Client
}

func NewStorage(client *redis.Client) *Storage {
	return &Storage{
		redis: client,
	}
}

func (s *Storage) Get(callbackID string) (string, error) {
	data, err := s.redis.Get(context.Background(), callbackID).Result()
	if err != nil {
		return "", err
	}
	return data, nil
}

// Set stores callback data in redis at random uuid key (callbackID).
// Returns callbackID and error
func (s *Storage) Set(data string, expiration time.Duration) (string, error) {
	callbackID := uuid.New().String()
	err := s.redis.Set(context.Background(), callbackID, data, expiration).Err()
	if err != nil {
		return "", err
	}
	return callbackID, nil
}

func (s *Storage) Delete(callbackID string) {
	s.redis.Del(context.Background(), callbackID)
}

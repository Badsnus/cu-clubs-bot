package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"

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

func (s *Storage) Get(userID int64) (entity.Event, error) {
	eventBytes, err := s.redis.Get(context.Background(), fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		return entity.Event{}, err
	}

	var event entity.Event
	if err = json.Unmarshal([]byte(eventBytes), &event); err != nil {
		return entity.Event{}, err
	}

	return event, nil
}

func (s *Storage) GetEventID(userID int64, key string) (string, error) {
	eventID, err := s.redis.Get(context.Background(), fmt.Sprintf("%s:%d", key, userID)).Result()
	if err != nil {
		return "", err
	}

	return eventID, nil
}

func (s *Storage) Set(userID int64, event entity.Event, expiration time.Duration) {
	eventBytes, _ := json.Marshal(event)
	s.redis.Set(context.Background(), fmt.Sprintf("%d", userID), eventBytes, expiration)
}

func (s *Storage) SetEventID(userID int64, key, eventID string, expiration time.Duration) {
	s.redis.Set(context.Background(), fmt.Sprintf("%s:%d", key, userID), eventID, expiration)
}

func (s *Storage) Clear(userID int64) {
	s.redis.Del(context.Background(), fmt.Sprintf("%d", userID))
}

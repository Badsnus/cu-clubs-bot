package dto

import "github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"

type ClubOwner struct {
	ClubID   string
	UserID   int64
	FIO      string
	Email    string
	Role     entity.Role
	IsBanned bool
}
package dto

import (
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/valueobject"
)

type ClubOwner struct {
	ClubID   string
	UserID   int64
	Username string
	FIO      valueobject.FIO
	Email    valueobject.Email
	Role     entity.Role
	IsBanned bool
	Warnings bool
}

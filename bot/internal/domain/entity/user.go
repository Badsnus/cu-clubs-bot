package entity

import (
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/valueobject"
)

type (
	Role  = valueobject.Role
	Roles = valueobject.Roles
)

type User struct {
	ID            int64 `gorm:"primaryKey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Localisation  string `gorm:"default:ru"`
	Username      string
	Role          Role              `gorm:"not null"`
	Email         valueobject.Email `gorm:"uniqueIndex:idx_users_email,where:email <> ''"`
	FIO           valueobject.FIO   `gorm:"not null"`
	QRCodeID      string
	QRFileID      string
	IsBanned      bool            `gorm:"default:false"`
	Clubs         []Club          `gorm:"many2many:club_owners;foreignKey:ID;joinForeignKey:UserID;References:ID;JoinReferences:ClubID"`
	IgnoreMailing []IgnoreMailing `gorm:"foreignKey:UserID;references:ID"`
}

type ClubOwner struct {
	UserID    int64  `gorm:"primaryKey"`
	ClubID    string `gorm:"primaryKey;type:uuid"`
	Warnings  bool
	CreatedAt time.Time
}

type EventParticipant struct {
	EventID   string `gorm:"primaryKey;type:uuid"`
	UserID    int64  `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	IsUserQr  bool
	IsEventQr bool
}

type IgnoreMailing struct {
	UserID    int64  `gorm:"primaryKey"`
	ClubID    string `gorm:"primaryKey;type:uuid"`
	CreatedAt time.Time
}

type FIO = valueobject.FIO

func (u *User) GetFIO() FIO {
	return u.FIO
}

func (u *User) GetEmail() valueobject.Email {
	return u.Email
}

func (u *User) IsMailingAllowed(clubID string) bool {
	for _, ignoreMailing := range u.IgnoreMailing {
		if ignoreMailing.ClubID == clubID {
			return false
		}
	}
	return true
}

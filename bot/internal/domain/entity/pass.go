package entity

import (
	"fmt"
	"time"
)

type PassStatus string

const (
	PassStatusPending   PassStatus = "pending"
	PassStatusSent      PassStatus = "sent"
	PassStatusCancelled PassStatus = "cancelled"
)

type PassType string

const (
	PassTypeEvent  PassType = "event"
	PassTypeManual PassType = "manual"
	PassTypeApi    PassType = "api"
)

type PassRequesterType string

const (
	PassRequesterTypeUser  PassRequesterType = "user"  // Запрос от пользователя (самостоятельная регистрация)
	PassRequesterTypeAdmin PassRequesterType = "admin" // Запрос от администратора
	PassRequesterTypeClub  PassRequesterType = "club"  // Запрос от клуба через API
)

type Pass struct {
	ID        string `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CreatedAt time.Time
	UpdatedAt time.Time

	EventID string     `gorm:"type:uuid;not null;index"`
	UserID  int64      `gorm:"not null;index"`
	Type    PassType   `gorm:"not null;default:'event'"`
	Status  PassStatus `gorm:"not null;default:'pending'"`

	RequesterType PassRequesterType `gorm:"type:varchar(20);not null;default:'user';index"`
	RequesterID   string            `gorm:"not null;index"` // ID запросчика (int64 как string для user/admin, UUID для club)

	// Расписание и отправка
	ScheduledAt time.Time
	SentAt      *time.Time

	// Дополнительная информация
	Reason       string // Причина создания пропуска
	Notes        string // Дополнительная информация
	EmailSent    bool   `gorm:"default:false"`
	TelegramSent bool   `gorm:"default:false"`
}

func (p *Pass) IsExpired() bool {
	if p.Status == PassStatusSent {
		return false
	}
	return time.Now().After(p.ScheduledAt.Add(24 * time.Hour))
}

func (p *Pass) CanBeSent() bool {
	return p.Status == PassStatusPending && !p.IsExpired()
}

func (p *Pass) MarkAsSent(emailSent, telegramSent bool) {
	p.Status = PassStatusSent
	now := time.Now()
	p.SentAt = &now
	p.EmailSent = emailSent
	p.TelegramSent = telegramSent
	p.UpdatedAt = now
}

func (p *Pass) Cancel() {
	p.Status = PassStatusCancelled
	p.UpdatedAt = time.Now()
}

func (p *Pass) SetRequester(requesterType PassRequesterType, requesterID any) {
	p.RequesterType = requesterType
	switch v := requesterID.(type) {
	case int64:
		p.RequesterID = fmt.Sprintf("%d", v)
	case string:
		p.RequesterID = v
	default:
		p.RequesterID = fmt.Sprintf("%v", v)
	}
}

func (p *Pass) GetRequesterUserID() (int64, error) {
	if p.RequesterType != PassRequesterTypeUser && p.RequesterType != PassRequesterTypeAdmin {
		return 0, fmt.Errorf("requester is not a user or admin")
	}
	var id int64
	_, err := fmt.Sscanf(p.RequesterID, "%d", &id)
	return id, err
}

func (p *Pass) GetRequesterClubID() (string, error) {
	if p.RequesterType != PassRequesterTypeClub {
		return "", fmt.Errorf("requester is not a club")
	}
	return p.RequesterID, nil
}

func (p *Pass) IsRequestedByUser() bool {
	return p.RequesterType == PassRequesterTypeUser
}

func (p *Pass) IsRequestedByAdmin() bool {
	return p.RequesterType == PassRequesterTypeAdmin
}

func (p *Pass) IsRequestedByClub() bool {
	return p.RequesterType == PassRequesterTypeClub
}

package start

import (
	"context"
	"errors"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/callbacks"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/banner"
	qr "github.com/Badsnus/cu-clubs-bot/bot/pkg/qrcode"
	"github.com/spf13/viper"
	"strings"
	"time"

	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/handlers/menu"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/codes"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/emails"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/smtp"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"gorm.io/gorm"
)

type userService interface {
	Create(ctx context.Context, user entity.User) (*entity.User, error)
	Get(ctx context.Context, userID int64) (*entity.User, error)
	GetByQRCodeID(ctx context.Context, qrCodeID string) (*entity.User, error)
	SendAuthCode(ctx context.Context, email string) (string, string, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
}

type clubService interface {
	GetByOwnerID(ctx context.Context, ownerID int64) ([]entity.Club, error)
}

type eventService interface {
	Get(ctx context.Context, id string) (*entity.Event, error)
	GetFutureByClubID(
		ctx context.Context,
		limit, offset int,
		order string,
		clubID string,
		additionalTime time.Duration,
	) ([]entity.Event, error)
	//CountFutureByClubID(ctx context.Context, clubID string) (int64, error)
}

type eventParticipantService interface {
	Register(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error)
	Get(ctx context.Context, eventID string, userID int64) (*entity.EventParticipant, error)
	Update(ctx context.Context, eventParticipant *entity.EventParticipant) (*entity.EventParticipant, error)
	CountByEventID(ctx context.Context, eventID string) (int, error)
}

type qrService interface {
	RevokeUserQR(ctx context.Context, userID int64) error
}

type Handler struct {
	userService             userService
	clubService             clubService
	eventService            eventService
	eventParticipantService eventParticipantService
	qrService               qrService

	callbacksStorage callbacks.CallbackStorage

	menuHandler *menu.Handler

	codesStorage  *codes.Storage
	emailsStorage *emails.Storage
	layout        *layout.Layout
	logger        *types.Logger
}

func New(b *bot.Bot) *Handler {
	userStorage := postgres.NewUserStorage(b.DB)
	studentDataStorage := postgres.NewStudentDataStorage(b.DB)
	eventStorage := postgres.NewEventStorage(b.DB)
	clubStorage := postgres.NewClubStorage(b.DB)
	eventParticipantStorage := postgres.NewEventParticipantStorage(b.DB)
	smtpClient := smtp.NewClient(b.SMTPDialer)

	userSrvc := service.NewUserService(userStorage, studentDataStorage, nil, smtpClient)
	qrSrvc, err := service.NewQrService(
		b.Bot,
		qr.CU,
		userSrvc,
		viper.GetInt64("bot.qr.chat-id"),
		viper.GetString("settings.qr.logo-path"),
	)
	if err != nil {
		b.Logger.Fatalf("failed to create qr service: %v", err)
	}

	return &Handler{
		userService:             userSrvc,
		clubService:             service.NewClubService(clubStorage),
		eventService:            service.NewEventService(eventStorage),
		eventParticipantService: service.NewEventParticipantService(eventParticipantStorage),
		qrService:               qrSrvc,
		callbacksStorage:        b.Redis.Callbacks,
		menuHandler:             menu.New(b),
		codesStorage:            b.Redis.Codes,
		emailsStorage:           b.Redis.Emails,
		layout:                  b.Layout,
		logger:                  b.Logger,
	}
}

func (h Handler) Start(c tele.Context) error {
	h.logger.Infof("(user: %d) press start button", c.Sender().ID)

	user, err := h.userService.Get(context.Background(), c.Sender().ID)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		h.logger.Errorf("(user: %d) error while getting user from db: %v", c.Sender().ID, err)
		return c.Send(
			h.layout.Text(c, "technical_issues", err.Error()),
			h.layout.Markup(c, "core:hide"),
		)
	}

	payload := strings.Split(c.Message().Payload, "_")

	if len(payload) < 2 {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Send(
				banner.Auth.Caption(h.layout.Text(c, "personal_data_agreement_text")),
				h.layout.Markup(c, "auth:personalData:agreementMenu"),
			)
		}
		if user.IsBanned {
			return c.Send(
				h.layout.Text(c, "banned"),
				h.layout.Markup(c, "core:hide"),
			)
		}
		return h.menuHandler.SendMenu(c)
	}

	var payloadType, data string
	if len(payload) == 2 {
		payloadType, data = payload[0], payload[1]
	}

	switch payloadType {
	case "auth":
		return h.auth(c, data)
	case "userQR":
		return h.userQR(c, data)
	case "event":
		return h.eventMenu(c, data)
	default:
		return c.Send(
			h.layout.Text(c, "something_went_wrong"),
			h.layout.Markup(c, "core:hide"),
		)
	}
}

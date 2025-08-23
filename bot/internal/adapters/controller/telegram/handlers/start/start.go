package start

import (
	"context"
	"errors"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/callbacks"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/redis/events"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/banner"
	qr "github.com/Badsnus/cu-clubs-bot/bot/pkg/qrcode"
	"github.com/redis/go-redis/v9"
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
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"gorm.io/gorm"
)

type userService interface {
	Create(ctx context.Context, user entity.User) (*entity.User, error)
	Get(ctx context.Context, userID int64) (*entity.User, error)
	GetByQRCodeID(ctx context.Context, qrCodeID string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	ChangeRole(
		ctx context.Context,
		userID int64,
		role entity.Role,
		email string,
	) error
}

type clubService interface {
	Get(ctx context.Context, id string) (*entity.Club, error)
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
	GetByQRCodeID(ctx context.Context, qrCodeID string) (*entity.Event, error)
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

type notificationService interface {
	SendClubWarning(clubID string, what interface{}, opts ...interface{}) error
}

type Handler struct {
	userService             userService
	clubService             clubService
	eventService            eventService
	eventParticipantService eventParticipantService
	qrService               qrService
	notificationService     notificationService

	callbacksStorage callbacks.CallbackStorage

	menuHandler *menu.Handler

	codesStorage  *codes.Storage
	emailsStorage *emails.Storage
	eventsStorage *events.Storage
	layout        *layout.Layout
	logger        *types.Logger
}

func New(b *bot.Bot) *Handler {
	userStorage := postgres.NewUserStorage(b.DB)
	eventStorage := postgres.NewEventStorage(b.DB)
	clubStorage := postgres.NewClubStorage(b.DB)
	eventParticipantStorage := postgres.NewEventParticipantStorage(b.DB)
	clubOwnerStorage := postgres.NewClubOwnerStorage(b.DB)
	notificationStorage := postgres.NewNotificationStorage(b.DB)

	userSrvc := service.NewUserService(userStorage, nil, nil, "")
	eventSrvc := service.NewEventService(eventStorage)
	clubOwnerSrvc := service.NewClubOwnerService(clubOwnerStorage, userStorage)

	qrSrvc, err := service.NewQrService(
		b.Bot,
		qr.CU,
		userSrvc,
		eventSrvc,
		viper.GetInt64("bot.qr.channel-id"),
		viper.GetString("settings.qr.logo-path"),
	)
	if err != nil {
		b.Logger.Fatalf("failed to create qr service: %v", err)
	}

	return &Handler{
		userService:             userSrvc,
		clubService:             service.NewClubService(b.Bot, clubStorage),
		eventService:            eventSrvc,
		eventParticipantService: service.NewEventParticipantService(b.Bot, b.Layout, b.Logger, eventParticipantStorage, nil, nil, nil, nil, nil, 0),
		qrService:               qrSrvc,
		notificationService:     service.NewNotifyService(b.Bot, b.Layout, b.Logger, clubOwnerSrvc, eventStorage, notificationStorage, eventParticipantStorage),
		callbacksStorage:        b.Redis.Callbacks,
		menuHandler:             menu.New(b),
		codesStorage:            b.Redis.Codes,
		emailsStorage:           b.Redis.Emails,
		eventsStorage:           b.Redis.Events,
		layout:                  b.Layout,
		logger:                  b.Logger,
	}
}

func (h Handler) Start(c tele.Context) error {
	h.logger.Infof("(user: %d) enter /start", c.Sender().ID)

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
	case "emailCode":
		code, err := h.codesStorage.Get(c.Sender().ID)
		if err != nil && !errors.Is(err, redis.Nil) {
			h.logger.Errorf("(user: %d) error while getting user email from redis: %v", c.Sender().ID, err)
			return c.Send(
				h.layout.Text(c, "wrong_code"),
				h.layout.Markup(c, "core:hide"),
			)
		}
		switch code.Type {
		case codes.CodeTypeAuth:
			return h.auth(c, data)
		case codes.CodeTypeChangingRole:
			return h.changeRole(c, data)
		default:
			h.logger.Errorf("(user: %d) invalid code type: %v", c.Sender().ID, code.Type)
			return c.Send(
				h.layout.Text(c, "something_went_wrong"),
				h.layout.Markup(c, "core:hide"),
			)
		}

	case "userQR":
		return h.userQR(c, data)

	case "eventQR":
		return h.eventQR(c, data)

	case "event":
		return h.eventMenu(c, data)

	default:
		return c.Send(
			h.layout.Text(c, "something_went_wrong"),
			h.layout.Markup(c, "core:hide"),
		)
	}
}

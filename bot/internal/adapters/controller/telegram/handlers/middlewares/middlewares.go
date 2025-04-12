package middlewares

import (
	"context"
	"errors"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/banner"
	"strings"

	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger/types"
	"github.com/nlypage/intele"
	"gorm.io/gorm"

	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/entity"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/service"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type userService interface {
	Get(ctx context.Context, userID int64) (*entity.User, error)
	UpdateData(ctx context.Context, c tele.Context) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
}

type clubService interface {
	GetByOwnerID(ctx context.Context, id int64) ([]entity.Club, error)
}

type Handler struct {
	bot         *tele.Bot
	layout      *layout.Layout
	logger      *types.Logger
	userService userService
	clubService clubService
	input       *intele.InputManager
}

func New(b *bot.Bot) *Handler {
	userStorageLocal := postgres.NewUserStorage(b.DB)
	userServiceLocal := service.NewUserService(userStorageLocal, nil, nil, nil, "")
	clubStorage := postgres.NewClubStorage(b.DB)

	return &Handler{
		bot:         b.Bot,
		layout:      b.Layout,
		logger:      b.Logger,
		userService: userServiceLocal,
		clubService: service.NewClubService(clubStorage),
		input:       b.Input,
	}
}

func (h Handler) LoadBanners(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		err := banner.Load(c.Bot())
		if err != nil {
			return err
		}
		return next(c)
	}
}

func (h Handler) Authorized(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		user, err := h.userService.Get(context.Background(), c.Sender().ID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				h.logger.Errorf("(user: %d) error while getting user from db: %v", c.Sender().ID, err)
				return c.Send(
					banner.Auth.Caption(h.layout.Text(c, "technical_issues", err.Error())),
					h.layout.Markup(c, "core:hide"),
				)
			}
			return c.Send(
				banner.Auth.Caption(h.layout.Text(c, "auth_required")),
				h.layout.Markup(c, "core:hide"),
			)
		}

		if c.Sender().Username != user.Username {
			h.logger.Infof("(user: %d) update username", c.Sender().ID)
			user.Username = c.Sender().Username
			_, err = h.userService.Update(context.Background(), user)
			if err != nil {
				return c.Send(
					banner.Auth.Caption(h.layout.Text(c, "technical_issues", err.Error())),
					h.layout.Markup(c, "core:hide"),
				)
			}
		}

		if user.IsBanned {
			return c.Send(
				banner.Auth.Caption(h.layout.TextLocale(user.Localisation, "banned")),
				h.layout.MarkupLocale(user.Localisation, "core:hide"),
			)
		}

		return next(c)
	}
}

func (h Handler) IsClubOwner(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		user, err := h.userService.Get(context.Background(), c.Sender().ID)
		if err != nil {
			h.logger.Errorf("(user: %d) error while getting user from db: %v", c.Sender().ID, err)
			return c.Send(
				banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
				h.layout.Markup(c, "core:hide"),
			)
		}

		clubs, err := h.clubService.GetByOwnerID(context.Background(), user.ID)
		if err != nil {
			h.logger.Errorf("(user: %d) error while getting user's clubs from db: %v", c.Sender().ID, err)
			return c.Send(
				banner.ClubOwner.Caption(h.layout.Text(c, "technical_issues", err.Error())),
				h.layout.Markup(c, "core:hide"),
			)
		}

		if len(clubs) < 1 {
			return c.Send(
				banner.ClubOwner.Caption(h.layout.Text(c, "no_clubs")),
				h.layout.Markup(c, "core:hide"),
			)
		}

		return next(c)
	}
}

//func (h Handler) Localisation(next tele.HandlerFunc) tele.HandlerFunc {
//	return func(c tele.Context) error {
//		user, err := h.userService.Get(context.Background(), c.Sender().ID)
//		if err != nil {
//			return err
//		}
//		if user.Localisation == "" {
//			return c.Send(
//				"<b>🌍 Выберите предпочтительный язык</b>\n"+
//					"<b>🌍 Choose your preferred language</b>\n",
//				h.layout.MarkupLocale("en", "pickLanguage"),
//			)
//		}
//
//		return next(c)
//	}
//}

//func (h Handler) SetupLocalisation(r tele.Recipient) string {
//	userID, err := strconv.Atoi(r.Recipient())
//	if err != nil {
//		return ""
//	}
//
//	user, err := h.userService.Get(context.Background(), int64(userID))
//	if err != nil {
//		return ""
//	}
//	return user.Localisation
//}

// ResetInputOnBack middleware clears the input state when the back button is pressed.
func (h Handler) ResetInputOnBack(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {

		if c.Callback() != nil {
			if strings.Contains(c.Callback().Data, "back") || strings.Contains(c.Callback().Unique, "back") {
				h.input.Cancel(c.Sender().ID)
			}
		}
		if c.Message() != nil {
			if strings.HasPrefix(c.Message().Text, "/") {
				h.input.Cancel(c.Sender().ID)
			}
		}

		return next(c)
	}
}

// PrivateChatOnly middleware ensures the bot only responds to commands in private chats
// and ignores commands in group chats.
func (h Handler) PrivateChatOnly(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		// Check if the chat is private
		if c.Chat().Type != tele.ChatPrivate {
			// If it's not a private chat, silently ignore the command
			h.logger.Infof("Ignoring command in non-private chat (chat_id: %d, chat_type: %s)", c.Chat().ID, c.Chat().Type)
			return nil
		}

		// If it's a private chat, proceed to the next handler
		return next(c)
	}
}

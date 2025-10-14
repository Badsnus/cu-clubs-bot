package banner

import (
	"fmt"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/config"
	"github.com/Badsnus/cu-clubs-bot/bot/pkg/logger"
	tele "gopkg.in/telebot.v3"
)

type Banner tele.File

var (
	Auth            Banner
	Menu            Banner
	PersonalAccount Banner
	Clubs           Banner
	ClubOwner       Banner
	Events          Banner

	loaded bool
)

func Load(b *tele.Bot, cfg config.BannerConfig) error {
	if loaded {
		return nil
	}

	// Validate all banner IDs are configured
	if cfg.AuthID() == "" {
		return fmt.Errorf("auth banner ID is not configured")
	}
	if cfg.MenuID() == "" {
		return fmt.Errorf("menu banner ID is not configured")
	}
	if cfg.PersonalAccountID() == "" {
		return fmt.Errorf("personal account banner ID is not configured")
	}
	if cfg.ClubsID() == "" {
		return fmt.Errorf("clubs banner ID is not configured")
	}
	if cfg.ClubOwnerID() == "" {
		return fmt.Errorf("club owner banner ID is not configured")
	}
	if cfg.EventsID() == "" {
		return fmt.Errorf("events banner ID is not configured")
	}

	// Load Auth banner
	auth, err := b.FileByID(cfg.AuthID())
	if err != nil {
		return fmt.Errorf("failed to load auth banner (ID: %s): %w", cfg.AuthID(), err)
	}
	Auth = Banner(auth)

	// Load Menu banner
	menu, err := b.FileByID(cfg.MenuID())
	if err != nil {
		return fmt.Errorf("failed to load menu banner (ID: %s): %w", cfg.MenuID(), err)
	}
	Menu = Banner(menu)

	// Load PersonalAccount banner
	personalAccount, err := b.FileByID(cfg.PersonalAccountID())
	if err != nil {
		return fmt.Errorf("failed to load personal account banner (ID: %s): %w", cfg.PersonalAccountID(), err)
	}
	PersonalAccount = Banner(personalAccount)

	// Load Clubs banner
	clubs, err := b.FileByID(cfg.ClubsID())
	if err != nil {
		return fmt.Errorf("failed to load clubs banner (ID: %s): %w", cfg.ClubsID(), err)
	}
	Clubs = Banner(clubs)

	// Load ClubOwner banner
	clubOwner, err := b.FileByID(cfg.ClubOwnerID())
	if err != nil {
		return fmt.Errorf("failed to load club owner banner (ID: %s): %w", cfg.ClubOwnerID(), err)
	}
	ClubOwner = Banner(clubOwner)

	// Load Events banner
	events, err := b.FileByID(cfg.EventsID())
	if err != nil {
		return fmt.Errorf("failed to load events banner (ID: %s): %w", cfg.EventsID(), err)
	}
	Events = Banner(events)

	loaded = true
	logger.Log.Info("All banners loaded successfully")
	return nil
}

func (b *Banner) Caption(caption string) interface{} {
	if b == nil {
		return caption
	}
	return &tele.Photo{File: tele.File{
		FileID:     b.FileID,
		UniqueID:   b.UniqueID,
		FileSize:   b.FileSize,
		FilePath:   b.FilePath,
		FileLocal:  b.FileLocal,
		FileURL:    b.FileURL,
		FileReader: b.FileReader,
	}, Caption: caption}
}

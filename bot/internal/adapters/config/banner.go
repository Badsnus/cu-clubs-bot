package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type BannerConfig interface {
	AuthID() string
	MenuID() string
	PersonalAccountID() string
	ClubsID() string
	ClubOwnerID() string
	EventsID() string
	Validate() error
}

type bannerConfig struct {
	authID            string
	menuID            string
	personalAccountID string
	clubsID           string
	clubOwnerID       string
	eventsID          string
}

func NewBannerConfig() BannerConfig {
	return &bannerConfig{
		authID:            viper.GetString("bot.banner.auth"),
		menuID:            viper.GetString("bot.banner.menu"),
		personalAccountID: viper.GetString("bot.banner.personal-account"),
		clubsID:           viper.GetString("bot.banner.clubs"),
		clubOwnerID:       viper.GetString("bot.banner.club-owner"),
		eventsID:          viper.GetString("bot.banner.events"),
	}
}

func (cfg *bannerConfig) AuthID() string {
	return cfg.authID
}

func (cfg *bannerConfig) MenuID() string {
	return cfg.menuID
}

func (cfg *bannerConfig) PersonalAccountID() string {
	return cfg.personalAccountID
}

func (cfg *bannerConfig) ClubsID() string {
	return cfg.clubsID
}

func (cfg *bannerConfig) ClubOwnerID() string {
	return cfg.clubOwnerID
}

func (cfg *bannerConfig) EventsID() string {
	return cfg.eventsID
}

func (cfg *bannerConfig) Validate() error {
	if cfg.authID == "" {
		return fmt.Errorf("auth banner ID is not configured")
	}
	if cfg.menuID == "" {
		return fmt.Errorf("menu banner ID is not configured")
	}
	if cfg.personalAccountID == "" {
		return fmt.Errorf("personal account banner ID is not configured")
	}
	if cfg.clubsID == "" {
		return fmt.Errorf("clubs banner ID is not configured")
	}
	if cfg.clubOwnerID == "" {
		return fmt.Errorf("club owner banner ID is not configured")
	}
	if cfg.eventsID == "" {
		return fmt.Errorf("events banner ID is not configured")
	}
	return nil
}

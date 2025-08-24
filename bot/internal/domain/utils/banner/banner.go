package banner

import (
	"github.com/spf13/viper"
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

func Load(b *tele.Bot) error {
	if loaded {
		return nil
	}
	auth, err := b.FileByID(viper.GetString("bot.banner.auth"))
	if err != nil {
		return err
	}
	Auth = Banner(auth)

	menu, err := b.FileByID(viper.GetString("bot.banner.menu"))
	if err != nil {
		return err
	}
	Menu = Banner(menu)

	personalAccount, err := b.FileByID(viper.GetString("bot.banner.personal-account"))
	if err != nil {
		return err
	}
	PersonalAccount = Banner(personalAccount)

	clubOwner, err := b.FileByID(viper.GetString("bot.banner.club-owner"))
	if err != nil {
		return err
	}
	ClubOwner = Banner(clubOwner)

	events, err := b.FileByID(viper.GetString("bot.banner.events"))
	if err != nil {
		return err
	}
	Events = Banner(events)

	loaded = true
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

package main

import (
	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/config"
	setupBot "github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/setup"
	"log"

	_ "github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/database/postgres/ent/runtime"
	_ "time/tzdata"
)

func main() {
	cfg := config.Get()
	b, err := bot.New(cfg)
	if err != nil {
		log.Panic(err)
	}

	setupBot.Setup(b)

	b.Start()
}

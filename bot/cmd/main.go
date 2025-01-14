package main

import (
	"github.com/Badsnus/cu-clubs-bot/cmd/bot"
	setupBot "github.com/Badsnus/cu-clubs-bot/internal/adapters/controller/telegram/setup"

	"github.com/Badsnus/cu-clubs-bot/internal/adapters/config"
	"log"
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
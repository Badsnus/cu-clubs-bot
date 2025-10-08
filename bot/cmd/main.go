package main

import (
	"log"

	"github.com/Badsnus/cu-clubs-bot/bot/cmd/bot"
	"github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/config"
	setupBot "github.com/Badsnus/cu-clubs-bot/bot/internal/adapters/controller/telegram/setup"

	_ "time/tzdata"
)

func main() {
	cfg, err := config.Get()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	b, err := bot.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	setupBot.Setup(b)

	b.Start()
}

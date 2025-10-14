package main

import (
	"context"
	"log"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/app"

	_ "time/tzdata"
)

func main() {
	ctx := context.Background()

	a, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	a.Run()
}

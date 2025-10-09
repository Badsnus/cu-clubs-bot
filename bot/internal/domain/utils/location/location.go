package location

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

var location *time.Location

func Location() *time.Location {
	if location != nil {
		return location
	}
	var err error
	location, err = time.LoadLocation(viper.GetString("settings.timezone"))
	if err != nil {
		log.Fatalf("error while load time location: %v", err)
	}
	return location
}

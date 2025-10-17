package location

import (
	"log"
	"time"
)

var location *time.Location

func Init(timezone string) {
	if location != nil {
		return
	}
	var err error
	location, err = time.LoadLocation(timezone)
	if err != nil {
		log.Fatalf("error while load time location: %v", err)
	}
}

func Location() *time.Location {
	if location == nil {
		log.Fatal("location not initialized")
	}
	return location
}

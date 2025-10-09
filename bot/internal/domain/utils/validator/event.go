package validator

import (
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/Badsnus/cu-clubs-bot/bot/internal/domain/utils/location"
)

func EventName(name string, _ map[string]interface{}) bool {
	return utf8.RuneCountInString(name) >= 5 && utf8.RuneCountInString(name) <= 45
}

func EventDescription(description string, _ map[string]interface{}) bool {
	return utf8.RuneCountInString(description) <= 250
}

func EventLocation(location string, _ map[string]interface{}) bool {
	return utf8.RuneCountInString(location) >= 5 && utf8.RuneCountInString(location) <= 75
}

func EventStartTime(start string, _ map[string]interface{}) bool {
	const layout = "02.01.2006 15:04"

	startTime, err := time.ParseInLocation(layout, start, location.Location())
	if err != nil {
		return false
	}

	currentTime := time.Now()

	moscowTime := currentTime.In(location.Location())

	tomorrow := moscowTime.Add(time.Hour * time.Duration(24))

	return !startTime.Before(tomorrow)
}

func EventEndTime(end string, params map[string]interface{}) bool {
	const layout = "02.01.2006 15:04"

	startTimeStr, ok := params["startTime"].(string)
	if !ok {
		return false
	}
	startTime, _ := time.ParseInLocation(layout, startTimeStr, location.Location())

	endTime, err := time.ParseInLocation(layout, end, location.Location())
	if err != nil {
		return false
	}

	if !endTime.In(location.Location()).After(startTime) {
		return false
	}

	return true
}

func EventRegisteredEndTime(registeredEnd string, params map[string]interface{}) bool {
	const layout = "02.01.2006 15:04"

	startTimeStr, ok := params["startTime"].(string)
	if !ok {
		return false
	}
	startTime, _ := time.ParseInLocation(layout, startTimeStr, location.Location())
	registeredEndTime, err := time.ParseInLocation(layout, registeredEnd, location.Location())
	if err != nil {
		return false
	}

	if registeredEndTime.After(startTime) {
		return false
	}

	// Check if registration end time is at least 1 hour later than current time
	now := time.Now().In(location.Location())
	return registeredEndTime.After(now.Add(time.Hour))
}

func EventAfterRegistrationText(afterRegistrationText string, _ map[string]interface{}) bool {
	return utf8.RuneCountInString(afterRegistrationText) >= 10 && utf8.RuneCountInString(afterRegistrationText) <= 150
}

func EventMaxParticipants(maxParticipantsStr string, _ map[string]interface{}) bool {
	maxParticipants, err := strconv.Atoi(maxParticipantsStr)
	if err != nil {
		return false
	}
	return maxParticipants >= 0
}

func EventExpectedParticipants(expectedParticipants string, _ map[string]interface{}) bool {
	expected, err := strconv.Atoi(expectedParticipants)
	if err != nil {
		return false
	}
	return expected >= 0
}

func EventEditMaxParticipants(maxParticipantsStr string, params map[string]interface{}) bool {
	previousMaxParticipants, ok := params["previousMaxParticipants"].(int)
	if !ok {
		return false
	}
	maxParticipants, err := strconv.Atoi(maxParticipantsStr)
	if err != nil {
		return false
	}
	if maxParticipants == 0 {
		return true
	}
	return maxParticipants > 0 && maxParticipants > previousMaxParticipants
}

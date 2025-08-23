package validator

import (
	"net/url"
	"unicode/utf8"
)

func ClubName(name string, _ map[string]interface{}) bool {
	return utf8.RuneCountInString(name) >= 3 && utf8.RuneCountInString(name) <= 30
}

func ClubDescription(description string, _ map[string]interface{}) bool {
	return utf8.RuneCountInString(description) <= 400
}

func ClubLink(link string, _ map[string]interface{}) bool {
	if _, err := url.ParseRequestURI(link); err != nil {
		return false
	}

	return utf8.RuneCountInString(link) <= 100
}

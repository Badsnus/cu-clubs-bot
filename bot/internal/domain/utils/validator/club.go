package validator

import (
	"unicode/utf8"
)

func ClubName(name string) bool {
	return utf8.RuneCountInString(name) >= 3 && utf8.RuneCountInString(name) <= 30
}
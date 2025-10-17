package validator

import (
	"net/mail"
	"regexp"
	"strings"
)

func Fio(fio string, _ map[string]interface{}) bool {
	if splitFio := strings.Split(fio, " "); len(splitFio) != 3 {
		return false
	}
	re := regexp.MustCompile(`^[А-ЯЁ][а-яё]+(?:-[А-ЯЁ][а-яё]+)? [А-ЯЁ][а-яё]+ [А-ЯЁ][а-яё]+$`)
	return re.MatchString(strings.TrimSpace(fio))
}

func Email(email string, validDomains []string) bool {
	return emailFormat(email) && emailDomain(email, validDomains)
}

func emailFormat(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func emailDomain(email string, validDomains []string) bool {
	for _, domain := range validDomains {
		if strings.HasSuffix(email, domain) {
			return true
		}
	}
	return false
}

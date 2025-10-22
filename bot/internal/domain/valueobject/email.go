package valueobject

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
)

// Email represents an email address value object
type Email string

// NewEmail creates a new Email value object
func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" {
		return "", fmt.Errorf("email cannot be empty")
	}

	if !isValidEmail(email) {
		return "", fmt.Errorf("invalid email format")
	}

	return Email(email), nil
}

// String returns the email address
func (e Email) String() string {
	return string(e)
}

// IsValid checks if email is valid
func (e Email) IsValid() bool {
	return isValidEmail(e.String())
}

// isValidEmail validates email format using regex
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// Value implements driver.Valuer interface for GORM
func (e Email) Value() (driver.Value, error) {
	if e.String() == "" {
		return "", nil
	}
	return e.String(), nil
}

// Scan implements sql.Scanner interface for GORM
func (e *Email) Scan(value interface{}) error {
	if value == nil {
		*e = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*e = Email(v)
		if !e.IsValid() {
			return fmt.Errorf("invalid email value: %s", v)
		}
		return nil
	case []byte:
		*e = Email(v)
		if !e.IsValid() {
			return fmt.Errorf("invalid email value: %s", v)
		}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into Email", value)
	}
}

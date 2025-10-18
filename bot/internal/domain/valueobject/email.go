package valueobject

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Email represents an email address value object
type Email struct {
	value string
}

// NewEmail creates a new Email value object
func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" {
		return Email{}, fmt.Errorf("email cannot be empty")
	}

	if !isValidEmail(email) {
		return Email{}, fmt.Errorf("invalid email format")
	}

	return Email{value: email}, nil
}

// String returns the email address
func (e Email) String() string {
	return e.value
}

// IsValid checks if email is valid
func (e Email) IsValid() bool {
	return isValidEmail(e.value)
}

// isValidEmail validates email format using regex
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// Value implements driver.Valuer interface for GORM
func (e Email) Value() (driver.Value, error) {
	if e.value == "" {
		return "", nil
	}
	return e.value, nil
}

// Scan implements sql.Scanner interface for GORM
func (e *Email) Scan(value interface{}) error {
	if value == nil {
		e.value = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		// Handle empty strings gracefully - allow empty emails when scanning from database
		if v == "" {
			e.value = ""
			return nil
		}
		email, err := NewEmail(v)
		if err != nil {
			return err
		}
		e.value = email.value
		return nil
	case []byte:
		str := string(v)
		// Handle empty byte slices gracefully - allow empty emails when scanning from database
		if str == "" {
			e.value = ""
			return nil
		}
		email, err := NewEmail(str)
		if err != nil {
			return err
		}
		e.value = email.value
		return nil
	default:
		return fmt.Errorf("cannot scan %T into Email", value)
	}
}

// MarshalJSON implements json.Marshaler interface for Redis serialization
func (e Email) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.value)
}

// UnmarshalJSON implements json.Unmarshaler interface for Redis deserialization
func (e *Email) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	// Handle empty string case
	if value == "" {
		e.value = ""
		return nil
	}

	email, err := NewEmail(value)
	if err != nil {
		return err
	}
	e.value = email.value
	return nil
}

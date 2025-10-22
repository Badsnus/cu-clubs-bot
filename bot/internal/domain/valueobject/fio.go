package valueobject

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// FIO represents person's full name (Surname, Name, Patronymic)
type FIO struct {
	Surname    string
	Name       string
	Patronymic string
}

// NewFIO creates a new FIO value object
func NewFIO(surname, name, patronymic string) (FIO, error) {
	fio := FIO{
		Surname:    strings.TrimSpace(surname),
		Name:       strings.TrimSpace(name),
		Patronymic: strings.TrimSpace(patronymic),
	}

	if err := fio.validate(); err != nil {
		return FIO{}, err
	}

	return fio, nil
}

// NewFIOFromString creates FIO from a single string
func NewFIOFromString(fioStr string) (FIO, error) {
	parts := strings.Split(fioStr, " ")
	if len(parts) < 2 {
		return FIO{}, fmt.Errorf("invalid FIO format: at least surname and name required")
	}

	surname := parts[0]
	name := parts[1]
	patronymic := ""

	if len(parts) > 2 {
		patronymic = strings.Join(parts[2:], " ")
	}

	return NewFIO(surname, name, patronymic)
}

// String returns formatted FIO string
func (f FIO) String() string {
	if f.Patronymic != "" {
		return fmt.Sprintf("%s %s %s", f.Surname, f.Name, f.Patronymic)
	}
	return fmt.Sprintf("%s %s", f.Surname, f.Name)
}

// ShortName returns short version (Surname and Name)
func (f FIO) ShortName() string {
	return fmt.Sprintf("%s %s", f.Surname, f.Name)
}

// IsValid checks if FIO is valid
func (f FIO) IsValid() bool {
	return f.validate() == nil
}

// validate checks FIO validity
func (f FIO) validate() error {
	if strings.TrimSpace(f.Surname) == "" {
		return fmt.Errorf("surname cannot be empty")
	}
	if strings.TrimSpace(f.Name) == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return nil
}

// Value implements driver.Valuer interface for Ent.
func (f FIO) Value() (driver.Value, error) {
	return f.String(), nil
}

// Scan implements sql.Scanner interface for Ent.
func (f *FIO) Scan(value interface{}) error {
	if value == nil {
		*f = FIO{}
		return nil
	}

	var fioStr string
	switch v := value.(type) {
	case string:
		fioStr = v
	case []byte:
		fioStr = string(v)
	default:
		return fmt.Errorf("cannot scan %T into FIO", value)
	}

	if fioStr == "" {
		*f = FIO{}
		return nil
	}

	fio, err := NewFIOFromString(fioStr)
	if err != nil {
		return err
	}
	*f = fio
	return nil
}

// MarshalJSON implements json.Marshaler interface for Redis serialization
func (f FIO) MarshalJSON() ([]byte, error) {
	type Alias FIO
	return json.Marshal(&struct {
		Alias
		StringValue string `json:"string_value"`
	}{
		Alias:       (Alias)(f),
		StringValue: f.String(),
	})
}

// UnmarshalJSON implements json.Unmarshaler interface for Redis deserialization
func (f *FIO) UnmarshalJSON(data []byte) error {
	// First try to unmarshal as a simple string (for backward compatibility)
	var str string
	if err := json.Unmarshal(data, &str); err == nil && str != "" {
		fio, err := NewFIOFromString(str)
		if err != nil {
			return err
		}
		*f = fio
		return nil
	}

	// Otherwise, unmarshal as struct with fields
	type Alias FIO
	aux := &struct {
		Alias
		StringValue string `json:"string_value"`
	}{
		Alias: (Alias)(*f),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	*f = FIO(aux.Alias)
	return nil
}

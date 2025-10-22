package valueobject

import (
	"database/sql/driver"
	"fmt"
)

// Role represents a user role in the system
type Role string

func (r Role) String() string {
	return string(r)
}

func (r Role) IsValid() bool {
	switch r {
	case ExternalUser, GrantUser, Student:
		return true
	default:
		return false
	}
}

// Value implements driver.Valuer interface for Ent.
func (r Role) Value() (driver.Value, error) {
	return r.String(), nil
}

// Scan implements sql.Scanner interface for Ent.
func (r *Role) Scan(value interface{}) error {
	if value == nil {
		*r = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*r = Role(v)
		if !r.IsValid() {
			return fmt.Errorf("invalid role value: %s", v)
		}
		return nil
	case []byte:
		*r = Role(v)
		if !r.IsValid() {
			return fmt.Errorf("invalid role value: %s", v)
		}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into Role", value)
	}
}

const (
	ExternalUser Role = "external_user"
	GrantUser    Role = "grant_user"
	Student      Role = "student"
)

// Roles represents a collection of roles
type Roles []Role

func (r Roles) Contains(role Role) bool {
	for _, userRole := range r {
		if userRole == role {
			return true
		}
	}
	return false
}

func (r Roles) Strings() []string {
	result := make([]string, len(r))
	for i, role := range r {
		result[i] = role.String()
	}
	return result
}

// AllRoles returns all available roles in the system
func AllRoles() Roles {
	return Roles{ExternalUser, GrantUser, Student}
}

// Values provides list valid values for Enum.
func (Role) Values() (kinds []string) {
	for _, s := range AllRoles() {
		kinds = append(kinds, s.String())
	}
	return kinds
}

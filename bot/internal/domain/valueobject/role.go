package valueobject

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

package valueobject

import (
	"database/sql/driver"
	"fmt"
)

type PassType string

const (
	PassTypeEvent  PassType = "event"
	PassTypeManual PassType = "manual"
	PassTypeAPI    PassType = "api"
)

func (pt PassType) String() string {
	return string(pt)
}

func (pt PassType) IsValid() bool {
	switch pt {
	case PassTypeEvent, PassTypeManual, PassTypeAPI:
		return true
	default:
		return false
	}
}

func (pt *PassType) Scan(value interface{}) error {
	if value == nil {
		*pt = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*pt = PassType(v)
		if !pt.IsValid() {
			return fmt.Errorf("invalid pass type value: %s", v)
		}
		return nil
	case []byte:
		*pt = PassType(v)
		if !pt.IsValid() {
			return fmt.Errorf("invalid pass type value: %s", v)
		}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into PassType", value)
	}
}

func (pt PassType) Value() (driver.Value, error) {
	return pt.String(), nil
}

func (pt PassType) Values() (kinds []string) {
	for _, s := range []PassType{
		PassTypeEvent,
		PassTypeManual,
		PassTypeAPI,
	} {
		kinds = append(kinds, s.String())
	}
	return kinds
}

type PassStatus string

const (
	PassStatusPending   PassStatus = "pending"
	PassStatusSent      PassStatus = "sent"
	PassStatusCancelled PassStatus = "cancelled"
)

func (ps PassStatus) String() string {
	return string(ps)
}

func (ps PassStatus) IsValid() bool {
	switch ps {
	case PassStatusPending, PassStatusSent, PassStatusCancelled:
		return true
	default:
		return false
	}
}

func (ps *PassStatus) Scan(value interface{}) error {
	if value == nil {
		*ps = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*ps = PassStatus(v)
		if !ps.IsValid() {
			return fmt.Errorf("invalid pass status value: %s", v)
		}
		return nil
	case []byte:
		*ps = PassStatus(v)
		if !ps.IsValid() {
			return fmt.Errorf("invalid pass status value: %s", v)
		}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into PassStatus", value)
	}
}

func (ps PassStatus) Value() (driver.Value, error) {
	return ps.String(), nil
}

func (ps PassStatus) Values() (kinds []string) {
	for _, s := range []PassStatus{
		PassStatusPending,
		PassStatusSent,
		PassStatusCancelled,
	} {
		kinds = append(kinds, s.String())
	}
	return kinds
}

type RequesterType string

const (
	RequesterTypeUser  RequesterType = "user"
	RequesterTypeAdmin RequesterType = "admin"
	RequesterTypeClub  RequesterType = "club"
)

func (rt RequesterType) String() string {
	return string(rt)
}

func (rt RequesterType) IsValid() bool {
	switch rt {
	case RequesterTypeUser, RequesterTypeAdmin, RequesterTypeClub:
		return true
	default:
		return false
	}
}

func (rt *RequesterType) Scan(value interface{}) error {
	if value == nil {
		*rt = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*rt = RequesterType(v)
		if !rt.IsValid() {
			return fmt.Errorf("invalid requester type value: %s", v)
		}
		return nil
	case []byte:
		*rt = RequesterType(v)
		if !rt.IsValid() {
			return fmt.Errorf("invalid requester type value: %s", v)
		}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into RequesterType", value)
	}
}

func (rt RequesterType) Value() (driver.Value, error) {
	return rt.String(), nil
}

func (rt RequesterType) Values() (kinds []string) {
	for _, s := range []RequesterType{
		RequesterTypeUser,
		RequesterTypeAdmin,
		RequesterTypeClub,
	} {
		kinds = append(kinds, s.String())
	}
	return kinds
}

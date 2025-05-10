package entity

import (
	"time"

	"github.com/google/uuid"
)

const (
	UserStatusBlocked = "blocked"
	UserStatusActive  = "active"
)

type (
	User struct {
		ID         uuid.UUID
		Status     UserStatus
		Phone      string
		CreateTime time.Time
		UpdateTime time.Time
	}

	ParamsCreateUser struct {
		Status UserStatus
		Phone  string
	}

	UserStatus string
)

func (u *User) IsBlocked() bool {
	return u.Status == UserStatusBlocked
}

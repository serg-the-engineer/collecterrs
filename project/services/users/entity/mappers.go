package entity

import (
	"your-company.com/project/services/users/storage/db"
)

func MakeDBUserToEntity(user db.User) (*User, error) {
	status := UserStatusActive
	if !user.Status {
		status = UserStatusBlocked
	}
	result := &User{
		Status: UserStatus(status),
		Phone:  user.Phone,
	}
	return result, nil
}

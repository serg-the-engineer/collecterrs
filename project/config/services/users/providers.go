package users

import (
	"your-company.com/project/pkg/redis"
	"your-company.com/project/services/users/storage"
	"your-company.com/project/specs/proto/otp"
)

type Providers struct {
	Redis   *redis.Client
	Storage storage.Storage
	Otp     otp.OtpClient
}

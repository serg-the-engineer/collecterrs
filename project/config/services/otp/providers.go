package otp

import (
	o "your-company.com/project/pkg/otp"
	"your-company.com/project/pkg/redis"
)

type Providers struct {
	ProviderOtp o.ProviderOtp
	Redis       *redis.Client
}

package usecase

import (
	"context"
	"sync"
	"your-company.com/project/pkg/redis"
	"your-company.com/project/services/dummy/storage"
	"your-company.com/project/specs/proto/otp"
)

var _ Dummy = (*dummyImpl)(nil)

type Dummy interface {
	Dummy1(ctx context.Context)
}

type ServiceLocatorImpl struct {
	Storage storage.Storage
	Otp     otp.OtpClient
	Redis   *redis.Client

	mu sync.RWMutex // Мьютекс для синхронизации доступа к структуре
}

type dummyImpl struct {
	Dummy
	Providers ServiceLocatorImpl
}

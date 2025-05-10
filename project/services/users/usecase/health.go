package usecase

import (
	"context"

	"your-company.com/project/services/users/entity"
)

func (u *useCasesImpl) HealthCheck(ctx context.Context) (*entity.Health, error) {
	err := u.Providers.Storage.Check(ctx)
	if err != nil {
		return nil, err
	}
	return entity.MakeHealth(true), nil
}

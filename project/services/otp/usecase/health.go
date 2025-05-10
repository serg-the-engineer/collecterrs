package usecase

import (
	"context"

	"your-company.com/project/services/otp/entity"
)

func (u *useCasesImpl) HealthCheck(ctx context.Context) (*entity.Health, error) {
	return entity.MakeHealth(true), nil
}

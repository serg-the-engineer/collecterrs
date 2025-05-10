package usecase

import (
	"context"

	"your-company.com/project/config/services/otp"

	"your-company.com/project/services/otp/entity"
)

var _ Otp = (*useCasesImpl)(nil)

type Otp interface {
	HealthCheck(ctx context.Context) (*entity.Health, error)
	GenerateCode(ctx context.Context, req *entity.GenerateCodeReq) (*entity.GenerateCode, error)
	GenerateRetryCode(ctx context.Context, req *entity.GenerateRetryCodeReq) (*entity.GenerateRetryCode, error)
	ValidateCode(ctx context.Context, req *entity.ValidateCodeReq) (*entity.ValidateCode, error)
}

type useCasesImpl struct {
	Otp
	cfg       *otp.Config
	Providers *otp.Providers
}

func New(cfg *otp.Config, providers *otp.Providers) Otp {
	useCases := &useCasesImpl{
		cfg:       cfg,
		Providers: providers,
	}

	useCases.Otp = useCases
	return useCases
}

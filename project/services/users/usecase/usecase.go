package usecase

import (
	"context"

	"your-company.com/project/config/services/users"
	"your-company.com/project/services/users/entity"
)

var _ Users = (*useCasesImpl)(nil)

type Users interface {
	HealthCheck(ctx context.Context) (*entity.Health, error)
	Login(ctx context.Context, req *entity.LoginReq) (*entity.Login, error)
	ConfirmLogin(ctx context.Context, req *entity.ConfirmLoginReq) (*entity.ConfirmLogin, error)
}

type useCasesImpl struct {
	Users
	cfg       *users.Config
	Providers *users.Providers
}

func New(cfg *users.Config, providers *users.Providers) Users {
	useCases := &useCasesImpl{
		cfg:       cfg,
		Providers: providers,
	}

	useCases.Users = useCases

	return useCases
}

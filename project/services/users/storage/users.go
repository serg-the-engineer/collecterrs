package storage

import (
	"context"
	"your-company.com/project/errs/errsUsers"

	"your-company.com/project/services/users/entity"
)

var _ User = (*storageImpl)(nil)

type User interface {
	GetUser(ctx context.Context, phone string) (*entity.User, error)
	CreateUser(ctx context.Context, user entity.ParamsCreateUser) (*entity.User, error)
}

func (s *storageImpl) GetUser(ctx context.Context, phone string) (*entity.User, error) {
	user, err := s.DBClient.GetUser(ctx, phone)
	if err != nil {
		if err.Error() == "not found" {
			return nil, errsUsers.UserNotFoundError
		}
		return nil, err
	}

	result, err := entity.MakeDBUserToEntity(*user)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *storageImpl) CreateUser(ctx context.Context, user entity.ParamsCreateUser) (*entity.User, error) {
	status := 1
	if user.Status == entity.UserStatusBlocked {
		status = 0
	}
	dbUser, err := s.DBClient.CreateUser(ctx, user.Phone, status)
	if err != nil {
		return nil, err
	}

	result, err := entity.MakeDBUserToEntity(*dbUser)
	if err != nil {
		return nil, err
	}

	return result, nil
}

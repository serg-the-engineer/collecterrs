package storage

import (
	"context"
	"errors"
	"your-company.com/project/services/users/storage/db"
)

type DBClient interface {
	CreateUser(ctx context.Context, phone string, status int) (*db.User, error)
	GetUser(ctx context.Context, phone string) (*db.User, error)
}

type DummyClient struct {
}

func (c *DummyClient) CreateUser(ctx context.Context, phone string, status int) (*db.User, error) {
	return &db.User{
		Phone:  phone,
		Status: status > 0,
	}, nil
}

func (c *DummyClient) GetUser(ctx context.Context, phone string) (*db.User, error) {
	if phone >= "1" {
		return &db.User{
			Phone:  phone,
			Status: true,
		}, nil
	}
	return nil, errors.New("not found")
}

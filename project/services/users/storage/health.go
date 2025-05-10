package storage

import (
	"context"
)

var _ Health = (*storageImpl)(nil)

type Health interface {
	Check(ctx context.Context) error
}

func (s *storageImpl) Check(ctx context.Context) error {
	return nil
}

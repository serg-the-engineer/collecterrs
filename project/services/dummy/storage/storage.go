package storage

import "your-company.com/project/errs/errsDummy"

type Storage interface {
	Dummy
}

var _ Dummy = (*storageImpl)(nil)

type storageImpl struct {
	Storage
}

type Dummy interface {
	GetDummy(val string) (string, error)
}

func (s *storageImpl) GetDummy(val string) (string, error) {
	if val == "1" {
		return "1", errsDummy.FromStorageHandledError
	}
	_, err := s.nestedDummy()
	if err != nil {
		return "", err
	}
	return "dummy", nil
}

func (s *storageImpl) nestedDummy() (string, error) {
	return "0", errsDummy.FromStorageUnhandledError
}

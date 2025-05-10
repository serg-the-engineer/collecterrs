package storage

var _ Storage = (*storageImpl)(nil)

type Storage interface {
	Health
	User
}

type (
	storageImpl struct {
		Storage
		DBClient
	}
)

func New() Storage {
	storage := &storageImpl{
		DBClient: &DummyClient{},
	}

	storage.Storage = storage

	return storage
}

package storage

func NewPostgresStorage() Storage {
	return NewMemStorage()
}

package storage

import "github.com/ian-kent/Go-MailHog/data"

// Storage represents a storage backend
type Storage interface {
	Store(m *data.Message) (string, error)
	List(start int, limit int) (*data.Messages, error)
	DeleteOne(id string) error
	DeleteAll() error
	Load(id string) (*data.Message, error)
}

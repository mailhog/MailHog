package storage

import (
	"io"
	"time"

	"github.com/mailhog/data"
)

// Storage represents a storage backend
type Storage interface {
	Store(m *data.Message) (string, error)
	List(start, limit int) (*data.Messages, error)
	Search(kind, query string, start, limit int) (*data.Messages, int, error)
	Count() int
	DeleteOne(id string) error
	DeleteAll() error
	DeleteOlderThan(cutoff time.Time) error
	Load(id string) (*data.Message, error)
}

// MessageWriter stores a message body incrementally. It is used by storage
// backends which can avoid keeping full SMTP DATA in memory.
type MessageWriter interface {
	WriteLine(line string) error
	Commit() (string, *data.Message, error)
	Abort() error
}

// StreamingStorage can persist SMTP DATA as it arrives.
type StreamingStorage interface {
	CreateMessageWriter(m *data.SMTPMessage, hostname string) (MessageWriter, error)
}

// PreviewStorage can return a memory-safe message representation for UI
// preview endpoints.
type PreviewStorage interface {
	LoadPreview(id string, maxSize int64) (*data.Message, error)
}

// DownloadStorage can stream an RFC822 message without loading it into memory.
type DownloadStorage interface {
	WriteMessageTo(id string, w io.Writer) error
}

// MessageBodyChunk is a bounded preview chunk of the displayable message body.
type MessageBodyChunk struct {
	ID         string
	Content    *data.Content
	Offset     int64
	NextOffset int64
	Limit      int64
	MaxSize    int64
	HasMore    bool
	Truncated  bool
	Source     string
}

// BodyPreviewStorage can return bounded display body chunks without loading
// the whole message or its attachments.
type BodyPreviewStorage interface {
	LoadBodyChunk(id string, offset int64, limit int64, maxSize int64) (*MessageBodyChunk, error)
}

package storage

import "github.com/ian-kent/Go-MailHog/data"

// InMemory is an in memory storage backend
type InMemory struct {
	Messages      map[string]*data.Message
	MessageIndex  []string
	MessageRIndex map[string]int
}

// CreateInMemory creates a new in memory storage backend
func CreateInMemory() *InMemory {
	return &InMemory{
		Messages:      make(map[string]*data.Message, 0),
		MessageIndex:  make([]string, 0),
		MessageRIndex: make(map[string]int, 0),
	}
}

// Store stores a message and returns its storage ID
func (memory *InMemory) Store(m *data.Message) (string, error) {
	memory.Messages[string(m.ID)] = m
	memory.MessageIndex = append(memory.MessageIndex, string(m.ID))
	memory.MessageRIndex[string(m.ID)] = len(memory.MessageIndex) - 1
	return string(m.ID), nil
}

// List lists stored messages by index
func (memory *InMemory) List(start int, limit int) (*data.Messages, error) {
	if limit > len(memory.MessageIndex) {
		limit = len(memory.MessageIndex)
	}
	var messages []data.Message
	for _, m := range memory.MessageIndex[start:limit] {
		messages = append(messages, *memory.Messages[m])
	}
	msgs := data.Messages(messages)
	return &msgs, nil
}

// DeleteOne deletes an individual message by storage ID
func (memory *InMemory) DeleteOne(id string) error {
	index := memory.MessageRIndex[string(id)]
	delete(memory.Messages, string(id))
	memory.MessageIndex = append(memory.MessageIndex[:index], memory.MessageIndex[index+1:]...)
	delete(memory.MessageRIndex, string(id))
	return nil
}

// DeleteAll deletes all in memory messages
func (memory *InMemory) DeleteAll() error {
	memory.Messages = make(map[string]*data.Message, 0)
	memory.MessageIndex = make([]string, 0)
	memory.MessageRIndex = make(map[string]int, 0)
	return nil
}

// Load returns an individual message by storage ID
func (memory *InMemory) Load(id string) (*data.Message, error) {
	return memory.Messages[string(id)], nil
}

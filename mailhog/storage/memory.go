package storage

import (
	"github.com/ian-kent/Go-MailHog/mailhog/config"
    "github.com/ian-kent/Go-MailHog/mailhog/data"
)

type Memory struct {
	Config *config.Config
	Messages map[string]*data.Message
	MessageIndex []string
	MessageRIndex map[string]int
}

func CreateMemory(c *config.Config) *Memory {
	return &Memory{
		Config: c,
		Messages: make(map[string]*data.Message, 0),
		MessageIndex: make([]string, 0),
		MessageRIndex: make(map[string]int, 0),
	}
}

func (memory *Memory) Store(m *data.Message) (string, error) {
	memory.Messages[m.Id] = m
	memory.MessageIndex = append(memory.MessageIndex, m.Id)
	memory.MessageRIndex[m.Id] = len(memory.MessageIndex)
	return m.Id, nil
}

func (memory *Memory) List(start int, limit int) ([]*data.Message, error) {
	if limit > len(memory.MessageIndex) { limit = len(memory.MessageIndex) }
	messages := make([]*data.Message, 0)
	for _, m := range memory.MessageIndex[start:limit] {
		messages = append(messages, memory.Messages[m])
	}
	return messages, nil;
}

func (memory *Memory) DeleteOne(id string) error {
	index := memory.MessageRIndex[id];
	delete(memory.Messages, id)
	memory.MessageIndex = append(memory.MessageIndex[:index], memory.MessageIndex[index+1:]...)
	delete(memory.MessageRIndex, id)
	return nil
}

func (memory *Memory) DeleteAll() error {
	memory.Messages = make(map[string]*data.Message, 0)
	memory.MessageIndex = make([]string, 0)
	memory.MessageRIndex = make(map[string]int, 0)
	return nil
}

func (memory *Memory) Load(id string) (*data.Message, error) {
	return memory.Messages[id], nil;
}

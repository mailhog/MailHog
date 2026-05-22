package api

import (
	"github.com/mailhog/data"
	"github.com/mailhog/storage"
)

func loadFullMessage(storageBackend storage.Storage, message data.Message) data.Message {
	if !needsFullMessage(message) {
		return message
	}
	full, err := storageBackend.Load(string(message.ID))
	if err != nil || full == nil {
		return message
	}
	return *full
}

func loadFullMessages(storageBackend storage.Storage, messages data.Messages) []data.Message {
	items := make([]data.Message, 0, len(messages))
	for _, message := range messages {
		items = append(items, loadFullMessage(storageBackend, message))
	}
	return items
}

func needsFullMessage(message data.Message) bool {
	if message.Raw == nil {
		return true
	}
	if message.Content == nil {
		return true
	}
	if message.Content.Body == "" && message.Content.Size > 0 {
		return true
	}
	return false
}

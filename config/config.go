package config

import (
	"github.com/ian-kent/Go-MailHog/data"
	"github.com/ian-kent/Go-MailHog/storage"
)

func DefaultConfig() *Config {
	return &Config{
		SMTPBindAddr: "0.0.0.0:1025",
		HTTPBindAddr: "0.0.0.0:8025",
		Hostname:     "mailhog.example",
		MongoUri:     "127.0.0.1:27017",
		MongoDb:      "mailhog",
		MongoColl:    "messages",
	}
}

type Config struct {
	SMTPBindAddr string
	HTTPBindAddr string
	Hostname     string
	MongoUri     string
	MongoDb      string
	MongoColl    string
	MessageChan  chan *data.Message
	Storage      storage.Storage
	Assets       func(asset string) ([]byte, error)
}

package config

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
	MessageChan  chan interface{}
	Storage      interface{}
	Assets       func(asset string) ([]byte, error)
}

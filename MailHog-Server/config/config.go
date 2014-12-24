package config

import (
	"flag"
	"log"

	"github.com/ian-kent/Go-MailHog/MailHog-Server/monkey"
	"github.com/ian-kent/Go-MailHog/data"
	"github.com/ian-kent/Go-MailHog/storage"
	"github.com/ian-kent/envconf"
)

func DefaultConfig() *Config {
	return &Config{
		SMTPBindAddr: "0.0.0.0:1025",
		HTTPBindAddr: "0.0.0.0:8025",
		Hostname:     "mailhog.example",
		MongoUri:     "127.0.0.1:27017",
		MongoDb:      "mailhog",
		MongoColl:    "messages",
		StorageType:  "memory",
		MessageChan:  make(chan *data.Message),
	}
}

type Config struct {
	SMTPBindAddr string
	HTTPBindAddr string
	Hostname     string
	MongoUri     string
	MongoDb      string
	MongoColl    string
	StorageType  string
	InviteJim    bool
	Storage      storage.Storage
	MessageChan  chan *data.Message
	Assets       func(asset string) ([]byte, error)
	Monkey       monkey.ChaosMonkey
}

var cfg = DefaultConfig()
var jim = &monkey.Jim{}

func Configure() *Config {
	switch cfg.StorageType {
	case "memory":
		log.Println("Using in-memory storage")
		cfg.Storage = storage.CreateInMemory()
	case "mongodb":
		log.Println("Using MongoDB message storage")
		s := storage.CreateMongoDB(cfg.MongoUri, cfg.MongoDb, cfg.MongoColl)
		if s == nil {
			log.Println("MongoDB storage unavailable, reverting to in-memory storage")
			cfg.Storage = storage.CreateInMemory()
		} else {
			log.Println("Connected to MongoDB")
			cfg.Storage = s
		}
	default:
		log.Fatalf("Invalid storage type %s", cfg.StorageType)
	}

	if cfg.InviteJim {
		jim.Configure(func(message string, args ...interface{}) {
			log.Printf(message, args...)
		})
		cfg.Monkey = jim
	}

	return cfg
}

func RegisterFlags() {
	flag.StringVar(&cfg.SMTPBindAddr, "smtpbindaddr", envconf.FromEnvP("MH_SMTP_BIND_ADDR", "0.0.0.0:1025").(string), "SMTP bind interface and port, e.g. 0.0.0.0:1025 or just :1025")
	flag.StringVar(&cfg.HTTPBindAddr, "httpbindaddr", envconf.FromEnvP("MH_HTTP_BIND_ADDR", "0.0.0.0:8025").(string), "HTTP bind interface and port, e.g. 0.0.0.0:8025 or just :8025")
	flag.StringVar(&cfg.Hostname, "hostname", envconf.FromEnvP("MH_HOSTNAME", "mailhog.example").(string), "Hostname for EHLO/HELO response, e.g. mailhog.example")
	flag.StringVar(&cfg.StorageType, "storage", envconf.FromEnvP("MH_STORAGE", "memory").(string), "Message storage: memory (default) or mongodb")
	flag.StringVar(&cfg.MongoUri, "mongouri", envconf.FromEnvP("MH_MONGO_URI", "127.0.0.1:27017").(string), "MongoDB URI, e.g. 127.0.0.1:27017")
	flag.StringVar(&cfg.MongoDb, "mongodb", envconf.FromEnvP("MH_MONGO_DB", "mailhog").(string), "MongoDB database, e.g. mailhog")
	flag.StringVar(&cfg.MongoColl, "mongocoll", envconf.FromEnvP("MH_MONGO_COLLECTION", "messages").(string), "MongoDB collection, e.g. messages")
	flag.BoolVar(&cfg.InviteJim, "invite-jim", envconf.FromEnvP("MH_INVITE_JIM", false).(bool), "Decide whether to invite Jim (beware, he causes trouble)")
	jim.RegisterFlags()
}

package config

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/doctolib/MailHog/pkg/data"
	"github.com/doctolib/MailHog/pkg/monkey"
	"github.com/doctolib/MailHog/pkg/storage"
	"github.com/ian-kent/envconf"
)

// DefaultConfig is the default config
func DefaultConfig() *Config {
	return &Config{
		SMTPBindAddr:  "0.0.0.0:1025",
		HTTPBindAddr:  "0.0.0.0:8025",
		Hostname:      "mailhog.example",
		MongoURI:      "127.0.0.1:27017",
		MongoDatabase: "mailhog",
		PostgresURI:   "postgres://127.0.0.1:5432/mailhog",
		MongoColl:     "messages",
		StorageType:   "memory",
		MessageChan:   make(chan *data.Message),
		OutgoingSMTP:  make(map[string]*OutgoingSMTP),
	}
}

// Config is the config, kind of
type Config struct {
	Verbose          bool
	SMTPBindAddr     string
	HTTPBindAddr     string
	Hostname         string
	MongoURI         string
	MongoDatabase    string
	MongoColl        string
	PostgresURI      string
	StorageType      string
	CORSOrigin       string
	MaildirPath      string
	InviteJim        bool
	Storage          storage.Storage
	MessageChan      chan *data.Message
	Assets           func(asset string) ([]byte, error)
	Monkey           monkey.ChaosMonkey
	OutgoingSMTPFile string
	OutgoingSMTP     map[string]*OutgoingSMTP
	WebPath          string
	AuthFile         string
}

// OutgoingSMTP is an outgoing SMTP server config
type OutgoingSMTP struct {
	Name      string
	Save      bool
	Email     string
	Host      string
	Port      string
	Username  string
	Password  string
	Mechanism string
}

var cfg = DefaultConfig()

// Jim is a monkey
var Jim = &monkey.Jim{}

// Configure configures stuff
func Configure() *Config {
	if cfg.Verbose {
		log.SetLevel(log.TraceLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	switch cfg.StorageType {
	case "memory":
		log.Info("Using in-memory storage")
		cfg.Storage = storage.CreateInMemory()
	case "mongodb":
		log.Info("Using MongoDB message storage")
		s := storage.CreateMongoDB(cfg.MongoURI, cfg.MongoDatabase, cfg.MongoColl)
		if s == nil {
			log.Fatal("MongoDB storage unavailable")
		} else {
			log.Infof("Connected to MongoDB")
			cfg.Storage = s
		}
	case "postgres":
		log.Info("Using PostgreSQL message storage")
		s := storage.CreatePostgreSQL(cfg.PostgresURI)
		if s == nil {
			log.Fatal("PostgreSQL storage unavailable")
		} else {
			log.Infof("Connected to PostgreSQL")
			cfg.Storage = s
		}
	case "maildir":
		log.Infof("Using Maildir message storage")
		s := storage.CreateMaildir(cfg.MaildirPath)
		cfg.Storage = s
	default:
		log.Fatalf("Invalid storage type %s", cfg.StorageType)
	}

	Jim.Configure()
	if cfg.InviteJim {
		cfg.Monkey = Jim
	}

	if len(cfg.OutgoingSMTPFile) > 0 {
		var o map[string]*OutgoingSMTP
		if b, err := ioutil.ReadFile(cfg.OutgoingSMTPFile); err != nil {
			log.Fatal(err)
		} else if err = json.Unmarshal(b, &o); err != nil {
			log.Fatal(err)
		}
		cfg.OutgoingSMTP = o
	}

	//sanitize webpath
	//add a leading slash
	if cfg.WebPath != "" && !(cfg.WebPath[0] == '/') {
		cfg.WebPath = "/" + cfg.WebPath
	}

	return cfg
}

// RegisterFlags registers flags
func RegisterFlags() {
	flag.BoolVar(&cfg.Verbose, "verbose", envconf.FromEnvP("MH_VERBOSE", false).(bool), "Be verbose (TRACE log level)")
	flag.StringVar(&cfg.SMTPBindAddr, "smtp-bind-addr", envconf.FromEnvP("MH_SMTP_BIND_ADDR", "0.0.0.0:1025").(string), "SMTP bind interface and port, e.g. 0.0.0.0:1025 or just :1025")
	flag.StringVar(&cfg.HTTPBindAddr, "api-bind-addr", envconf.FromEnvP("MH_API_BIND_ADDR", "0.0.0.0:8025").(string), "HTTP bind interface and port for API, e.g. 0.0.0.0:8025 or just :8025")
	flag.StringVar(&cfg.Hostname, "hostname", envconf.FromEnvP("MH_HOSTNAME", "mailhog.example").(string), "Hostname for EHLO/HELO response, e.g. mailhog.example")
	flag.StringVar(&cfg.StorageType, "storage", envconf.FromEnvP("MH_STORAGE", "memory").(string), "Message storage: 'memory' (default), 'mongodb' or 'maildir'")
	flag.StringVar(&cfg.MongoURI, "mongo-uri", envconf.FromEnvP("MH_MONGO_URI", "127.0.0.1:27017").(string), "MongoDB URI, e.g. 127.0.0.1:27017")
	flag.StringVar(&cfg.MongoDatabase, "mongo-db", envconf.FromEnvP("MH_MONGO_DB", "mailhog").(string), "MongoDB database, e.g. mailhog")
	flag.StringVar(&cfg.MongoColl, "mongo-coll", envconf.FromEnvP("MH_MONGO_COLLECTION", "messages").(string), "MongoDB collection, e.g. messages")
	flag.StringVar(&cfg.PostgresURI, "postgres-uri", envconf.FromEnvP("MH_POSTGRES_URI", "postgres://127.0.0.1:5432/mailhog").(string), "PostgrsQL URI, e.g. postgres://127.0.0.1:5432/mailhog")
	flag.StringVar(&cfg.CORSOrigin, "cors-origin", envconf.FromEnvP("MH_CORS_ORIGIN", "").(string), "CORS Access-Control-Allow-Origin header for API endpoints")
	flag.StringVar(&cfg.MaildirPath, "maildir-path", envconf.FromEnvP("MH_MAILDIR_PATH", "").(string), "Maildir path (if storage type is 'maildir')")
	flag.BoolVar(&cfg.InviteJim, "invite-jim", envconf.FromEnvP("MH_INVITE_JIM", false).(bool), "Decide whether to invite Jim (beware, he causes trouble)")
	flag.StringVar(&cfg.OutgoingSMTPFile, "outgoing-smtp", envconf.FromEnvP("MH_OUTGOING_SMTP", "").(string), "JSON file containing outgoing SMTP servers")
	Jim.RegisterFlags()
	flag.StringVar(&cfg.WebPath, "ui-web-path", envconf.FromEnvP("MH_UI_WEB_PATH", "").(string), "WebPath under which the UI is served (without leading or trailing slashes), e.g. 'mailhog'. Value defaults to ''")
	flag.StringVar(&cfg.AuthFile, "auth-file", envconf.FromEnvP("MH_AUTH_FILE", "").(string), "A username:bcryptpw mapping file")
}

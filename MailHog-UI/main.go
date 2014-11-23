package main

import (
	"flag"
	"io"
	"net"
	"os"

	"github.com/ian-kent/Go-MailHog/config"
	"github.com/ian-kent/Go-MailHog/data"
	mhhttp "github.com/ian-kent/Go-MailHog/http"
	"github.com/ian-kent/Go-MailHog/http/api"
	smtp "github.com/ian-kent/Go-MailHog/smtp/server"
	"github.com/ian-kent/Go-MailHog/storage"
	"github.com/ian-kent/envconf"
	"github.com/ian-kent/go-log/log"
	gotcha "github.com/ian-kent/gotcha/app"
	"github.com/ian-kent/gotcha/events"
	"github.com/ian-kent/gotcha/http"
)

var conf *config.Config
var exitCh chan int

func configure() {
	var smtpbindaddr, httpbindaddr, hostname, storage_type, mongouri, mongodb, mongocoll string

	flag.StringVar(&smtpbindaddr, "smtpbindaddr", envconf.FromEnvP("MH_SMTP_BIND_ADDR", "0.0.0.0:1025").(string), "SMTP bind interface and port, e.g. 0.0.0.0:1025 or just :1025")
	flag.StringVar(&httpbindaddr, "httpbindaddr", envconf.FromEnvP("MH_HTTP_BIND_ADDR", "0.0.0.0:8025").(string), "HTTP bind interface and port, e.g. 0.0.0.0:8025 or just :8025")
	flag.StringVar(&hostname, "hostname", envconf.FromEnvP("MH_HOSTNAME", "mailhog.example").(string), "Hostname for EHLO/HELO response, e.g. mailhog.example")
	flag.StringVar(&storage_type, "storage", envconf.FromEnvP("MH_STORAGE", "memory").(string), "Message storage: memory (default) or mongodb")
	flag.StringVar(&mongouri, "mongouri", envconf.FromEnvP("MH_MONGO_URI", "127.0.0.1:27017").(string), "MongoDB URI, e.g. 127.0.0.1:27017")
	flag.StringVar(&mongodb, "mongodb", envconf.FromEnvP("MH_MONGO_DB", "mailhog").(string), "MongoDB database, e.g. mailhog")
	flag.StringVar(&mongocoll, "mongocoll", envconf.FromEnvP("MH_MONGO_COLLECTION", "messages").(string), "MongoDB collection, e.g. messages")

	flag.Parse()

	conf = &config.Config{
		SMTPBindAddr: smtpbindaddr,
		HTTPBindAddr: httpbindaddr,
		Hostname:     hostname,
		MongoUri:     mongouri,
		MongoDb:      mongodb,
		MongoColl:    mongocoll,
		Assets:       Asset,
		MessageChan:  make(chan *data.Message),
	}

	if storage_type == "mongodb" {
		log.Println("Using MongoDB message storage")
		s := storage.CreateMongoDB(conf.MongoUri, conf.MongoDb, conf.MongoColl)
		if s == nil {
			log.Println("MongoDB storage unavailable, reverting to in-memory storage")
			conf.Storage = storage.CreateInMemory()
		} else {
			log.Println("Connected to MongoDB")
			conf.Storage = s
		}
	} else if storage_type == "memory" {
		log.Println("Using in-memory message storage")
		conf.Storage = storage.CreateInMemory()
	} else {
		log.Fatalf("Invalid storage type %s", storage_type)
	}
}

func main() {
	configure()

	exitCh = make(chan int)
	go web_listen()
	go smtp_listen()

	for {
		select {
		case <-exitCh:
			log.Printf("Received exit signal")
			os.Exit(0)
		}
	}
}

func web_listen() {
	log.Info("[HTTP] Binding to address: %s", conf.HTTPBindAddr)

	var app = gotcha.Create(Asset)
	app.Config.Listen = conf.HTTPBindAddr

	app.On(events.BeforeHandler, func(session *http.Session, next func()) {
		session.Stash["config"] = conf
		next()
	})

	r := app.Router

	r.Get("/images/(?P<file>.*)", r.Static("assets/images/{{file}}"))
	r.Get("/js/(?P<file>.*)", r.Static("assets/js/{{file}}"))
	r.Get("/", mhhttp.Index)

	api.CreateAPIv1(conf, app)

	app.Config.LeftDelim = "[:"
	app.Config.RightDelim = ":]"

	app.Start()

	<-make(chan int)
	exitCh <- 1
}

func smtp_listen() *net.TCPListener {
	log.Printf("[SMTP] Binding to address: %s\n", conf.SMTPBindAddr)
	ln, err := net.Listen("tcp", conf.SMTPBindAddr)
	if err != nil {
		log.Fatalf("[SMTP] Error listening on socket: %s\n", err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[SMTP] Error accepting connection: %s\n", err)
			continue
		}
		defer conn.Close()

		go smtp.Accept(
			conn.(*net.TCPConn).RemoteAddr().String(),
			io.ReadWriteCloser(conn),
			conf.Storage,
			conf.MessageChan,
			conf.Hostname,
		)
	}
}

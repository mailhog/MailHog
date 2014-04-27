package main

import (
	"flag"
	"github.com/ian-kent/MailHog/mailhog/config"
	"github.com/ian-kent/MailHog/mailhog/http"
	"github.com/ian-kent/MailHog/mailhog/smtp"
	"github.com/ian-kent/MailHog/mailhog/storage"
	"log"
	"net"
	"os"
)

var conf *config.Config
var exitCh chan int

func configure() {
	var smtpbindaddr, httpbindaddr, hostname, storage_type, mongouri, mongodb, mongocoll string

	flag.StringVar(&smtpbindaddr, "smtpbindaddr", "0.0.0.0:1025", "SMTP bind interface and port, e.g. 0.0.0.0:1025 or just :1025")
	flag.StringVar(&httpbindaddr, "httpbindaddr", "0.0.0.0:8025", "HTTP bind interface and port, e.g. 0.0.0.0:8025 or just :8025")
	flag.StringVar(&hostname, "hostname", "mailhog.example", "Hostname for EHLO/HELO response, e.g. mailhog.example")
	flag.StringVar(&storage_type, "storage", "memory", "Message storage: memory (default) or mongodb")
	flag.StringVar(&mongouri, "mongouri", "127.0.0.1:27017", "MongoDB URI, e.g. 127.0.0.1:27017")
	flag.StringVar(&mongodb, "mongodb", "mailhog", "MongoDB database, e.g. mailhog")
	flag.StringVar(&mongocoll, "mongocoll", "messages", "MongoDB collection, e.g. messages")

	flag.Parse()

	conf = &config.Config{
		SMTPBindAddr: smtpbindaddr,
		HTTPBindAddr: httpbindaddr,
		Hostname:     hostname,
		MongoUri:     mongouri,
		MongoDb:      mongodb,
		MongoColl:    mongocoll,
		Assets:       Asset,
	}

	if storage_type == "mongodb" {
		log.Println("Using MongoDB message storage")
		s := storage.CreateMongoDB(conf)
		if s == nil {
			log.Println("MongoDB storage unavailable, reverting to in-memory storage")
			conf.Storage = storage.CreateMemory(conf)
		} else {
			log.Println("Connected to MongoDB")
			conf.Storage = s
		}
	} else if storage_type == "memory" {
		log.Println("Using in-memory message storage")
		conf.Storage = storage.CreateMemory(conf)
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
	log.Printf("[HTTP] Binding to address: %s\n", conf.HTTPBindAddr)
	http.Start(exitCh, conf)
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

		go smtp.StartSession(conn.(*net.TCPConn), conf)
	}
}

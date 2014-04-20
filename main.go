package main

import (
	"flag"
	"log"
	"net"
	"os"
	"github.com/ian-kent/MailHog/mailhog"
	"github.com/ian-kent/MailHog/mailhog/http"
	"github.com/ian-kent/MailHog/mailhog/smtp"
)

var conf *mailhog.Config
var exitCh chan int

func config() {
	var smtpbindaddr, httpbindaddr, hostname, mongouri, mongodb, mongocoll string

	flag.StringVar(&smtpbindaddr, "smtpbindaddr", "0.0.0.0:1025", "SMTP bind interface and port, e.g. 0.0.0.0:1025 or just :1025")
	flag.StringVar(&httpbindaddr, "httpbindaddr", "0.0.0.0:8025", "HTTP bind interface and port, e.g. 0.0.0.0:8025 or just :8025")
	flag.StringVar(&hostname, "hostname", "mailhog.example", "Hostname for EHLO/HELO response, e.g. mailhog.example")
	flag.StringVar(&mongouri, "mongouri", "127.0.0.1:27017", "MongoDB URI, e.g. 127.0.0.1:27017")
	flag.StringVar(&mongodb, "mongodb", "mailhog", "MongoDB database, e.g. mailhog")
	flag.StringVar(&mongocoll, "mongocoll", "messages", "MongoDB collection, e.g. messages")

	flag.Parse()

	conf = &mailhog.Config{
		SMTPBindAddr: smtpbindaddr,
		HTTPBindAddr: httpbindaddr,
		Hostname: hostname,
		MongoUri: mongouri,
		MongoDb: mongodb,
		MongoColl: mongocoll,
	}
}

func main() {
	config()

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

func smtp_listen() (*net.TCPListener) {
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

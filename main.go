package main

import (
	"flag"
	"log"
	"net"
	"github.com/ian-kent/MailHog/mailhog"
	"github.com/ian-kent/MailHog/mailhog/smtp"
)

var conf *mailhog.Config

func config() {
	var listen, hostname string

	flag.StringVar(&listen, "listen", "0.0.0.0:1025", "Bind interface and port, e.g. 0.0.0.0:1025 or just :1025")
	flag.StringVar(&hostname, "hostname", "mailhog.example", "Hostname for EHLO/HELO response, e.g. mailhog.example")

	flag.Parse()

	conf = &mailhog.Config{
		BindAddr: listen,
		Hostname: hostname,
	}
}

func main() {
	config()

	ln := listen(conf.BindAddr)
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s\n", err)
			continue
		}
		defer conn.Close()

		go smtp.StartSession(conn.(*net.TCPConn), conf)
	}
}

func listen(bind string) (*net.TCPListener) {
	log.Printf("Binding to address: %s\n", bind)
	ln, err := net.Listen("tcp", bind)
	if err != nil {
		log.Fatalf("Error listening on socket: %s\n", err)
	}
	return ln.(*net.TCPListener)
}

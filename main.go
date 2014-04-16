package main

import (
	"flag"
	"log"
	"net"
)

var conf = map[string]string {
	"BIND_ADDRESS": "0.0.0.0:1025",
}

func config() {
	var listen string

	flag.StringVar(&listen, "listen", "0.0.0.0:1025", "Bind interface and port, e.g. 0.0.0.0:1025 or just :1025")
	flag.Parse()

	conf["BIND_ADDRESS"] = listen
}

func main() {
	config()

	ln := listen(conf["BIND_ADDRESS"])
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s\n", err)
			continue
		}
		defer conn.Close()

		go accept(conn)
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

func accept(conn net.Conn) {
	buf := make([]byte, 1024)	
	n, err := conn.Read(buf)

	if err != nil {
		log.Printf("Error reading from socket: %s", err)
		return
	}

	log.Printf("Received %s bytes: %s\n", n, string(buf))

	_, err = conn.Write(buf)
	if err != nil {
		log.Printf("Error writing to socket: %s\n", err)
		return
	}

	log.Printf("Reply sent\n")
}

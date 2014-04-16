package smtp

// http://www.rfc-editor.org/rfc/rfc5321.txt

import (
	"log"
	"net"
)

type Session struct {
	conn *net.TCPConn
}

type Message struct {
	From string
	To string
	Data []byte
	Helo string
}

func StartSession(conn *net.TCPConn) {
	conv := &Session{conn}
	conv.Begin()
}

func (c Session) Read() {
	buf := make([]byte, 1024)	
	n, err := c.conn.Read(buf)
	if n == 0 {
		log.Printf("Connection closed by remote host\n")
		return
	}
	if err != nil {
		log.Printf("Error reading from socket: %s\n", err)
		return
	}
	text := string(buf)
	log.Printf("Received %d bytes: %s\n", n, text)
	c.Parse(text)
}

func (c Session) Parse(content string) {
	log.Printf("Parsing string: %s", content)
	c.Read()
}

func (c Session) Begin() {
	_, err := c.conn.Write([]byte("220 Go-MailHog\n"))
	if err != nil {
		log.Printf("Failed writing to socket: %s\n", err)
		return
	}
	c.Read()
}

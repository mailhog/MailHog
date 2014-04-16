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

func StartSession(conn *net.TCPConn) (*Session) {
	conv := &Session{conn}
	conv.Begin()
	return conv
}

func (c Session) Begin() {
	_, err := c.conn.Write([]byte("220 Go-MailHog\n"))
	if err != nil {
		log.Printf("Failed writing to socket: %s", err)
		return
	}
}

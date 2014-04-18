package smtp

// http://www.rfc-editor.org/rfc/rfc5321.txt

import (
	"log"
	"net"
	"strings"
)

type Session struct {
	conn *net.TCPConn
	line string
}

type Message struct {
	From string
	To string
	Data []byte
	Helo string
}

func StartSession(conn *net.TCPConn) {
	conv := &Session{conn, ""}
	log.Printf("Starting session with %s", conn.RemoteAddr())
	conv.Begin()
}

func (c Session) log(message string, args ...interface{}) {
	message = strings.Join([]string{"[%s]", message}, " ")
	args = append([]interface{}{c.conn.RemoteAddr()}, args...)
	log.Printf(message, args...)
}

func (c Session) Read() {
	buf := make([]byte, 1024)
	n, err := c.conn.Read(buf)

	if n == 0 {
		c.log("Connection closed by remote host\n")
		return
	}
	if err != nil {
		c.log("Error reading from socket: %s\n", err)
		return
	}

	text := string(buf)
	c.log("Received %d bytes: %s\n", n, text)

	c.line += text

	c.Parse()
}

func (c Session) Parse() {
	for strings.Contains(c.line, "\n") {
		parts := strings.SplitN(c.line, "\n", 2)
		c.line = parts[1]
		c.log("Parsing string: %s", parts[0])
		c.Process(strings.Trim(parts[0], "\r\n"))
	}

	c.Read()
}

func (c Session) Write(code string, text ...string) {
	if len(text) == 1 {
		c.conn.Write([]byte(code + " " + text[0] + "\n"))
		return
	}
	for i := 0; i < len(text) - 2; i++ {
		c.conn.Write([]byte(code + "-" + text[i] + "\n"))
	}
	c.conn.Write([]byte(code + " " + text[len(text)] + "\n"))
}

func (c Session) Process(line string) {
	c.log("Processing line: %s", line)

	words := strings.Split(line, " ")

	switch words[0] {
		case "HELO":
			c.log("Got HELO command")
			c.Write("250", "HELO " + "my.hostname")
		case "EHLO":
			c.log("Got EHLO command")
			c.Write("250", "HELO " + "my.hostname")
		default:
			c.log("Got unknown command: '%s'", words[0])
	}
}

func (c Session) Begin() {
	_, err := c.conn.Write([]byte("220 Go-MailHog\n"))
	if err != nil {
		c.log("Failed writing to socket: %s\n", err)
		return
	}
	c.Read()
}

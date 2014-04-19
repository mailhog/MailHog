package smtp

// http://www.rfc-editor.org/rfc/rfc5321.txt

import (
	"log"
	"net"
	"strings"
	"github.com/ian-kent/MailHog/mailhog"
)

type Session struct {
	conn *net.TCPConn
	line string
	conf *mailhog.Config
	state int
}

const (
	ESTABLISH = iota
	MAIL
	RCPT
	DATA
)

type Message struct {
	From string
	To string
	Data []byte
	Helo string
}

func StartSession(conn *net.TCPConn, conf *mailhog.Config) {
	conv := &Session{conn, "", conf, ESTABLISH}
	conv.log("Starting session")
	conv.Write("220", "Go-MailHog")
	conv.Read()
}

func (c *Session) log(message string, args ...interface{}) {
	message = strings.Join([]string{"[%s, %d]", message}, " ")
	args = append([]interface{}{c.conn.RemoteAddr(), c.state}, args...)
	log.Printf(message, args...)
}

func (c *Session) Read() {
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

	text := string(buf[0:n])
	c.log("Received %d bytes: '%s'\n", n, text)

	c.line += text

	c.Parse()
}

func (c *Session) Parse() {
	for strings.Contains(c.line, "\n") {
		parts := strings.SplitN(c.line, "\n", 2)
		if len(parts) == 2 {
			c.line = parts[1]
		} else {
			c.line = ""
		}
		c.Process(strings.Trim(parts[0], "\r\n"))
	}

	c.Read()
}

func (c *Session) Write(code string, text ...string) {
	if len(text) == 1 {
		c.conn.Write([]byte(code + " " + text[0] + "\n"))
		return
	}
	for i := 0; i < len(text) - 2; i++ {
		c.conn.Write([]byte(code + "-" + text[i] + "\n"))
	}
	c.conn.Write([]byte(code + " " + text[len(text)] + "\n"))
}

func (c *Session) Process(line string) {
	c.log("Processing line: %s", line)

	words := strings.Split(line, " ")
	command := words[0]
	c.log("In state %d, got command '%s'", c.state, command)

	switch {
		case command == "RSET":
			c.log("Got RSET command, switching to ESTABLISH state")
			c.state = ESTABLISH
			c.Write("250", "OK")
		case command == "NOOP":
			c.log("Got NOOP command")
			c.Write("250", "OK")
		case command == "QUIT":
			c.log("Got QUIT command")
			c.Write("221", "OK")
		case c.state == ESTABLISH:
			switch command {
				case "HELO": 
					c.log("Got HELO command, switching to MAIL state")
					c.state = MAIL
					c.Write("250", "HELO " + c.conf.Hostname)
				case "EHLO":
					c.log("Got EHLO command, switching to MAIL state")
					c.state = MAIL
					c.Write("250", "EHLO " + c.conf.Hostname)
				default:
					c.log("Got unknown command for ESTABLISH state: '%s'", command)
			}
		case c.state == MAIL:
			switch command {
				case "MAIL":
					c.log("Got MAIL command")
				default:
					c.log("Got unknown command for MAIL state: '%s'", command)
			}
	}
}

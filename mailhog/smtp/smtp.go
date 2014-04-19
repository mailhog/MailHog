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
	message *Message
}

const (
	ESTABLISH = iota
	MAIL
	RCPT
	DATA
	DONE
)

type Message struct {
	From string
	To []string
	Data string
	Helo string
}

func StartSession(conn *net.TCPConn, conf *mailhog.Config) {
	conv := &Session{conn, "", conf, ESTABLISH, &Message{}}
	conv.log("Starting session")
	conv.Write("220", conv.conf.Hostname + " ESMTP Go-MailHog")
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
		if c.state == DATA {
			c.message.Data += parts[0] + "\n"
			if(strings.HasSuffix(c.message.Data, "\r\n.\r\n")) {
				c.state = DONE
				c.Write("250", "Ok: queued as nnnnnnnn")
			}
		} else {
			c.Process(strings.Trim(parts[0], "\r\n"))
		}
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
	args := strings.Join(words[1:len(words)], " ")
	c.log("In state %d, got command '%s', args '%s'", c.state, command, args)

	switch {
		case command == "RSET":
			c.log("Got RSET command, switching to ESTABLISH state")
			c.state = ESTABLISH
			c.message = &Message{}
			c.Write("250", "Ok")
		case command == "NOOP":
			c.log("Got NOOP command")
			c.Write("250", "Ok")
		case command == "QUIT":
			c.log("Got QUIT command")
			c.Write("221", "Bye")
		case c.state == ESTABLISH:
			switch command {
				case "HELO": 
					c.log("Got HELO command, switching to MAIL state")
					c.state = MAIL
					c.message.Helo = args
					c.Write("250", "Hello " + args)
				case "EHLO":
					c.log("Got EHLO command, switching to MAIL state")
					c.state = MAIL
					c.message.Helo = args
					c.Write("250", "Hello " + args)
				default:
					c.log("Got unknown command for ESTABLISH state: '%s'", command)
			}
		case c.state == MAIL:
			switch command {
				case "MAIL":
					c.log("Got MAIL command, switching to RCPT state")
					// TODO parse args
					c.message.From = args
					c.state = RCPT
					c.Write("250", "Ok")
				default:
					c.log("Got unknown command for MAIL state: '%s'", command)
			}
		case c.state == RCPT:
			switch command {
				case "RCPT":
					c.log("Got RCPT command")
					c.message.To = append(c.message.To, args)
					c.Write("250", "Ok")
				case "DATA":
					c.log("Got DATA command, switching to DATA state")
					c.state = DATA
					c.Write("354", "End data with <CR><LF>.<CR><LF>")
				default:
					c.log("Got unknown command for RCPT state: '%s'", command)
			}
		case c.state == DONE:
			switch command {
				case "MAIL":
					c.log("Got MAIL command")
					// TODO parse args
					c.message.From = args
					c.state = RCPT
					c.Write("250", "Ok")
				default:
					c.log("Got unknown command for DONE state: '%s'", command)
			}
	}
}

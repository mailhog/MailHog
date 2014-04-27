package smtp

// http://www.rfc-editor.org/rfc/rfc5321.txt

import (
	"log"
	"net"
	"strings"
	"regexp"
	"github.com/ian-kent/MailHog/mailhog/config"
	"github.com/ian-kent/MailHog/mailhog/storage"
	"github.com/ian-kent/MailHog/mailhog/data"
)

type Session struct {
	conn *net.TCPConn
	line string
	conf *config.Config
	state int
	message *data.SMTPMessage
	isTLS bool
}

const (
	ESTABLISH = iota
	AUTH
	AUTH2
	MAIL
	RCPT
	DATA
	DONE
)

// TODO replace ".." lines with . in data

func StartSession(conn *net.TCPConn, conf *config.Config) {
	conv := &Session{conn, "", conf, ESTABLISH, &data.SMTPMessage{}, false}
	conv.log("Starting session")
	conv.Write("220", conv.conf.Hostname + " ESMTP Go-MailHog")
	conv.Read()
}

func (c *Session) log(message string, args ...interface{}) {
	message = strings.Join([]string{"[SMTP %s, %d]", message}, " ")
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
				c.log("Got EOF, storing message and switching to MAIL state")
				//c.log("Full message data: %s", c.message.Data)
				c.message.Data = strings.TrimSuffix(c.message.Data, "\r\n.\r\n")
				msg := data.ParseSMTPMessage(c.message, c.conf.Hostname)
				var id string
				var err error
				switch c.conf.Storage.(type) {
					case *storage.MongoDB:
						c.log("Storing message using MongoDB")
						id, err = c.conf.Storage.(*storage.MongoDB).Store(msg)
					case *storage.Memory:
						c.log("Storing message using Memory")
						id, err = c.conf.Storage.(*storage.Memory).Store(msg)
					default:
						c.log("Unknown storage type")
						// TODO send error reply
				}
				c.state = MAIL
				if err != nil {
					c.log("Error storing message: %s", err)
					c.Write("452", "Unable to store message")
					return
				}
				c.Write("250", "Ok: queued as " + id)
				c.conf.MessageChan <- msg
			}
		} else {
			c.Process(strings.Trim(parts[0], "\r\n"))
		}
	}

	c.Read()
}

func (c *Session) Write(code string, text ...string) {
	if len(text) == 1 {
		c.log("Sent %d bytes: '%s'", len(text[0] + "\n"), text[0] + "\n")
		c.conn.Write([]byte(code + " " + text[0] + "\n"))
		return
	}
	for i := 0; i < len(text) - 1; i++ {
		c.log("Sent %d bytes: '%s'", len(text[i] + "\n"), text[i] + "\n")
		c.conn.Write([]byte(code + "-" + text[i] + "\n"))
	}
	c.log("Sent %d bytes: '%s'", len(text[len(text)-1] + "\n"), text[len(text)-1] + "\n")
	c.conn.Write([]byte(code + " " + text[len(text)-1] + "\n"))
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
			c.message = &data.SMTPMessage{}
			c.Write("250", "Ok")
		case command == "NOOP":
			c.log("Got NOOP command")
			c.Write("250", "Ok")
		case command == "QUIT":
			c.log("Got QUIT command")
			c.Write("221", "Bye")
			err := c.conn.Close()
			if err != nil {
				c.log("Error closing connection")
			}
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
					c.Write("250", "Hello " + args, "PIPELINING", "AUTH EXTERNAL CRAM-MD5 LOGIN PLAIN")
				default:
					c.log("Got unknown command for ESTABLISH state: '%s'", command)
					c.Write("500", "Unrecognised command")
			}
		case c.state == AUTH:
			c.log("Got authentication response: '%s', switching to MAIL state", args)
			c.state = MAIL
			c.Write("235", "Authentication successful")
		case c.state == AUTH2:
			c.log("Got LOGIN authentication response: '%s', switching to AUTH state", args)
			c.state = AUTH
			c.Write("334", "UGFzc3dvcmQ6")
		case c.state == MAIL: // TODO rename/split state
			switch command {
				case "AUTH":
					c.log("Got AUTH command, staying in MAIL state")
					switch {
						case strings.HasPrefix(args, "PLAIN "):
							c.log("Got PLAIN authentication: %s", strings.TrimPrefix(args, "PLAIN "))
							c.Write("235", "Authentication successful")
						case args == "LOGIN":
							c.log("Got LOGIN authentication, switching to AUTH state")
							c.state = AUTH2
							c.Write("334", "VXNlcm5hbWU6")
						case args == "PLAIN":
							c.log("Got PLAIN authentication (no args), switching to AUTH2 state")
							c.state = AUTH
							c.Write("334", "")
						case args == "CRAM-MD5":
							c.log("Got CRAM-MD5 authentication, switching to AUTH state")
							c.state = AUTH
							c.Write("334", "PDQxOTI5NDIzNDEuMTI4Mjg0NzJAc291cmNlZm91ci5hbmRyZXcuY211LmVkdT4=")
						case strings.HasPrefix(args, "EXTERNAL "):
							c.log("Got EXTERNAL authentication: %s", strings.TrimPrefix(args, "EXTERNAL "))
							c.Write("235", "Authentication successful")
						default:
							c.Write("504", "Unsupported authentication mechanism")
					}
				case "MAIL":
					c.log("Got MAIL command, switching to RCPT state")
					r, _ := regexp.Compile("(?i:From):<([^>]+)>")
					match := r.FindStringSubmatch(args)
					c.message.From = match[1]
					c.state = RCPT
					c.Write("250", "Sender " + match[1] + " ok")
				default:
					c.log("Got unknown command for MAIL state: '%s'", command)
					c.Write("500", "Unrecognised command")
			}
		case c.state == RCPT:
			switch command {
				case "RCPT":
					c.log("Got RCPT command")
					r, _ := regexp.Compile("(?i:To):<([^>]+)>")
					match := r.FindStringSubmatch(args)
					c.message.To = append(c.message.To, match[1])
					c.state = RCPT
					c.Write("250", "Recipient " + match[1] + " ok")
				case "DATA":
					c.log("Got DATA command, switching to DATA state")
					c.state = DATA
					c.Write("354", "End data with <CR><LF>.<CR><LF>")
				default:
					c.log("Got unknown command for RCPT state: '%s'", command)
					c.Write("500", "Unrecognised command")
			}
	}
}

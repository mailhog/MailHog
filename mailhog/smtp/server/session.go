package server

// http://www.rfc-editor.org/rfc/rfc5321.txt

import (
	"errors"
	"log"
	"net"
	"strings"

	"github.com/ian-kent/Go-MailHog/mailhog/config"
	"github.com/ian-kent/Go-MailHog/mailhog/data"
	"github.com/ian-kent/Go-MailHog/mailhog/smtp/protocol"
	"github.com/ian-kent/Go-MailHog/mailhog/storage"
)

// Session represents a SMTP session using net.TCPConn
type Session struct {
	conn  *net.TCPConn
	proto *protocol.Protocol
	conf  *config.Config
	isTLS bool
	line  string
}

// Accept starts a new SMTP session using net.TCPConn
func Accept(conn *net.TCPConn, conf *config.Config) {
	proto := protocol.NewProtocol()
	session := &Session{conn, proto, conf, false, ""}
	proto.LogHandler = session.logf
	proto.MessageReceivedHandler = session.acceptMessage
	proto.ValidateSenderHandler = session.validateSender
	proto.ValidateRecipientHandler = session.validateRecipient
	proto.ValidateAuthenticationHandler = session.validateAuthentication

	session.logf("Starting session")
	session.Write(proto.Start(conf.Hostname))
	for session.Read() == true {
	}
	session.logf("Session ended")
}

func (c *Session) validateAuthentication(mechanism string, args ...string) bool {
	return true
}
func (c *Session) validateRecipient(to string) bool {
	return true
}

func (c *Session) validateSender(from string) bool {
	return true
}

func (c *Session) acceptMessage(msg *data.Message) (id string, err error) {
	switch c.conf.Storage.(type) {
	case *storage.MongoDB:
		c.logf("Storing message using MongoDB")
		id, err = c.conf.Storage.(*storage.MongoDB).Store(msg)
	case *storage.InMemory:
		c.logf("Storing message using Memory")
		id, err = c.conf.Storage.(*storage.InMemory).Store(msg)
	default:
		err = errors.New("Unknown storage stype")
	}
	c.conf.MessageChan <- msg
	return
}

func (c *Session) logf(message string, args ...interface{}) {
	message = strings.Join([]string{"[SMTP %s]", message}, " ")
	args = append([]interface{}{c.conn.RemoteAddr()}, args...)
	log.Printf(message, args...)
}

// Read reads from the underlying net.TCPConn
func (c *Session) Read() bool {
	buf := make([]byte, 1024)
	n, err := c.conn.Read(buf)

	if n == 0 {
		c.logf("Connection closed by remote host\n")
		return false
	}
	if err != nil {
		c.logf("Error reading from socket: %s\n", err)
		return false
	}

	text := string(buf[0:n])
	logText := strings.Replace(text, "\n", "\\n", -1)
	logText = strings.Replace(logText, "\r", "\\r", -1)
	c.logf("Received %d bytes: '%s'\n", n, logText)

	c.line += text

	line, reply := c.proto.Parse(c.line)
	c.line = line

	if reply != nil {
		c.Write(reply)
		if reply.Status == 221 {
			c.conn.Close()
		}
	}

	return true
}

// Write writes a reply to the underlying net.TCPConn
func (c *Session) Write(reply *protocol.Reply) {
	lines := reply.Lines()
	for _, l := range lines {
		logText := strings.Replace(l, "\n", "\\n", -1)
		logText = strings.Replace(logText, "\r", "\\r", -1)
		c.logf("Sent %d bytes: '%s'", len(l), logText)
		c.conn.Write([]byte(l))
	}
}

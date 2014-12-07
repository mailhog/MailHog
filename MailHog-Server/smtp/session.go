package smtp

// http://www.rfc-editor.org/rfc/rfc5321.txt

import (
	"io"
	"log"
	"strings"

	"github.com/ian-kent/Go-MailHog/data"
	"github.com/ian-kent/Go-MailHog/smtp/protocol"
	"github.com/ian-kent/Go-MailHog/storage"
)

// Session represents a SMTP session using net.TCPConn
type Session struct {
	conn          io.ReadWriteCloser
	proto         *protocol.Protocol
	storage       storage.Storage
	messageChan   chan *data.Message
	remoteAddress string
	isTLS         bool
	line          string
}

// Accept starts a new SMTP session using io.ReadWriteCloser
func Accept(remoteAddress string, conn io.ReadWriteCloser, storage storage.Storage, messageChan chan *data.Message, hostname string) {
	proto := protocol.NewProtocol()
	proto.Hostname = hostname
	session := &Session{conn, proto, storage, messageChan, remoteAddress, false, ""}
	proto.LogHandler = session.logf
	proto.MessageReceivedHandler = session.acceptMessage
	proto.ValidateSenderHandler = session.validateSender
	proto.ValidateRecipientHandler = session.validateRecipient
	proto.ValidateAuthenticationHandler = session.validateAuthentication

	session.logf("Starting session")
	session.Write(proto.Start())
	for session.Read() == true {
	}
	session.logf("Session ended")
}

func (c *Session) validateAuthentication(mechanism string, args ...string) (errorReply *protocol.Reply, ok bool) {
	return nil, true
}

func (c *Session) validateRecipient(to string) bool {
	return true
}

func (c *Session) validateSender(from string) bool {
	return true
}

func (c *Session) acceptMessage(msg *data.Message) (id string, err error) {
	c.logf("Storing message %s", msg.ID)
	id, err = c.storage.Store(msg)
	c.messageChan <- msg
	return
}

func (c *Session) logf(message string, args ...interface{}) {
	message = strings.Join([]string{"[SMTP %s]", message}, " ")
	args = append([]interface{}{c.remoteAddress}, args...)
	log.Printf(message, args...)
}

// Read reads from the underlying net.TCPConn
func (c *Session) Read() bool {
	buf := make([]byte, 1024)
	n, err := io.Reader(c.conn).Read(buf)

	if n == 0 {
		c.logf("Connection closed by remote host\n")
		io.Closer(c.conn).Close() // not sure this is necessary?
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

	for strings.Contains(c.line, "\n") {
		line, reply := c.proto.Parse(c.line)
		c.line = line

		if reply != nil {
			c.Write(reply)
			if reply.Status == 221 {
				io.Closer(c.conn).Close()
			}
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
		io.Writer(c.conn).Write([]byte(l))
	}
}

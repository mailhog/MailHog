package smtp

// http://www.rfc-editor.org/rfc/rfc5321.txt

import (
	"io"
	"log"
	"strings"

	"github.com/ian-kent/linkio"
	"github.com/mailhog/MailHog-Server/monkey"
	"github.com/mailhog/data"
	"github.com/mailhog/smtp"
	storagepkg "github.com/mailhog/storage"
)

// Session represents a SMTP session using net.TCPConn
type Session struct {
	conn          io.ReadWriteCloser
	proto         *smtp.Protocol
	storage       storagepkg.Storage
	messageChan   chan *data.Message
	remoteAddress string
	isTLS         bool
	line          string
	link          *linkio.Link

	reader io.Reader
	writer io.Writer
	monkey monkey.ChaosMonkey

	messageWriter storagepkg.MessageWriter
}

// Accept starts a new SMTP session using io.ReadWriteCloser
func Accept(remoteAddress string, conn io.ReadWriteCloser, storage storagepkg.Storage, messageChan chan *data.Message, hostname string, monkey monkey.ChaosMonkey) {
	defer conn.Close()

	proto := smtp.NewProtocol()
	proto.Hostname = hostname
	var link *linkio.Link
	reader := io.Reader(conn)
	writer := io.Writer(conn)
	if monkey != nil {
		linkSpeed := monkey.LinkSpeed()
		if linkSpeed != nil {
			link = linkio.NewLink(*linkSpeed * linkio.BytePerSecond)
			reader = link.NewLinkReader(io.Reader(conn))
			writer = link.NewLinkWriter(io.Writer(conn))
		}
	}

	session := &Session{conn, proto, storage, messageChan, remoteAddress, false, "", link, reader, writer, monkey, nil}
	proto.LogHandler = session.logf
	proto.MessageReceivedHandler = session.acceptMessage
	if _, ok := storage.(storagepkg.StreamingStorage); ok {
		proto.DataBeginHandler = session.beginMessage
		proto.DataLineHandler = session.writeMessageLine
		proto.DataAbortHandler = session.abortMessage
	}
	proto.ValidateSenderHandler = session.validateSender
	proto.ValidateRecipientHandler = session.validateRecipient
	proto.ValidateAuthenticationHandler = session.validateAuthentication
	proto.GetAuthenticationMechanismsHandler = func() []string { return []string{"PLAIN"} }

	session.logf("Starting session")
	session.Write(proto.Start())
	for session.Read() == true {
		if monkey != nil && monkey.Disconnect() {
			session.conn.Close()
			break
		}
	}
	session.abortMessage()
	session.logf("Session ended")
}

func (c *Session) validateAuthentication(mechanism string, args ...string) (errorReply *smtp.Reply, ok bool) {
	if c.monkey != nil {
		ok := c.monkey.ValidAUTH(mechanism, args...)
		if !ok {
			// FIXME better error?
			return smtp.ReplyUnrecognisedCommand(), false
		}
	}
	return nil, true
}

func (c *Session) validateRecipient(to string) bool {
	if c.monkey != nil {
		ok := c.monkey.ValidRCPT(to)
		if !ok {
			return false
		}
	}
	return true
}

func (c *Session) validateSender(from string) bool {
	if c.monkey != nil {
		ok := c.monkey.ValidMAIL(from)
		if !ok {
			return false
		}
	}
	return true
}

func (c *Session) acceptMessage(msg *data.SMTPMessage) (id string, err error) {
	if c.messageWriter != nil {
		var m *data.Message
		id, m, err = c.messageWriter.Commit()
		c.messageWriter = nil
		if err != nil {
			return "", err
		}
		c.messageChan <- m
		return id, nil
	}

	m := msg.Parse(c.proto.Hostname)
	c.logf("Storing message %s", m.ID)
	id, err = c.storage.Store(m)
	c.messageChan <- m
	return
}

func (c *Session) beginMessage(msg *data.SMTPMessage) error {
	streamingStorage, ok := c.storage.(storagepkg.StreamingStorage)
	if !ok {
		return nil
	}
	writer, err := streamingStorage.CreateMessageWriter(msg, c.proto.Hostname)
	if err != nil {
		return err
	}
	c.messageWriter = writer
	return nil
}

func (c *Session) writeMessageLine(line string) error {
	if c.messageWriter == nil {
		return nil
	}
	return c.messageWriter.WriteLine(line)
}

func (c *Session) abortMessage() {
	if c.messageWriter == nil {
		return
	}
	c.messageWriter.Abort()
	c.messageWriter = nil
}

func (c *Session) logf(message string, args ...interface{}) {
	message = strings.Join([]string{"[SMTP %s]", message}, " ")
	args = append([]interface{}{c.remoteAddress}, args...)
	log.Printf(message, args...)
}

// Read reads from the underlying net.TCPConn
func (c *Session) Read() bool {
	buf := make([]byte, 1024)
	n, err := c.reader.Read(buf)

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
	if c.proto.State != smtp.DATA {
		c.logf("Received %d bytes\n", n)
	}

	c.line += text

	for strings.Contains(c.line, "\r\n") {
		line, reply := c.proto.Parse(c.line)
		c.line = line

		if reply != nil {
			c.Write(reply)
			if reply.Status == 221 {
				io.Closer(c.conn).Close()
				return false
			}
		}
	}

	return true
}

// Write writes a reply to the underlying net.TCPConn
func (c *Session) Write(reply *smtp.Reply) {
	lines := reply.Lines()
	for _, l := range lines {
		logText := strings.Replace(l, "\n", "\\n", -1)
		logText = strings.Replace(logText, "\r", "\\r", -1)
		c.logf("Sent %d bytes: '%s'", len(l), logText)
		c.writer.Write([]byte(l))
	}
}

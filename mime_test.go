package main

import (
	"github.com/ian-kent/MailHog/mailhog/config"
	"github.com/ian-kent/MailHog/mailhog/storage"
	"github.com/stretchr/testify/assert"
	"net"
	"regexp"
	"strings"
	"testing"
)

// FIXME requires a running instance of MailHog

func TestBasicMIMEHappyPath(t *testing.T) {
	buf := make([]byte, 1024)

	// Open a connection
	conn, err := net.Dial("tcp", "127.0.0.1:1025")
	assert.Nil(t, err)

	// Read the greeting
	n, err := conn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, string(buf[0:n]), "220 mailhog.example ESMTP Go-MailHog\n")

	// Send EHLO
	_, err = conn.Write([]byte("EHLO localhost\r\n"))
	assert.Nil(t, err)

	// Read the response
	n, err = conn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, string(buf[0:n]), "250-Hello localhost\n")
	n, err = conn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, string(buf[0:n]), "250-PIPELINING\n")
	n, err = conn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, string(buf[0:n]), "250 AUTH EXTERNAL CRAM-MD5 LOGIN PLAIN\n")

	// Send MAIL
	_, err = conn.Write([]byte("MAIL From:<nobody@mailhog.example>\r\n"))
	assert.Nil(t, err)

	// Read the response
	n, err = conn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, string(buf[0:n]), "250 Sender nobody@mailhog.example ok\n")

	// Send RCPT
	_, err = conn.Write([]byte("RCPT To:<someone@mailhog.example>\r\n"))
	assert.Nil(t, err)

	// Read the response
	n, err = conn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, string(buf[0:n]), "250 Recipient someone@mailhog.example ok\n")

	// Send DATA
	_, err = conn.Write([]byte("DATA\r\n"))
	assert.Nil(t, err)

	// Read the response
	n, err = conn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, string(buf[0:n]), "354 End data with <CR><LF>.<CR><LF>\n")

	// Send the message
	content := "Content-Type: multipart/alternative; boundary=\"--mailhog-test-boundary\"\r\n"
	content += "Content-Length: 220\r\n"
	content += "From: Nobody <nobody@mailhog.example>\r\n"
	content += "To: Someone <someone@mailhog.example>\r\n"
	content += "Subject: Example message\r\n"
	content += "MIME-Version: 1.0\r\n"
	content += "\r\n"
	content += "----mailhog-test-boundary\r\n"
	content += "Content-Type: text/plain\r\n"
	content += "\r\n"
	content += "Hi there :)\r\n"
	content += "----mailhog-test-boundary\r\n"
	content += "Content-Type: text/html\r\n"
	content += "\r\n"
	content += "<html>\r\n"
	content += "  <head>\r\n"
	content += "    <title>Example message</title>\r\n"
	content += "  </head>\r\n"
	content += "  <body>\r\n"
	content += "    <p style=\"font-weight: bold; color: #ff0000; text-decoration: underline\">\r\n"
	content += "      Hi there :)\r\n"
	content += "    </p>\r\n"
	content += "  </body>\r\n"
	content += "</html>\r\n"
	content += "----mailhog-test-boundary\r\n"
	content += ".\r\n"
	_, err = conn.Write([]byte(content))
	assert.Nil(t, err)

	// Read the response
	n, err = conn.Read(buf)
	assert.Nil(t, err)
	r, _ := regexp.Compile("250 Ok: queued as ([0-9a-f]+)\n")
	match := r.FindStringSubmatch(string(buf[0:n]))
	assert.NotNil(t, match)

	// Send QUIT
	_, err = conn.Write([]byte("QUIT\r\n"))
	assert.Nil(t, err)

	// Read the response
	n, err = conn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, string(buf[0:n]), "221 Bye\n")

	s := storage.CreateMongoDB(config.DefaultConfig())
	message, err := s.Load(match[1])
	assert.Nil(t, err)
	assert.NotNil(t, message)

	assert.Equal(t, message.From.Domain, "mailhog.example", "sender domain is mailhog.example")
	assert.Equal(t, message.From.Mailbox, "nobody", "sender mailbox is nobody")
	assert.Equal(t, message.From.Params, "", "sender params is empty")
	assert.Equal(t, len(message.From.Relays), 0, "sender has no relays")

	assert.Equal(t, len(message.To), 1, "message has 1 recipient")

	assert.Equal(t, message.To[0].Domain, "mailhog.example", "recipient domain is mailhog.example")
	assert.Equal(t, message.To[0].Mailbox, "someone", "recipient mailbox is someone")
	assert.Equal(t, message.To[0].Params, "", "recipient params is empty")
	assert.Equal(t, len(message.To[0].Relays), 0, "recipient has no relays")

	assert.Equal(t, len(message.Content.Headers), 9, "message has 7 headers")
	assert.Equal(t, message.Content.Headers["Content-Type"], []string{"multipart/alternative; boundary=\"--mailhog-test-boundary\""}, "Content-Type header is multipart/alternative; boundary=\"--mailhog-test-boundary\"")
	assert.Equal(t, message.Content.Headers["Subject"], []string{"Example message"}, "Subject header is Example message")
	assert.Equal(t, message.Content.Headers["Content-Length"], []string{"220"}, "Content-Length is 220")
	assert.Equal(t, message.Content.Headers["To"], []string{"Someone <someone@mailhog.example>"}, "To is Someone <someone@mailhog.example>")
	assert.Equal(t, message.Content.Headers["From"], []string{"Nobody <nobody@mailhog.example>"}, "From is Nobody <nobody@mailhog.example>")
	assert.True(t, strings.HasPrefix(message.Content.Headers["Received"][0], "from localhost by mailhog.example (Go-MailHog)\r\n          id "+match[1]+"@mailhog.example; "), "Received header is correct")
	assert.Equal(t, message.Content.Headers["Return-Path"], []string{"<nobody@mailhog.example>"}, "Return-Path is <nobody@mailhog.example>")
	assert.Equal(t, message.Content.Headers["Message-ID"], []string{match[1] + "@mailhog.example"}, "Message-ID is "+match[1]+"@mailhog.example")

	expected := "----mailhog-test-boundary\r\nContent-Type: text/plain\r\n\r\nHi there :)\r\n----mailhog-test-boundary\r\nContent-Type: text/html\r\n\r\n<html>\r\n  <head>\r\n    <title>Example message</title>\r\n  </head>\r\n  <body>\r\n    <p style=\"font-weight: bold; color: #ff0000; text-decoration: underline\">\r\n      Hi there :)\r\n    </p>\r\n  </body>\r\n</html>\r\n----mailhog-test-boundary"
	assert.Equal(t, message.Content.Body, expected, "message has correct body")

	assert.NotNil(t, message.MIME)
	assert.Equal(t, len(message.MIME.Parts), 2)

	plain := message.MIME.Parts[0]
	assert.NotNil(t, plain)
	assert.Equal(t, plain.Headers["Content-Type"], []string{"text/plain"}, "Content-Type header is text/plain")
	assert.Equal(t, plain.Body, "Hi there :)", "plain text body is correct")

	html := message.MIME.Parts[1]
	assert.NotNil(t, html)
	assert.Equal(t, html.Headers["Content-Type"], []string{"text/html"}, "Content-Type header is text/html")
	expected = "<html>\r\n  <head>\r\n    <title>Example message</title>\r\n  </head>\r\n  <body>\r\n    <p style=\"font-weight: bold; color: #ff0000; text-decoration: underline\">\r\n      Hi there :)\r\n    </p>\r\n  </body>\r\n</html>"
	assert.Equal(t, html.Body, expected, "html body is correct")

}

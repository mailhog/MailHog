package main

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

// FIXME requires a running instance of MailHog
// FIXME clean up tests, repeated conn.Read(buf) is a mess

func TestBasicSMTPAuth(t *testing.T) {
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

	// Send AUTH
	_, err = conn.Write([]byte("AUTH EXTERNAL =\r\n"))
	assert.Nil(t, err)

	// Read the response
	n, err = conn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, string(buf[0:n]), "235 Authentication successful\n")

	// Send RSET and EHLO
	_, err = conn.Write([]byte("RSET\r\n"))
	n, err = conn.Read(buf)
	_, err = conn.Write([]byte("EHLO localhost\r\n"))
	n, err = conn.Read(buf)
	n, err = conn.Read(buf)
	n, err = conn.Read(buf)

	// Send AUTH
	_, err = conn.Write([]byte("AUTH PLAIN foobar\r\n"))
	assert.Nil(t, err)

	// Read the response
	n, err = conn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, string(buf[0:n]), "235 Authentication successful\n")

	// Send RSET and EHLO
	_, err = conn.Write([]byte("RSET\r\n"))
	n, err = conn.Read(buf)
	_, err = conn.Write([]byte("EHLO localhost\r\n"))
	n, err = conn.Read(buf)
	n, err = conn.Read(buf)
	n, err = conn.Read(buf)

	// Send AUTH
	_, err = conn.Write([]byte("AUTH PLAIN\r\n"))
	assert.Nil(t, err)

	// Read the response
	n, err = conn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, string(buf[0:n]), "334 \n")

	// Send AUTH
	_, err = conn.Write([]byte("foobar\r\n"))
	assert.Nil(t, err)

	// Read the response
	n, err = conn.Read(buf)
	assert.Nil(t, err)
	assert.Equal(t, string(buf[0:n]), "235 Authentication successful\n")
}

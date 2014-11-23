package protocol

// http://www.rfc-editor.org/rfc/rfc5321.txt

import (
	"testing"

	"github.com/ian-kent/Go-MailHog/mailhog/data"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProtocol(t *testing.T) {
	Convey("NewProtocol returns a new Protocol", t, func() {
		proto := NewProtocol()
		So(proto, ShouldNotBeNil)
		So(proto, ShouldHaveSameTypeAs, &Protocol{})
		So(proto.Hostname, ShouldEqual, "mailhog.example")
		So(proto.Ident, ShouldEqual, "ESMTP Go-MailHog")
		So(proto.state, ShouldEqual, INVALID)
		So(proto.message, ShouldNotBeNil)
		So(proto.message, ShouldHaveSameTypeAs, &data.SMTPMessage{})
	})

	Convey("Start should modify the state correctly", t, func() {
		proto := NewProtocol()
		So(proto.state, ShouldEqual, INVALID)
		reply := proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)
		So(reply, ShouldNotBeNil)
		So(reply, ShouldHaveSameTypeAs, &Reply{})
		So(reply.Status, ShouldEqual, 220)
		So(reply.Lines(), ShouldResemble, []string{"220 mailhog.example ESMTP Go-MailHog\n"})
	})

	Convey("Modifying the hostname should modify the ident reply", t, func() {
		proto := NewProtocol()
		proto.Ident = "OinkSMTP Go-MailHog"
		reply := proto.Start()
		So(reply, ShouldNotBeNil)
		So(reply, ShouldHaveSameTypeAs, &Reply{})
		So(reply.Status, ShouldEqual, 220)
		So(reply.Lines(), ShouldResemble, []string{"220 mailhog.example OinkSMTP Go-MailHog\n"})
	})

	Convey("Modifying the ident should modify the ident reply", t, func() {
		proto := NewProtocol()
		proto.Hostname = "oink.oink"
		reply := proto.Start()
		So(reply, ShouldNotBeNil)
		So(reply, ShouldHaveSameTypeAs, &Reply{})
		So(reply.Status, ShouldEqual, 220)
		So(reply.Lines(), ShouldResemble, []string{"220 oink.oink ESMTP Go-MailHog\n"})
	})
}

func TestEHLO(t *testing.T) {
	Convey("EHLO should modify the state correctly", t, func() {
		proto := NewProtocol()
		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)
		So(proto.message.Helo, ShouldEqual, "")
		reply := proto.EHLO("localhost")
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(reply.Lines(), ShouldResemble, []string{"250-Hello localhost\n", "250-PIPELINING\n", "250 AUTH EXTERNAL CRAM-MD5 LOGIN PLAIN\n"})
		So(proto.state, ShouldEqual, MAIL)
		So(proto.message.Helo, ShouldEqual, "localhost")
	})
}

func TestHELO(t *testing.T) {
	Convey("HELO should modify the state correctly", t, func() {
		proto := NewProtocol()
		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)
		So(proto.message.Helo, ShouldEqual, "")
		reply := proto.HELO("localhost")
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(reply.Lines(), ShouldResemble, []string{"250 Hello localhost\n"})
		So(proto.state, ShouldEqual, MAIL)
		So(proto.message.Helo, ShouldEqual, "localhost")
	})
}

func TestRSET(t *testing.T) {
	Convey("RSET should reset the state correctly", t, func() {
		proto := NewProtocol()
		proto.Start()
		proto.HELO("localhost")
		proto.Command(&Command{"MAIL", "FROM:<test>"})
		proto.Command(&Command{"RCPT", "TO:<test>"})
		So(proto.state, ShouldEqual, RCPT)
		So(proto.message.From, ShouldEqual, "test")
		So(len(proto.message.To), ShouldEqual, 1)
		So(proto.message.To[0], ShouldEqual, "test")
		reply := proto.Command(&Command{"RSET", ""})
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(reply.Lines(), ShouldResemble, []string{"250 Ok\n"})
		So(proto.state, ShouldEqual, MAIL)
		So(proto.message.From, ShouldEqual, "")
		So(len(proto.message.To), ShouldEqual, 0)
	})
}

func TestParseMAIL(t *testing.T) {
	Convey("ParseMAIL should parse MAIL command arguments", t, func() {
		m, err := ParseMAIL("FROM:<oink@mailhog.example>")
		So(err, ShouldBeNil)
		So(m, ShouldEqual, "oink@mailhog.example")
		m, err = ParseMAIL("FROM:<oink>")
		So(err, ShouldBeNil)
		So(m, ShouldEqual, "oink")
	})
	Convey("ParseMAIL should return an error for invalid syntax", t, func() {
		m, err := ParseMAIL("FROM:oink")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Invalid syntax in MAIL command")
		So(m, ShouldEqual, "")
	})
	Convey("ParseMAIL should be case-insensitive", t, func() {
		m, err := ParseMAIL("FROM:<oink>")
		So(err, ShouldBeNil)
		So(m, ShouldEqual, "oink")
		m, err = ParseMAIL("from:<oink@mailhog.example>")
		So(err, ShouldBeNil)
		So(m, ShouldEqual, "oink@mailhog.example")
		m, err = ParseMAIL("FrOm:<oink@oink.mailhog.example>")
		So(err, ShouldBeNil)
		So(m, ShouldEqual, "oink@oink.mailhog.example")
	})
}

func TestParseRCPT(t *testing.T) {
	Convey("ParseRCPT should parse RCPT command arguments", t, func() {
		m, err := ParseRCPT("TO:<oink@mailhog.example>")
		So(err, ShouldBeNil)
		So(m, ShouldEqual, "oink@mailhog.example")
		m, err = ParseRCPT("TO:<oink>")
		So(err, ShouldBeNil)
		So(m, ShouldEqual, "oink")
	})
	Convey("ParseRCPT should return an error for invalid syntax", t, func() {
		m, err := ParseRCPT("TO:oink")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Invalid syntax in RCPT command")
		So(m, ShouldEqual, "")
	})
	Convey("ParseRCPT should be case-insensitive", t, func() {
		m, err := ParseRCPT("TO:<oink>")
		So(err, ShouldBeNil)
		So(m, ShouldEqual, "oink")
		m, err = ParseRCPT("to:<oink@mailhog.example>")
		So(err, ShouldBeNil)
		So(m, ShouldEqual, "oink@mailhog.example")
		m, err = ParseRCPT("To:<oink@oink.mailhog.example>")
		So(err, ShouldBeNil)
		So(m, ShouldEqual, "oink@oink.mailhog.example")
	})
}

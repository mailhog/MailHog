package protocol

// http://www.rfc-editor.org/rfc/rfc5321.txt

import (
	"errors"
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

	Convey("LogHandler should be called for logging", t, func() {
		proto := NewProtocol()
		handlerCalled := false
		proto.LogHandler = func(message string, args ...interface{}) {
			handlerCalled = true
			So(message, ShouldEqual, "[PROTO: %s] Test message %s %s")
			So(len(args), ShouldEqual, 3)
			So(args[0], ShouldEqual, "INVALID")
			So(args[1], ShouldEqual, "test arg 1")
			So(args[2], ShouldEqual, "test arg 2")
		}
		proto.logf("Test message %s %s", "test arg 1", "test arg 2")
		So(handlerCalled, ShouldBeTrue)
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

func TestProcessCommand(t *testing.T) {
	Convey("ProcessCommand should attempt to process anything", t, func() {
		proto := NewProtocol()

		reply := proto.ProcessCommand("OINK mailhog.example")
		So(proto.state, ShouldEqual, INVALID)
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 500)
		So(reply.Lines(), ShouldResemble, []string{"500 Unrecognised command\n"})

		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)

		reply = proto.ProcessCommand("HELO localhost")
		So(proto.state, ShouldEqual, MAIL)
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(reply.Lines(), ShouldResemble, []string{"250 Hello localhost\n"})

		reply = proto.ProcessCommand("OINK mailhog.example")
		So(proto.state, ShouldEqual, MAIL)
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 500)
		So(reply.Lines(), ShouldResemble, []string{"500 Unrecognised command\n"})
	})
}

func TestParse(t *testing.T) {
	Convey("Parse can parse partial and multiple commands", t, func() {
		proto := NewProtocol()
		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)

		line, reply := proto.Parse("HELO localhost")
		So(proto.state, ShouldEqual, ESTABLISH)
		So(reply, ShouldBeNil)
		So(line, ShouldEqual, "HELO localhost")

		line, reply = proto.Parse("HELO localhost\nMAIL Fro")
		So(proto.state, ShouldEqual, MAIL)
		So(reply, ShouldNotBeNil)
		So(line, ShouldEqual, "MAIL Fro")

		line, reply = proto.Parse("MAIL From:<test>\n")
		So(proto.state, ShouldEqual, RCPT)
		So(reply, ShouldNotBeNil)
		So(line, ShouldEqual, "")
	})
	Convey("Parse can call ProcessData", t, func() {
		proto := NewProtocol()
		proto.Start()
		proto.Command(ParseCommand("EHLO localhost"))
		proto.Command(ParseCommand("MAIL From:<test>"))
		proto.Command(ParseCommand("RCPT To:<test>"))
		proto.Command(ParseCommand("DATA"))
		So(proto.state, ShouldEqual, DATA)

		line, reply := proto.Parse("Hi\n")
		So(proto.state, ShouldEqual, DATA)
		So(line, ShouldEqual, "")
		So(proto.message.Data, ShouldEqual, "Hi\n")
		So(reply, ShouldBeNil)

		line, reply = proto.Parse("\r\n")
		So(proto.state, ShouldEqual, DATA)
		So(line, ShouldEqual, "")
		So(proto.message.Data, ShouldEqual, "Hi\n\r\n")
		So(reply, ShouldBeNil)

		line, reply = proto.Parse(".\r\n")
		So(proto.state, ShouldEqual, MAIL)
		So(line, ShouldEqual, "")
		So(reply, ShouldNotBeNil)
		So(proto.message.Data, ShouldEqual, "Hi\n")
	})
}

func TestUnknownCommands(t *testing.T) {
	Convey("Unknown command in INVALID state", t, func() {
		proto := NewProtocol()
		So(proto.state, ShouldEqual, INVALID)
		reply := proto.Command(ParseCommand("OINK"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 500)
		So(reply.Lines(), ShouldResemble, []string{"500 Unrecognised command\n"})
	})
	Convey("Unknown command in ESTABLISH state", t, func() {
		proto := NewProtocol()
		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)
		reply := proto.Command(ParseCommand("OINK"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 500)
		So(reply.Lines(), ShouldResemble, []string{"500 Unrecognised command\n"})
	})
	Convey("Unknown command in MAIL state", t, func() {
		proto := NewProtocol()
		proto.Start()
		proto.Command(ParseCommand("EHLO localhost"))
		So(proto.state, ShouldEqual, MAIL)
		reply := proto.Command(ParseCommand("OINK"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 500)
		So(reply.Lines(), ShouldResemble, []string{"500 Unrecognised command\n"})
	})
	Convey("Unknown command in RCPT state", t, func() {
		proto := NewProtocol()
		proto.Start()
		proto.Command(ParseCommand("EHLO localhost"))
		proto.Command(ParseCommand("MAIL FROM:<test>"))
		So(proto.state, ShouldEqual, RCPT)
		reply := proto.Command(ParseCommand("OINK"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 500)
		So(reply.Lines(), ShouldResemble, []string{"500 Unrecognised command\n"})
	})
}

func TestESTABLISHCommands(t *testing.T) {
	Convey("EHLO should work in ESTABLISH state", t, func() {
		proto := NewProtocol()
		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)
		reply := proto.Command(ParseCommand("EHLO localhost"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
	})
	Convey("HELO should work in ESTABLISH state", t, func() {
		proto := NewProtocol()
		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)
		reply := proto.Command(ParseCommand("HELO localhost"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
	})
	Convey("RSET should work in ESTABLISH state", t, func() {
		proto := NewProtocol()
		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)
		reply := proto.Command(ParseCommand("RSET"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
	})
	Convey("NOOP should work in ESTABLISH state", t, func() {
		proto := NewProtocol()
		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)
		reply := proto.Command(ParseCommand("NOOP"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
	})
	Convey("QUIT should work in ESTABLISH state", t, func() {
		proto := NewProtocol()
		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)
		reply := proto.Command(ParseCommand("QUIT"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 221)
	})
	Convey("MAIL shouldn't work in ESTABLISH state", t, func() {
		proto := NewProtocol()
		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)
		reply := proto.Command(ParseCommand("MAIL"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 500)
		So(reply.Lines(), ShouldResemble, []string{"500 Unrecognised command\n"})
	})
	Convey("RCPT shouldn't work in ESTABLISH state", t, func() {
		proto := NewProtocol()
		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)
		reply := proto.Command(ParseCommand("RCPT"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 500)
		So(reply.Lines(), ShouldResemble, []string{"500 Unrecognised command\n"})
	})
	Convey("DATA shouldn't work in ESTABLISH state", t, func() {
		proto := NewProtocol()
		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)
		reply := proto.Command(ParseCommand("DATA"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 500)
		So(reply.Lines(), ShouldResemble, []string{"500 Unrecognised command\n"})
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
	Convey("EHLO should work using Command", t, func() {
		proto := NewProtocol()
		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)
		So(proto.message.Helo, ShouldEqual, "")
		reply := proto.Command(ParseCommand("EHLO localhost"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(reply.Lines(), ShouldResemble, []string{"250-Hello localhost\n", "250-PIPELINING\n", "250 AUTH EXTERNAL CRAM-MD5 LOGIN PLAIN\n"})
		So(proto.state, ShouldEqual, MAIL)
		So(proto.message.Helo, ShouldEqual, "localhost")
	})
	Convey("HELO should work in MAIL state", t, func() {
		proto := NewProtocol()
		proto.Start()
		proto.Command(ParseCommand("HELO localhost"))
		So(proto.state, ShouldEqual, MAIL)
		So(proto.message.Helo, ShouldEqual, "localhost")
		reply := proto.Command(ParseCommand("EHLO localhost"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(reply.Lines(), ShouldResemble, []string{"250-Hello localhost\n", "250-PIPELINING\n", "250 AUTH EXTERNAL CRAM-MD5 LOGIN PLAIN\n"})
		So(proto.state, ShouldEqual, MAIL)
		So(proto.message.Helo, ShouldEqual, "localhost")
	})
	Convey("HELO should work in RCPT state", t, func() {
		proto := NewProtocol()
		proto.Start()
		proto.Command(ParseCommand("HELO localhost"))
		proto.Command(ParseCommand("MAIL From:<test>"))
		So(proto.state, ShouldEqual, RCPT)
		So(proto.message.Helo, ShouldEqual, "localhost")
		reply := proto.Command(ParseCommand("EHLO localhost"))
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
	Convey("HELO should work using Command", t, func() {
		proto := NewProtocol()
		proto.Start()
		So(proto.state, ShouldEqual, ESTABLISH)
		So(proto.message.Helo, ShouldEqual, "")
		reply := proto.Command(ParseCommand("HELO localhost"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(reply.Lines(), ShouldResemble, []string{"250 Hello localhost\n"})
		So(proto.state, ShouldEqual, MAIL)
		So(proto.message.Helo, ShouldEqual, "localhost")
	})
	Convey("HELO should work in MAIL state", t, func() {
		proto := NewProtocol()
		proto.Start()
		proto.Command(ParseCommand("HELO localhost"))
		So(proto.state, ShouldEqual, MAIL)
		So(proto.message.Helo, ShouldEqual, "localhost")
		reply := proto.Command(ParseCommand("HELO localhost"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(reply.Lines(), ShouldResemble, []string{"250 Hello localhost\n"})
		So(proto.state, ShouldEqual, MAIL)
		So(proto.message.Helo, ShouldEqual, "localhost")
	})
	Convey("HELO should work in RCPT state", t, func() {
		proto := NewProtocol()
		proto.Start()
		proto.Command(ParseCommand("HELO localhost"))
		proto.Command(ParseCommand("MAIL From:<test>"))
		So(proto.state, ShouldEqual, RCPT)
		So(proto.message.Helo, ShouldEqual, "localhost")
		reply := proto.Command(ParseCommand("HELO localhost"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(reply.Lines(), ShouldResemble, []string{"250 Hello localhost\n"})
		So(proto.state, ShouldEqual, MAIL)
		So(proto.message.Helo, ShouldEqual, "localhost")
	})
}

func TestDATA(t *testing.T) {
	Convey("DATA should accept data", t, func() {
		proto := NewProtocol()
		handlerCalled := false
		proto.MessageReceivedHandler = func(msg *data.Message) (string, error) {
			handlerCalled = true
			return "abc", nil
		}
		proto.Start()
		proto.HELO("localhost")
		proto.Command(ParseCommand("MAIL FROM:<test>"))
		proto.Command(ParseCommand("RCPT TO:<test>"))
		reply := proto.Command(ParseCommand("DATA"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 354)
		So(reply.Lines(), ShouldResemble, []string{"354 End data with <CR><LF>.<CR><LF>\n"})
		So(proto.state, ShouldEqual, DATA)
		reply = proto.ProcessData("Hi")
		So(reply, ShouldBeNil)
		So(proto.state, ShouldEqual, DATA)
		So(proto.message.Data, ShouldEqual, "Hi\n")
		reply = proto.ProcessData("How are you?")
		So(reply, ShouldBeNil)
		So(proto.state, ShouldEqual, DATA)
		So(proto.message.Data, ShouldEqual, "Hi\nHow are you?\n")
		reply = proto.ProcessData("\r\n.\r")
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(proto.state, ShouldEqual, MAIL)
		So(reply.Lines(), ShouldResemble, []string{"250 Ok: queued as abc\n"})
		So(handlerCalled, ShouldBeTrue)
	})
	Convey("Should return error if missing storage backend", t, func() {
		proto := NewProtocol()
		proto.Start()
		proto.HELO("localhost")
		proto.Command(ParseCommand("MAIL FROM:<test>"))
		proto.Command(ParseCommand("RCPT TO:<test>"))
		reply := proto.Command(ParseCommand("DATA"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 354)
		So(reply.Lines(), ShouldResemble, []string{"354 End data with <CR><LF>.<CR><LF>\n"})
		So(proto.state, ShouldEqual, DATA)
		reply = proto.ProcessData("Hi")
		So(reply, ShouldBeNil)
		So(proto.state, ShouldEqual, DATA)
		So(proto.message.Data, ShouldEqual, "Hi\n")
		reply = proto.ProcessData("How are you?")
		So(reply, ShouldBeNil)
		So(proto.state, ShouldEqual, DATA)
		So(proto.message.Data, ShouldEqual, "Hi\nHow are you?\n")
		reply = proto.ProcessData("\r\n.\r")
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 452)
		So(proto.state, ShouldEqual, MAIL)
		So(reply.Lines(), ShouldResemble, []string{"452 No storage backend\n"})
	})
	Convey("Should return error if storage backend fails", t, func() {
		proto := NewProtocol()
		handlerCalled := false
		proto.MessageReceivedHandler = func(msg *data.Message) (string, error) {
			handlerCalled = true
			return "", errors.New("abc")
		}
		proto.Start()
		proto.HELO("localhost")
		proto.Command(ParseCommand("MAIL FROM:<test>"))
		proto.Command(ParseCommand("RCPT TO:<test>"))
		reply := proto.Command(ParseCommand("DATA"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 354)
		So(reply.Lines(), ShouldResemble, []string{"354 End data with <CR><LF>.<CR><LF>\n"})
		So(proto.state, ShouldEqual, DATA)
		reply = proto.ProcessData("Hi")
		So(reply, ShouldBeNil)
		So(proto.state, ShouldEqual, DATA)
		So(proto.message.Data, ShouldEqual, "Hi\n")
		reply = proto.ProcessData("How are you?")
		So(reply, ShouldBeNil)
		So(proto.state, ShouldEqual, DATA)
		So(proto.message.Data, ShouldEqual, "Hi\nHow are you?\n")
		reply = proto.ProcessData("\r\n.\r")
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 452)
		So(proto.state, ShouldEqual, MAIL)
		So(reply.Lines(), ShouldResemble, []string{"452 Unable to store message\n"})
		So(handlerCalled, ShouldBeTrue)
	})
}

func TestRSET(t *testing.T) {
	Convey("RSET should reset the state correctly", t, func() {
		proto := NewProtocol()
		proto.Start()
		proto.HELO("localhost")
		proto.Command(ParseCommand("MAIL FROM:<test>"))
		proto.Command(ParseCommand("RCPT TO:<test>"))
		So(proto.state, ShouldEqual, RCPT)
		So(proto.message.From, ShouldEqual, "test")
		So(len(proto.message.To), ShouldEqual, 1)
		So(proto.message.To[0], ShouldEqual, "test")
		reply := proto.Command(ParseCommand("RSET"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(reply.Lines(), ShouldResemble, []string{"250 Ok\n"})
		So(proto.state, ShouldEqual, MAIL)
		So(proto.message.From, ShouldEqual, "")
		So(len(proto.message.To), ShouldEqual, 0)
	})
}

func TestNOOP(t *testing.T) {
	Convey("NOOP shouldn't modify the state", t, func() {
		proto := NewProtocol()
		proto.Start()
		proto.HELO("localhost")
		proto.Command(ParseCommand("MAIL FROM:<test>"))
		proto.Command(ParseCommand("RCPT TO:<test>"))
		So(proto.state, ShouldEqual, RCPT)
		So(proto.message.From, ShouldEqual, "test")
		So(len(proto.message.To), ShouldEqual, 1)
		So(proto.message.To[0], ShouldEqual, "test")
		reply := proto.Command(ParseCommand("NOOP"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(reply.Lines(), ShouldResemble, []string{"250 Ok\n"})
		So(proto.state, ShouldEqual, RCPT)
		So(proto.message.From, ShouldEqual, "test")
		So(len(proto.message.To), ShouldEqual, 1)
		So(proto.message.To[0], ShouldEqual, "test")
	})
}

func TestQUIT(t *testing.T) {
	Convey("QUIT should modify the state correctly", t, func() {
		proto := NewProtocol()
		proto.Start()
		reply := proto.Command(ParseCommand("QUIT"))
		So(proto.state, ShouldEqual, DONE)
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 221)
		So(reply.Lines(), ShouldResemble, []string{"221 Bye\n"})
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
	Convey("Error should be returned via Command", t, func() {
		proto := NewProtocol()
		proto.Start()
		proto.Command(ParseCommand("HELO localhost"))
		So(proto.state, ShouldEqual, MAIL)
		reply := proto.Command(ParseCommand("MAIL oink"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 550)
		So(reply.Lines(), ShouldResemble, []string{"550 Invalid syntax in MAIL command\n"})
		So(proto.state, ShouldEqual, MAIL)
	})
	Convey("ValidateSenderHandler should be called", t, func() {
		proto := NewProtocol()
		handlerCalled := false
		proto.ValidateSenderHandler = func(sender string) bool {
			handlerCalled = true
			So(sender, ShouldEqual, "oink@mailhog.example")
			return true
		}
		proto.Start()
		proto.Command(ParseCommand("HELO localhost"))
		So(proto.state, ShouldEqual, MAIL)
		reply := proto.Command(ParseCommand("MAIL From:<oink@mailhog.example>"))
		So(handlerCalled, ShouldBeTrue)
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(reply.Lines(), ShouldResemble, []string{"250 Sender oink@mailhog.example ok\n"})
		So(proto.state, ShouldEqual, RCPT)
	})
	Convey("ValidateSenderHandler errors should be returned", t, func() {
		proto := NewProtocol()
		handlerCalled := false
		proto.ValidateSenderHandler = func(sender string) bool {
			handlerCalled = true
			So(sender, ShouldEqual, "oink@mailhog.example")
			return false
		}
		proto.Start()
		proto.Command(ParseCommand("HELO localhost"))
		So(proto.state, ShouldEqual, MAIL)
		reply := proto.Command(ParseCommand("MAIL From:<oink@mailhog.example>"))
		So(handlerCalled, ShouldBeTrue)
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 550)
		So(reply.Lines(), ShouldResemble, []string{"550 Invalid sender oink@mailhog.example\n"})
		So(proto.state, ShouldEqual, MAIL)
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
	Convey("Error should be returned via Command", t, func() {
		proto := NewProtocol()
		proto.Start()
		proto.Command(ParseCommand("HELO localhost"))
		proto.Command(ParseCommand("MAIL FROM:<test>"))
		So(proto.state, ShouldEqual, RCPT)
		reply := proto.Command(ParseCommand("RCPT oink"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 550)
		So(reply.Lines(), ShouldResemble, []string{"550 Invalid syntax in RCPT command\n"})
		So(proto.state, ShouldEqual, RCPT)
	})
	Convey("ValidateRecipientHandler should be called", t, func() {
		proto := NewProtocol()
		handlerCalled := false
		proto.ValidateRecipientHandler = func(recipient string) bool {
			handlerCalled = true
			So(recipient, ShouldEqual, "oink@mailhog.example")
			return true
		}
		proto.Start()
		proto.Command(ParseCommand("HELO localhost"))
		proto.Command(ParseCommand("MAIL FROM:<test>"))
		So(proto.state, ShouldEqual, RCPT)
		reply := proto.Command(ParseCommand("RCPT To:<oink@mailhog.example>"))
		So(handlerCalled, ShouldBeTrue)
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(reply.Lines(), ShouldResemble, []string{"250 Recipient oink@mailhog.example ok\n"})
		So(proto.state, ShouldEqual, RCPT)
	})
	Convey("ValidateRecipientHandler errors should be returned", t, func() {
		proto := NewProtocol()
		handlerCalled := false
		proto.ValidateRecipientHandler = func(recipient string) bool {
			handlerCalled = true
			So(recipient, ShouldEqual, "oink@mailhog.example")
			return false
		}
		proto.Start()
		proto.Command(ParseCommand("HELO localhost"))
		proto.Command(ParseCommand("MAIL FROM:<test>"))
		So(proto.state, ShouldEqual, RCPT)
		reply := proto.Command(ParseCommand("RCPT To:<oink@mailhog.example>"))
		So(handlerCalled, ShouldBeTrue)
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 550)
		So(reply.Lines(), ShouldResemble, []string{"550 Invalid recipient oink@mailhog.example\n"})
		So(proto.state, ShouldEqual, RCPT)
	})
}

func TestAuth(t *testing.T) {
	Convey("AUTH should be listed in EHLO response", t, func() {
		proto := NewProtocol()
		proto.Start()
		reply := proto.Command(ParseCommand("EHLO localhost"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 250)
		So(reply.Lines(), ShouldResemble, []string{"250-Hello localhost\n", "250-PIPELINING\n", "250 AUTH EXTERNAL CRAM-MD5 LOGIN PLAIN\n"})
	})

	Convey("Invalid mechanism should be rejected", t, func() {
		proto := NewProtocol()
		proto.Start()
		proto.Command(ParseCommand("EHLO localhost"))
		reply := proto.Command(ParseCommand("AUTH OINK"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 504)
		So(reply.Lines(), ShouldResemble, []string{"504 Unsupported authentication mechanism\n"})
	})
}

func TestAuthExternal(t *testing.T) {
	Convey("AUTH EXTERNAL should call ValidateAuthenticationHandler", t, func() {
		proto := NewProtocol()
		handlerCalled := false
		proto.ValidateAuthenticationHandler = func(mechanism string, args ...string) (*Reply, bool) {
			handlerCalled = true
			So(mechanism, ShouldEqual, "EXTERNAL")
			So(len(args), ShouldEqual, 1)
			So(args[0], ShouldEqual, "oink!")
			return nil, true
		}
		proto.Start()
		proto.Command(ParseCommand("EHLO localhost"))
		reply := proto.Command(ParseCommand("AUTH EXTERNAL oink!"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 235)
		So(reply.Lines(), ShouldResemble, []string{"235 Authentication successful\n"})
		So(handlerCalled, ShouldBeTrue)
	})

	Convey("AUTH EXTERNAL ValidateAuthenticationHandler errors should be returned", t, func() {
		proto := NewProtocol()
		handlerCalled := false
		proto.ValidateAuthenticationHandler = func(mechanism string, args ...string) (*Reply, bool) {
			handlerCalled = true
			return ReplyError(errors.New("OINK :(")), false
		}
		proto.Start()
		proto.Command(ParseCommand("EHLO localhost"))
		reply := proto.Command(ParseCommand("AUTH EXTERNAL oink!"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 550)
		So(reply.Lines(), ShouldResemble, []string{"550 OINK :(\n"})
		So(handlerCalled, ShouldBeTrue)
	})
}

func TestAuthPlain(t *testing.T) {
	Convey("Inline AUTH PLAIN should call ValidateAuthenticationHandler", t, func() {
		proto := NewProtocol()
		handlerCalled := false
		proto.ValidateAuthenticationHandler = func(mechanism string, args ...string) (*Reply, bool) {
			handlerCalled = true
			So(mechanism, ShouldEqual, "PLAIN")
			So(len(args), ShouldEqual, 1)
			So(args[0], ShouldEqual, "oink!")
			return nil, true
		}
		proto.Start()
		proto.Command(ParseCommand("EHLO localhost"))
		reply := proto.Command(ParseCommand("AUTH PLAIN oink!"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 235)
		So(reply.Lines(), ShouldResemble, []string{"235 Authentication successful\n"})
		So(handlerCalled, ShouldBeTrue)
	})

	Convey("Inline AUTH PLAIN ValidateAuthenticationHandler errors should be returned", t, func() {
		proto := NewProtocol()
		handlerCalled := false
		proto.ValidateAuthenticationHandler = func(mechanism string, args ...string) (*Reply, bool) {
			handlerCalled = true
			return ReplyError(errors.New("OINK :(")), false
		}
		proto.Start()
		proto.Command(ParseCommand("EHLO localhost"))
		reply := proto.Command(ParseCommand("AUTH PLAIN oink!"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 550)
		So(reply.Lines(), ShouldResemble, []string{"550 OINK :(\n"})
		So(handlerCalled, ShouldBeTrue)
	})

	Convey("Two part AUTH PLAIN should call ValidateAuthenticationHandler", t, func() {
		proto := NewProtocol()
		handlerCalled := false
		proto.ValidateAuthenticationHandler = func(mechanism string, args ...string) (*Reply, bool) {
			handlerCalled = true
			So(mechanism, ShouldEqual, "PLAIN")
			So(len(args), ShouldEqual, 1)
			So(args[0], ShouldEqual, "oink!")
			return nil, true
		}
		proto.Start()
		proto.Command(ParseCommand("EHLO localhost"))
		reply := proto.Command(ParseCommand("AUTH PLAIN"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 334)
		So(reply.Lines(), ShouldResemble, []string{"334 \n"})

		_, reply = proto.Parse("oink!\n")
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 235)
		So(reply.Lines(), ShouldResemble, []string{"235 Authentication successful\n"})
		So(handlerCalled, ShouldBeTrue)
	})

	Convey("Two part AUTH PLAIN ValidateAuthenticationHandler errors should be returned", t, func() {
		proto := NewProtocol()
		handlerCalled := false
		proto.ValidateAuthenticationHandler = func(mechanism string, args ...string) (*Reply, bool) {
			handlerCalled = true
			return ReplyError(errors.New("OINK :(")), false
		}
		proto.Start()
		proto.Command(ParseCommand("EHLO localhost"))
		reply := proto.Command(ParseCommand("AUTH PLAIN"))
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 334)
		So(reply.Lines(), ShouldResemble, []string{"334 \n"})

		_, reply = proto.Parse("oink!\n")
		So(reply, ShouldNotBeNil)
		So(reply.Status, ShouldEqual, 550)
		So(reply.Lines(), ShouldResemble, []string{"550 OINK :(\n"})
		So(handlerCalled, ShouldBeTrue)
	})
}

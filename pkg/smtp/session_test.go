package smtp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/doctolib/MailHog/pkg/data"
	"github.com/doctolib/MailHog/pkg/storage"
)

type fakeRw struct {
	_read  func(p []byte) (n int, err error)
	_write func(p []byte) (n int, err error)
	_close func() error
}

func (rw *fakeRw) Read(p []byte) (n int, err error) {
	if rw._read != nil {
		return rw._read(p)
	}
	return 0, nil
}
func (rw *fakeRw) Close() error {
	if rw._close != nil {
		return rw._close()
	}
	return nil
}
func (rw *fakeRw) Write(p []byte) (n int, err error) {
	if rw._write != nil {
		return rw._write(p)
	}
	return len(p), nil
}

func TestAccept(t *testing.T) {
	Convey("Accept should handle a connection", t, func() {
		frw := &fakeRw{}
		mChan := make(chan *data.Message)
		Accept("1.1.1.1:11111", frw, storage.CreateInMemory(), mChan, "localhost", nil)
	})
}

func TestSocketError(t *testing.T) {
	Convey("Socket errors should return from Accept", t, func() {
		frw := &fakeRw{
			_read: func(p []byte) (n int, err error) {
				return -1, errors.New("OINK")
			},
		}
		mChan := make(chan *data.Message)
		Accept("1.1.1.1:11111", frw, storage.CreateInMemory(), mChan, "localhost", nil)
	})
}

func TestAcceptMessage(t *testing.T) {
	Convey("acceptMessage should be called", t, func() {
		serverReader, clientWritter := io.Pipe()
		clientReader, serverWritter := io.Pipe()

		serverConn := &fakeRw{
			_read: func(p []byte) (int, error) {
				n, err := serverReader.Read(p)
				return n, err
			},
			_write: func(p []byte) (int, error) {
				n, err := serverWritter.Write(p)
				return n, err
			},
			_close: func() error {
				var err error

				if rErr := serverReader.Close(); rErr != nil {
					err = rErr
				}
				if wErr := serverWritter.Close(); wErr != nil {
					err = wErr
				}

				return err
			},
		}

		mChan := make(chan *data.Message)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			Accept("1.1.1.1:11111", serverConn, storage.CreateInMemory(), mChan, "localhost", nil)
			wg.Done()
		}()

		clientBufReader := bufio.NewReader(clientReader)
		clientBufWriter := bufio.NewWriter(clientWritter)

		if msg, _, err := clientBufReader.ReadLine(); true {
			So(err, ShouldBeNil)
			So(string(msg), ShouldEqual, "220 localhost ESMTP MailHog")
		}

		if _, err := clientBufWriter.Write([]byte("EHLO localhost\r\n")); true {
			So(err, ShouldBeNil)
		}
		So(clientBufWriter.Flush(), ShouldBeNil)
		for _, expectedResponse := range []string{"250-Hello localhost", "250-PIPELINING", "250 AUTH PLAIN"} {
			if msg, _, err := clientBufReader.ReadLine(); true {
				So(err, ShouldBeNil)
				So(string(msg), ShouldEqual, expectedResponse)
			}
		}

		sender := "test@test.test"
		if _, err := clientBufWriter.Write([]byte(fmt.Sprintf("MAIL FROM:<%s>\r\n", sender))); true {
			So(err, ShouldBeNil)
		}
		So(clientBufWriter.Flush(), ShouldBeNil)
		if msg, _, err := clientBufReader.ReadLine(); true {
			So(err, ShouldBeNil)
			So(string(msg), ShouldEqual, fmt.Sprintf("250 Sender %s ok", sender))
		}

		recipient := "test@test.test"
		if _, err := clientBufWriter.Write([]byte(fmt.Sprintf("RCPT TO:<%s>\r\n", recipient))); true {
			So(err, ShouldBeNil)
		}
		So(clientBufWriter.Flush(), ShouldBeNil)
		if msg, _, err := clientBufReader.ReadLine(); true {
			So(err, ShouldBeNil)
			So(string(msg), ShouldEqual, fmt.Sprintf("250 Recipient %s ok", recipient))
		}

		data := "Hi.\r\n"
		endMark := "\r\n.\r\n"
		if _, err := clientBufWriter.Write([]byte("DATA\r\n")); true {
			So(err, ShouldBeNil)
		}
		So(clientBufWriter.Flush(), ShouldBeNil)
		if msg, _, err := clientBufReader.ReadLine(); true {
			So(err, ShouldBeNil)
			So(string(msg), ShouldEqual, "354 End data with <CR><LF>.<CR><LF>")
		}
		if _, err := clientBufWriter.Write([]byte(data)); true {
			So(err, ShouldBeNil)
		}
		if _, err := clientBufWriter.Write([]byte(endMark)); true {
			So(err, ShouldBeNil)
		}

		flushBeforeAcceptChan := make(chan error)
		go func() {
			flushBeforeAcceptChan <- clientBufWriter.Flush()
		}()

		flushed := false
		accepted := false
		for !flushed || !accepted {
			select {
			case err := <-flushBeforeAcceptChan:
				So(err, ShouldBeNil)
				flushed = true
			case m := <-mChan:
				So(accepted, ShouldBeFalse)
				So(m, ShouldNotBeNil)
				accepted = true
			}
		}

		if msg, _, err := clientBufReader.ReadLine(); true {
			So(err, ShouldBeNil)
			So(string(msg), ShouldStartWith, "250 Ok:")
		}

		if _, err := clientBufWriter.Write([]byte("QUIT\r\n")); true {
			So(err, ShouldBeNil)
		}
		So(clientBufWriter.Flush(), ShouldBeNil)
		if msg, _, err := clientBufReader.ReadLine(); true {
			So(err, ShouldBeNil)
			So(string(msg), ShouldEqual, "221 Bye")
		}
		if msg, _, err := clientBufReader.ReadLine(); true {
			So(err, ShouldEqual, io.EOF)
			So(string(msg), ShouldBeBlank)
		}

		wg.Wait()
	})
}

func TestValidateAuthentication(t *testing.T) {
	Convey("validateAuthentication is always successful", t, func() {
		c := &Session{}

		err, ok := c.validateAuthentication("OINK")
		So(err, ShouldBeNil)
		So(ok, ShouldBeTrue)

		err, ok = c.validateAuthentication("OINK", "arg1")
		So(err, ShouldBeNil)
		So(ok, ShouldBeTrue)

		err, ok = c.validateAuthentication("OINK", "arg1", "arg2")
		So(err, ShouldBeNil)
		So(ok, ShouldBeTrue)
	})
}

func TestValidateRecipient(t *testing.T) {
	Convey("validateRecipient is always successful", t, func() {
		c := &Session{}

		So(c.validateRecipient("OINK"), ShouldBeTrue)
		So(c.validateRecipient("foo@bar.mailhog"), ShouldBeTrue)
	})
}

func TestValidateSender(t *testing.T) {
	Convey("validateSender is always successful", t, func() {
		c := &Session{}

		So(c.validateSender("OINK"), ShouldBeTrue)
		So(c.validateSender("foo@bar.mailhog"), ShouldBeTrue)
	})
}

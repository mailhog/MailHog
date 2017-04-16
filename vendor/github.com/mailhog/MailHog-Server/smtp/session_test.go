package smtp

import (
	"errors"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/mailhog/data"
	"github.com/mailhog/storage"
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
		mbuf := "EHLO localhost\nMAIL FROM:<test>\nRCPT TO:<test>\nDATA\nHi.\r\n.\r\nQUIT\n"
		var rbuf []byte
		frw := &fakeRw{
			_read: func(p []byte) (n int, err error) {
				if len(p) >= len(mbuf) {
					ba := []byte(mbuf)
					mbuf = ""
					for i, b := range ba {
						p[i] = b
					}
					return len(ba), nil
				}

				ba := []byte(mbuf[0:len(p)])
				mbuf = mbuf[len(p):]
				for i, b := range ba {
					p[i] = b
				}
				return len(ba), nil
			},
			_write: func(p []byte) (n int, err error) {
				rbuf = append(rbuf, p...)
				return len(p), nil
			},
			_close: func() error {
				return nil
			},
		}
		mChan := make(chan *data.Message)
		var wg sync.WaitGroup
		wg.Add(1)
		handlerCalled := false
		go func() {
			handlerCalled = true
			<-mChan
			//FIXME breaks some tests (in drone.io)
			//m := <-mChan
			//So(m, ShouldNotBeNil)
			wg.Done()
		}()
		Accept("1.1.1.1:11111", frw, storage.CreateInMemory(), mChan, "localhost", nil)
		wg.Wait()
		So(handlerCalled, ShouldBeTrue)
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

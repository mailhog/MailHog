package smtp

import (
	"errors"
	"testing"

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
	frw := &fakeRw{}
	mChan := make(chan *data.Message)
	Accept("1.1.1.1:11111", frw, storage.CreateInMemory(), mChan, "localhost", nil)
}

func TestSocketError(t *testing.T) {
	frw := &fakeRw{
		_read: func(p []byte) (n int, err error) {
			return -1, errors.New("OINK")
		},
	}
	mChan := make(chan *data.Message)
	Accept("1.1.1.1:11111", frw, storage.CreateInMemory(), mChan, "localhost", nil)
}

func TestAcceptMessage(t *testing.T) {
	mbuf := "EHLO localhost\r\nMAIL FROM:<test>\r\nRCPT TO:<test>\r\nDATA\r\nHi.\r\n.\r\nQUIT\r\n"
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
	mChan := make(chan *data.Message, 1)
	Accept("1.1.1.1:11111", frw, storage.CreateInMemory(), mChan, "localhost", nil)
	if got := len(mChan); got != 1 {
		t.Fatalf("got %d accepted messages, want 1", got)
	}
	if msg := <-mChan; msg == nil {
		t.Fatal("accepted message is nil")
	}
	if len(rbuf) == 0 {
		t.Fatal("SMTP session wrote no replies")
	}
}

func TestValidateAuthentication(t *testing.T) {
	c := &Session{}

	err, ok := c.validateAuthentication("OINK")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected authentication to be valid")
	}

	err, ok = c.validateAuthentication("OINK", "arg1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected authentication with one arg to be valid")
	}

	err, ok = c.validateAuthentication("OINK", "arg1", "arg2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected authentication with two args to be valid")
	}
}

func TestValidateRecipient(t *testing.T) {
	c := &Session{}

	for _, recipient := range []string{"OINK", "foo@bar.mailhog"} {
		if !c.validateRecipient(recipient) {
			t.Fatalf("expected recipient %q to be valid", recipient)
		}
	}
}

func TestValidateSender(t *testing.T) {
	c := &Session{}

	for _, sender := range []string{"OINK", "foo@bar.mailhog"} {
		if !c.validateSender(sender) {
			t.Fatalf("expected sender %q to be valid", sender)
		}
	}
}

package protocol

// http://www.rfc-editor.org/rfc/rfc5321.txt

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestReply(t *testing.T) {
	Convey("Reply creates properly formatted responses", t, func() {
		r := &Reply{200, []string{}}
		l := r.Lines()
		So(l[0], ShouldEqual, "200\n")

		r = &Reply{200, []string{"Ok"}}
		l = r.Lines()
		So(l[0], ShouldEqual, "200 Ok\n")

		r = &Reply{200, []string{"Ok", "Still ok!"}}
		l = r.Lines()
		So(l[0], ShouldEqual, "200-Ok\n")
		So(l[1], ShouldEqual, "200 Still ok!\n")

		r = &Reply{200, []string{"Ok", "Still ok!", "OINK!"}}
		l = r.Lines()
		So(l[0], ShouldEqual, "200-Ok\n")
		So(l[1], ShouldEqual, "200-Still ok!\n")
		So(l[2], ShouldEqual, "200 OINK!\n")
	})
}

func TestBuiltInReplies(t *testing.T) {
	Convey("ReplyIdent is correct", t, func() {
		r := ReplyIdent("oink")
		So(r.Status, ShouldEqual, 220)
		So(len(r.lines), ShouldEqual, 1)
		So(r.lines[0], ShouldEqual, "oink")
	})

	Convey("ReplyBye is correct", t, func() {
		r := ReplyBye()
		So(r.Status, ShouldEqual, 221)
		So(len(r.lines), ShouldEqual, 1)
		So(r.lines[0], ShouldEqual, "Bye")
	})

	Convey("ReplyAuthOk is correct", t, func() {
		r := ReplyAuthOk()
		So(r.Status, ShouldEqual, 235)
		So(len(r.lines), ShouldEqual, 1)
		So(r.lines[0], ShouldEqual, "Authentication successful")
	})

	Convey("ReplyOk is correct", t, func() {
		r := ReplyOk()
		So(r.Status, ShouldEqual, 250)
		So(len(r.lines), ShouldEqual, 1)
		So(r.lines[0], ShouldEqual, "Ok")

		r = ReplyOk("oink")
		So(r.Status, ShouldEqual, 250)
		So(len(r.lines), ShouldEqual, 1)
		So(r.lines[0], ShouldEqual, "oink")

		r = ReplyOk("mailhog", "OINK!")
		So(r.Status, ShouldEqual, 250)
		So(len(r.lines), ShouldEqual, 2)
		So(r.lines[0], ShouldEqual, "mailhog")
		So(r.lines[1], ShouldEqual, "OINK!")
	})

	Convey("ReplySenderOk is correct", t, func() {
		r := ReplySenderOk("test")
		So(r.Status, ShouldEqual, 250)
		So(len(r.lines), ShouldEqual, 1)
		So(r.lines[0], ShouldEqual, "Sender test ok")
	})

	Convey("ReplyRecipientOk is correct", t, func() {
		r := ReplyRecipientOk("test")
		So(r.Status, ShouldEqual, 250)
		So(len(r.lines), ShouldEqual, 1)
		So(r.lines[0], ShouldEqual, "Recipient test ok")
	})

	Convey("ReplyAuthResponse is correct", t, func() {
		r := ReplyAuthResponse("test")
		So(r.Status, ShouldEqual, 334)
		So(len(r.lines), ShouldEqual, 1)
		So(r.lines[0], ShouldEqual, "test")
	})

	Convey("ReplyDataResponse is correct", t, func() {
		r := ReplyDataResponse()
		So(r.Status, ShouldEqual, 354)
		So(len(r.lines), ShouldEqual, 1)
		So(r.lines[0], ShouldEqual, "End data with <CR><LF>.<CR><LF>")
	})

	Convey("ReplyStorageFailed is correct", t, func() {
		r := ReplyStorageFailed("test")
		So(r.Status, ShouldEqual, 452)
		So(len(r.lines), ShouldEqual, 1)
		So(r.lines[0], ShouldEqual, "test")
	})

	Convey("ReplyUnrecognisedCommand is correct", t, func() {
		r := ReplyUnrecognisedCommand()
		So(r.Status, ShouldEqual, 500)
		So(len(r.lines), ShouldEqual, 1)
		So(r.lines[0], ShouldEqual, "Unrecognised command")
	})

	Convey("ReplyUnsupportedAuth is correct", t, func() {
		r := ReplyUnsupportedAuth()
		So(r.Status, ShouldEqual, 504)
		So(len(r.lines), ShouldEqual, 1)
		So(r.lines[0], ShouldEqual, "Unsupported authentication mechanism")
	})
}

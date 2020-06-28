package storage

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMaildirLoadStore(t *testing.T) {
	Convey("test memory message store-load", t, func() {
		maildir := CreateMaildir("")
		So(maildir, ShouldNotBeNil)
		defer os.RemoveAll(maildir.Path)

		testStoreLoad(maildir)
	})
}

func TestMaildirDelete(t *testing.T) {
	Convey("test memory message delete", t, func() {
		maildir := CreateMaildir("")
		So(maildir, ShouldNotBeNil)
		defer os.RemoveAll(maildir.Path)

		testDelete(maildir)
	})
}

func TestMaildirDeleteAll(t *testing.T) {
	Convey("test memory message delete-all", t, func() {
		maildir := CreateMaildir("")
		So(maildir, ShouldNotBeNil)
		defer os.RemoveAll(maildir.Path)

		testDeleteAll(maildir)
	})
}

func TestMaildirList(t *testing.T) {
	Convey("test memory message list", t, func() {
		maildir := CreateMaildir("")
		So(maildir, ShouldNotBeNil)
		defer os.RemoveAll(maildir.Path)

		testList(maildir)
	})
}

func TestMaildirSearch(t *testing.T) {
	Convey("test memory message search", t, func() {
		maildir := CreateMaildir("")
		So(maildir, ShouldNotBeNil)
		defer os.RemoveAll(maildir.Path)

		testSearch(maildir)
	})
}

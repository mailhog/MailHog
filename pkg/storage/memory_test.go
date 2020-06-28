package storage

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMemoryLoadStore(t *testing.T) {
	Convey("test memory message store-load", t, func() {
		s := Storage(CreateInMemory())
		So(s, ShouldNotBeNil)

		testStoreLoad(s)
	})
}

func TestMemoryDelete(t *testing.T) {
	Convey("test memory message delete", t, func() {
		s := Storage(CreateInMemory())
		So(s, ShouldNotBeNil)

		testDelete(s)
	})
}

func TestMemoryDeleteAll(t *testing.T) {
	Convey("test memory message delete-all", t, func() {
		s := Storage(CreateInMemory())
		So(s, ShouldNotBeNil)

		testDeleteAll(s)
	})
}

func TestMemoryList(t *testing.T) {
	Convey("test memory message list", t, func() {
		s := Storage(CreateInMemory())
		So(s, ShouldNotBeNil)

		testList(s)
	})
}

func TestMemorySearch(t *testing.T) {
	Convey("test memory message search", t, func() {
		s := Storage(CreateInMemory())
		So(s, ShouldNotBeNil)

		testSearch(s)
	})
}

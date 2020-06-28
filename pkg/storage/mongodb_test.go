package storage

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/google/uuid"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMongoDBLoadStore(t *testing.T) {
	Convey("test mongodb message store-load", t, func() {
		s := createTestMongoDBstorage(t)
		So(s, ShouldNotBeNil)
		defer cleanupMongoDB(s)

		testStoreLoad(s)
	})
}

func TestMongoDBDelete(t *testing.T) {
	Convey("test mongodb message delete", t, func() {
		s := createTestMongoDBstorage(t)
		So(s, ShouldNotBeNil)
		defer cleanupMongoDB(s)

		testDelete(s)
	})
}

func TestMongoDBDeleteAll(t *testing.T) {
	Convey("test mongodb message delete-all", t, func() {
		s := createTestMongoDBstorage(t)
		So(s, ShouldNotBeNil)
		defer cleanupMongoDB(s)

		testDeleteAll(s)
	})
}

func TestMongoDBList(t *testing.T) {
	Convey("test mongodb message list", t, func() {
		s := createTestMongoDBstorage(t)
		So(s, ShouldNotBeNil)
		defer cleanupMongoDB(s)

		testList(s)
	})
}

func TestMongoDBSearch(t *testing.T) {
	Convey("test mongodb message search", t, func() {
		s := createTestMongoDBstorage(t)
		So(s, ShouldNotBeNil)
		defer cleanupMongoDB(s)

		testSearch(s)
	})
}

func createTestMongoDBstorage(t *testing.T) (mongo *MongoDB) {
	uri := os.Getenv("TEST_MONGODB_URI")
	if uri == "" {
		t.SkipNow()
	}

	var db string
	if u, err := url.Parse(uri); err != nil {
		panic(err)
	} else {
		db = u.Path
		u.Path = ""
		uri = u.String()
	}

	coll := fmt.Sprintf("_%v", uuid.New())
	return CreateMongoDB(uri, db, coll)
}

func cleanupMongoDB(s *MongoDB) {
	if err := s.Collection.DropCollection(); err != nil {
		panic(err)
	}
	s.Session.Close()
}

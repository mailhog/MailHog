package storage

import (
	"context"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPostgreSQLLoadStore(t *testing.T) {
	Convey("test mongodb message store-load", t, func() {
		s := createTestPostgreSQLstorage(t)
		So(s, ShouldNotBeNil)
		defer cleanupPostgreSQL(s)

		testStoreLoad(s)
	})
}

func TestPostgreSQLDelete(t *testing.T) {
	Convey("test mongodb message delete", t, func() {
		s := createTestPostgreSQLstorage(t)
		So(s, ShouldNotBeNil)
		defer cleanupPostgreSQL(s)

		testDelete(s)
	})
}

func TestPostgreSQLDeleteAll(t *testing.T) {
	Convey("test mongodb message delete-all", t, func() {
		s := createTestPostgreSQLstorage(t)
		So(s, ShouldNotBeNil)
		defer cleanupPostgreSQL(s)

		testDeleteAll(s)
	})
}

func TestPostgreSQLList(t *testing.T) {
	Convey("test mongodb message list", t, func() {
		s := createTestPostgreSQLstorage(t)
		So(s, ShouldNotBeNil)
		defer cleanupPostgreSQL(s)

		testList(s)
	})
}

func TestPostgreSQLSearch(t *testing.T) {
	Convey("test mongodb message search", t, func() {
		s := createTestPostgreSQLstorage(t)
		So(s, ShouldNotBeNil)

		testSearch(s)
	})
}

func createTestPostgreSQLstorage(t *testing.T) (mongo *PostgreSQL) {
	uri := os.Getenv("TEST_POSTGRESQL_URI")
	if uri == "" {
		t.SkipNow()
	}

	return CreatePostgreSQL(uri)
}

func cleanupPostgreSQL(s *PostgreSQL) {
	s.Pool.Exec(context.Background(), "DROP TABLE IF EXISTS messages")
}

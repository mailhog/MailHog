package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMAILParsing(t *testing.T) {
	from, err := ParseMAIL("From:<foo@bar>")
	assert.Equal(t, from, "foo@bar")
	assert.Nil(t, err)

	from, err = ParseMAIL("From:<foo@bar.com>")
	assert.Equal(t, from, "foo@bar.com")
	assert.Nil(t, err)

	from, err = ParseMAIL("From:<foo>")
	assert.Equal(t, from, "foo")
	assert.Nil(t, err)

	from, err = ParseMAIL("To:<foo@bar>")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid sender")

	from, err = ParseMAIL("To:<foo@bar.com>")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid sender")

	from, err = ParseMAIL("To:<foo>")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid sender")

	from, err = ParseMAIL("INVALID")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid sender")

	from, err = ParseMAIL("From:INVALID")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid sender")

	from, err = ParseMAIL("From:foo")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid sender")

	from, err = ParseMAIL("From:foo@bar")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid sender")

	from, err = ParseMAIL("From: <foo@bar>")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid sender")
}

func TestRCPTParsing(t *testing.T) {
	from, err := ParseRCPT("To:<foo@bar>")
	assert.Equal(t, from, "foo@bar")
	assert.Nil(t, err)

	from, err = ParseRCPT("To:<foo@bar.com>")
	assert.Equal(t, from, "foo@bar.com")
	assert.Nil(t, err)

	from, err = ParseRCPT("To:<foo>")
	assert.Equal(t, from, "foo")
	assert.Nil(t, err)

	from, err = ParseRCPT("From:<foo@bar>")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid recipient")

	from, err = ParseRCPT("From:<foo@bar.com>")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid recipient")

	from, err = ParseRCPT("From:<foo>")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid recipient")

	from, err = ParseRCPT("INVALID")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid recipient")

	from, err = ParseRCPT("To:INVALID")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid recipient")

	from, err = ParseRCPT("To:foo")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid recipient")

	from, err = ParseRCPT("To:foo@bar")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid recipient")

	from, err = ParseRCPT("To: <foo@bar>")
	assert.Equal(t, from, "")
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "Invalid recipient")
}

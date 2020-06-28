package storage

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/doctolib/MailHog/pkg/data"
)

// testStoreLoad tests s.Store and s.Load
func testStoreLoad(s Storage) {
	Convey("test message store-load", func() {
		message := newSimpleMessage(
			"anoir.pattyre@delivery-swipe.test",
			"saddik.shipper@portapacket.test",
			"test subject",
			"So, how about that storage, eh?",
		)

		if id, err := s.Store(cloneMessage(message)); true {
			So(err, ShouldBeNil)
			So(data.MessageID(id), ShouldEqual, message.ID)
		}
		if loadedMessage, err := s.Load(string(message.ID)); true {
			So(err, ShouldBeNil)
			if messageString, err := json.MarshalIndent(message, "", "  "); err != nil {
				panic(err)
			} else {
				fmt.Printf("%s\n", messageString)
			}
			if messageString, err := json.MarshalIndent(loadedMessage, "", "  "); err != nil {
				panic(err)
			} else {
				fmt.Printf("%s\n", messageString)
			}
			So(loadedMessage, shouldResembleAsMessage, message)
		}
	})
}

// testDelete tests s.Store, s.Count, s.DeleteOne, and s.Load
func testDelete(s Storage) {
	Convey("test message delete", func() {
		messageToDelete := newSimpleMessage(
			"zoya.preskitt@mailient.test",
			"steeven.prest@mailsy.test",
			"Burn after reading",
			"Read this message once then delete it.",
		)
		messageToKeep := newSimpleMessage(
			"tanvir-tibbit@mailsy.test",
			"ram.gersain@mailient.test",
			"Please save this message",
			"Don't delete this message.",
		)

		So(s.Count(), ShouldEqual, 0)

		if id, err := s.Store(cloneMessage(messageToDelete)); true {
			So(err, ShouldBeNil)
			So(data.MessageID(id), ShouldEqual, messageToDelete.ID)
		}
		if id, err := s.Store(cloneMessage(messageToKeep)); true {
			So(err, ShouldBeNil)
			So(data.MessageID(id), ShouldEqual, messageToKeep.ID)
		}

		So(s.Count(), ShouldEqual, 2)

		if err := s.DeleteOne(string(messageToDelete.ID)); true {
			So(err, ShouldBeNil)
		}

		So(s.Count(), ShouldEqual, 1)

		// Should return nil as message has been deleted
		if loadedMessage, err := s.Load(string(messageToDelete.ID)); true {
			So(err, ShouldBeNil) // Not an error, just not present.
			So(loadedMessage, ShouldBeNil)
		}
		// This one should still be present.
		if loadedMessage, err := s.Load(string(messageToKeep.ID)); true {
			So(err, ShouldBeNil) // Not an error, just not present.
			So(loadedMessage, shouldResembleAsMessage, messageToKeep)
		}
	})
}

// testDeleteAll tests s.Store, s.Count, s.DeleteAll, and s.Load
func testDeleteAll(s Storage) {
	Convey("test message delete-all", func() {
		messages := []*data.Message{
			newSimpleMessage(
				"zoya.preskitt@delivery-junction.test",
				"steeven.prest@delivery-junction.test",
				"message 1",
				"first message",
			),
			newSimpleMessage(
				"tanvir-tibbit@mailsy.test",
				"ram.gersain@mailient.test",
				"message 2",
				"second message",
			),
			newSimpleMessage(
				"anoir.pattyre@delivery-swipe.test",
				"saddik.shipper@portapacket.test",
				"message 3",
				"third message",
			),
		}

		So(s.Count(), ShouldEqual, 0)

		for _, message := range messages {
			if id, err := s.Store(cloneMessage(message)); true {
				So(err, ShouldBeNil)
				So(data.MessageID(id), ShouldEqual, message.ID)
			}
		}

		So(s.Count(), ShouldEqual, len(messages))

		s.DeleteAll()

		So(s.Count(), ShouldEqual, 0)

		for _, message := range messages {
			// Should return nil as message has been deleted
			if loadedMessage, err := s.Load(string(message.ID)); true {
				So(err, ShouldBeNil) // Not an error, just not present.
				So(loadedMessage, ShouldBeNil)
			}
		}
	})
}

// testList tests s.Count, s.Store, and s.List
func testList(s Storage) {
	message := newSimpleMessage(
		"zoya.preskitt@delivery-junction.test",
		"steeven.prest@delivery-junction.test",
		"message to store and list",
		"a message",
	)

	So(s.Count(), ShouldEqual, 0)

	if listedMessages, err := s.List(0, 10); true {
		So(err, ShouldBeNil)
		So(listedMessages, ShouldNotBeNil)
		So(len(*listedMessages), ShouldEqual, 0)
	}

	if id, err := s.Store(cloneMessage(message)); true {
		So(err, ShouldBeNil)
		So(data.MessageID(id), ShouldEqual, message.ID)
	}

	if listedMessages, err := s.List(0, 10); true {
		So(err, ShouldBeNil)
		So(listedMessages, ShouldNotBeNil)
		So(len(*listedMessages), ShouldEqual, 1)
		So(&(*listedMessages)[0], shouldResembleAsMessage, message)
	}
}

// testSearch tests s.Count, s.Store, and s.Search (first page only)
func testSearch(s Storage) {
	Convey("test message search", func() {
		messages := []*data.Message{
			newSimpleMessage(
				"zoya.preskitt@delivery-junction.test",
				"steeven.prest@delivery-junction.test",
				"message 1",
				"first message",
			),
			newSimpleMessage(
				"tanvir-tibbit@mailsy.test",
				"steeven.prest@delivery-junction.test",
				"message 1",
				"another message to steeven.preset",
			),
			newSimpleMessage(
				"ram.gersain@mailient.test",
				"tanvir-tibbit@mailsy.test",
				"message 2",
				"second message",
			),
			newSimpleMessage(
				"anoir.pattyre@delivery-swipe.test",
				"saddik.shipper@portapacket.test",
				"message 3",
				"third message",
			),
		}

		So(s.Count(), ShouldEqual, 0)

		for _, message := range messages {
			if id, err := s.Store(cloneMessage(message)); true {
				So(err, ShouldBeNil)
				So(data.MessageID(id), ShouldEqual, message.ID)
			}
		}

		searches := []struct {
			kind    string
			query   string
			matches data.Messages
		}{
			{
				kind:    "to",
				query:   "nonexistant@test.test",
				matches: data.Messages([]data.Message{}),
			},
			{
				kind:  "to",
				query: pathToEmail(messages[3].To[0]),
				matches: data.Messages([]data.Message{
					*messages[3],
				}),
			},
			{
				kind:  "to",
				query: pathToEmail(messages[0].To[0]),
				matches: data.Messages([]data.Message{
					*messages[0],
					*messages[1],
				}),
			},
			{
				kind:  "from",
				query: pathToEmail(messages[1].From),
				matches: data.Messages([]data.Message{
					*messages[1],
					// messages[2]'s recipeient is the same as the sender for messages[1]
				}),
			},
			{
				kind:  "from",
				query: pathToEmail(messages[2].From),
				matches: data.Messages([]data.Message{
					*messages[2],
					// messages[1]'s sender is the same as the recipient for messages[2]
				}),
			},
			{
				kind:  "containing",
				query: "message",
				matches: data.Messages([]data.Message{
					*messages[0],
					*messages[1],
					*messages[2],
					*messages[3],
				}),
			},
			{
				kind:    "containing",
				query:   "textnotfound",
				matches: data.Messages([]data.Message{}),
			},
			{
				kind:  "containing",
				query: "third",
				matches: data.Messages([]data.Message{
					*messages[3],
				}),
			},
		}
		for _, search := range searches {
			if matchedMessages, i, err := s.Search(search.kind, search.query, 0, 10); true {
				So(err, ShouldBeNil) // Not an error, just not present.
				So(i, ShouldEqual, len(search.matches))
				So(len(*matchedMessages), ShouldEqual, i)
				for _, expectedMatch := range search.matches {
					So(matchedMessages, shouldContainAsMessage, &expectedMatch)
				}
			}
		}
	})
}

// cloneMessage produces a deep copy of message
func cloneMessage(message *data.Message) *data.Message {
	clone := &data.Message{}
	if d, err := json.Marshal(message); err != nil {
		panic(err)
	} else if err := json.Unmarshal(d, clone); err != nil {
		panic(err)
	}

	return clone
}

// pathToEmail converts a data.Path to an email address
func pathToEmail(path *data.Path) string {
	return fmt.Sprintf("%s@%s", path.Mailbox, path.Domain)
}

// newSimpleMessage generates a new message without MIME parts, unneeded headers, mail relays, or multiple recipients
func newSimpleMessage(sender, recipient, subject, content string) *data.Message {
	return (&data.SMTPMessage{
		Helo: "localhost", // Domain part of sender email
		From: sender,
		To:   []string{recipient},
		Data: strings.Join([]string{
			fmt.Sprintf("From: %s", sender),
			fmt.Sprintf("To: %s", recipient),
			fmt.Sprintf("Subject: %s", subject),
			"",
			content,
		}, "\r\n"),
	}).Parse("mailhog.example")
}

// shouldResembleAsMessage tests if actual and expected are data.Messages of the same content
func shouldResembleAsMessage(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return "Wrong number of arguments, need exactly one expected message"
	}

	var actualMessage, expectedMessage *data.Message
	if asMessage, ok := actual.(*data.Message); ok {
		actualMessage = cloneMessage(asMessage)
	} else {
		return fmt.Sprintf("Actual (type %s) is not of type *data.Message", reflect.TypeOf(actual).String())
	}
	if asMessage, ok := expected[0].(*data.Message); ok {
		expectedMessage = cloneMessage(asMessage)
	} else {
		return fmt.Sprintf("Expected (type %s) is not of type *data.Message", reflect.TypeOf(expected).String())
	}

	actualMessage.Created = time.Time{}
	expectedMessage.Created = time.Time{}

	// TODO: create a Marshal/Unmarshal function for messages to allow the Maildir driver to store and load without
	// modifying the message.
	for _, header := range []string{"Received", "Return-Path", "Message-ID"} {
		delete(actualMessage.Content.Headers, header)
		delete(expectedMessage.Content.Headers, header)
	}

	if result := ShouldResemble(*actualMessage, *expectedMessage); result == "" {
		return ""
	} else if actualMessageJson, err := json.MarshalIndent(actualMessage, "", "  "); err != nil {
		panic(err)
	} else if expectedMessageJson, err := json.MarshalIndent(expectedMessage, "", "  "); err != nil {
		panic(err)
	} else {
		return fmt.Sprintf(
			"ACTUAL:\n%s\n\nEXPECTED:\n%s\n\nACTUAL was expected to resmeble EXPECTED (but it didn't!)",
			actualMessageJson,
			expectedMessageJson,
		)
	}
}

// shouldContainAsMessage tests if actual contains a list of messagse at least one of which are expected
func shouldContainAsMessage(actual interface{}, expected ...interface{}) string {
	var actualMessages *data.Messages
	var expectedMessage *data.Message
	if asMessages, ok := actual.(*data.Messages); ok {
		actualMessages = asMessages
	} else {
		return fmt.Sprintf("Actual (type %s) is not of type *data.Messages", reflect.TypeOf(actual).String())
	}
	if asMessage, ok := expected[0].(*data.Message); ok {
		expectedMessage = asMessage
	} else {
		return fmt.Sprintf("Expected (type %s) is not of type *data.Message", reflect.TypeOf(expected).String())
	}

	for _, actualMessage := range *actualMessages {
		if result := shouldResembleAsMessage(&actualMessage, expectedMessage); result == "" {
			return ""
		}
	}

	if actualMessagesJson, err := json.MarshalIndent(actualMessages, "", "  "); err != nil {
		panic(err)
	} else if expectedMessageJson, err := json.MarshalIndent(expectedMessage, "", "  "); err != nil {
		panic(err)
	} else {
		return fmt.Sprintf("ACTUAL:\n%s\n\nEXPECTED:\n%s\n\nACTUAL was expected to contain EXPECTED (but it didn't!)", actualMessagesJson, expectedMessageJson)
	}
}

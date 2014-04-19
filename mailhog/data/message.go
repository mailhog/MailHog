package data;

import (
	"log"
	"strings"
	"time"
    "labix.org/v2/mgo/bson"
    "github.com/ian-kent/MailHog/mailhog"
)

type Message struct {
	Id string
	From *Path
	To []*Path
	Content *Content
	Created time.Time
}

type Path struct {
	Relays []string
	Mailbox string
	Domain string
	Params string
}

type Content struct {
	Headers map[string][]string
	Body string
	Size int
}

type SMTPMessage struct {
	From string
	To []string
	Data string
	Helo string
}

func ParseSMTPMessage(c *mailhog.Config, m *SMTPMessage) *Message {
	arr := make([]*Path, 0)
	for _, path := range m.To {
		arr = append(arr, PathFromString(path))
	}
	msg := &Message{
		Id: bson.NewObjectId().Hex(),
		From: PathFromString(m.From),
		To: arr,
		Content: ContentFromString(m.Data),
		Created: time.Now(),
	}
	msg.Content.Headers["Message-ID"] = []string{msg.Id + "@" + c.Hostname} // FIXME
	msg.Content.Headers["Received"] = []string{"from " + m.Helo + " by " + c.Hostname + " (Go-MailHog)"} // FIXME
	msg.Content.Headers["Return-Path"] = []string{"<" + m.From + ">"}
	return msg
}

func PathFromString(path string) *Path {
	relays := make([]string, 0)
	email := path
	if(strings.Contains(path, ":")) {
		x := strings.SplitN(path, ":", 2)
		r, e := x[0], x[1]
		email = e
		relays = strings.Split(r, ",")
	}
	mailbox, domain := "", ""
	if(strings.Contains(email, "@")) {
		x := strings.SplitN(email, "@", 2)
		mailbox, domain = x[0], x[1]
	} else {
		mailbox = email
	}

	return &Path{
		Relays: relays,
		Mailbox: mailbox,
		Domain: domain,
		Params: "", // FIXME?
	}
}

func ContentFromString(data string) *Content {
	x := strings.Split(data, "\r\n\r\n")	
	headers, body := x[0], x[1]

	h := make(map[string][]string, 0)
	hdrs := strings.Split(headers, "\r\n")
	for _, hdr := range hdrs {
		if(strings.Contains(hdr, ": ")) {
			y := strings.SplitN(hdr, ": ", 2)
			key, value := y[0], y[1]
			h[key] = []string{value}
		} else {
			log.Printf("Found invalid header: '%s'", hdr)
		}
	}

	return &Content{
		Size: len(data),
		Headers: h,
		Body: body,
	}
}

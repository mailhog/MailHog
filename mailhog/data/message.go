package data;

import (
	"log"
	"strings"
	"time"
	"regexp"
    "labix.org/v2/mgo/bson"
    "github.com/ian-kent/MailHog/mailhog"
)

type Messages []Message

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

type MIMEBody struct {
	Parts []Content
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
	log.Printf("Is MIME: %t\n", msg.Content.IsMIME());
	msg.Content.Headers["Message-ID"] = []string{msg.Id + "@" + c.Hostname}
	msg.Content.Headers["Received"] = []string{"from " + m.Helo + " by " + c.Hostname + " (Go-MailHog)\r\n          id " + msg.Id + "@" + c.Hostname + "; " + time.Now().Format(time.RFC1123Z)}
	msg.Content.Headers["Return-Path"] = []string{"<" + m.From + ">"}
	return msg
}

func (content *Content) IsMIME() bool {
	if strings.HasPrefix(content.Headers["Content-Type"][0], "multipart/") {
		return true
	} else {
		return false
	}
}

func (content *Content) ParseMIMEBody() *MIMEBody {
	re := regexp.MustCompile("boundary=\"([^\"]+)\"")
	match := re.FindStringSubmatch(content.Body)
	log.Printf("Got boundary: %s", match[1])
	return nil
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
	x := strings.SplitN(data, "\r\n\r\n", 2)
	headers, body := x[0], x[1]

	h := make(map[string][]string, 0)
	hdrs := strings.Split(headers, "\r\n")
	for _, hdr := range hdrs {
		if(strings.Contains(hdr, ": ")) {
			y := strings.SplitN(hdr, ": ", 2)
			key, value := y[0], y[1]
			// TODO multiple header fields
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

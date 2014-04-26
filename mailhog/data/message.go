package data;

import (
	"log"
	"strings"
	"time"
	"regexp"
    "labix.org/v2/mgo/bson"
)

type Messages []Message

type Message struct {
	Id string
	From *Path
	To []*Path
	Content *Content
	Created time.Time
	MIME *MIMEBody
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
	Parts []*Content
}

func ParseSMTPMessage(m *SMTPMessage, hostname string) *Message {
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

	if msg.Content.IsMIME() {
		log.Printf("Parsing MIME body")
		msg.MIME = msg.Content.ParseMIMEBody()
	}
	
	msg.Content.Headers["Message-ID"] = []string{msg.Id + "@" + hostname}
	msg.Content.Headers["Received"] = []string{"from " + m.Helo + " by " + hostname + " (Go-MailHog)\r\n          id " + msg.Id + "@" + hostname + "; " + time.Now().Format(time.RFC1123Z)}
	msg.Content.Headers["Return-Path"] = []string{"<" + m.From + ">"}
	return msg
}

func (content *Content) IsMIME() bool {
	return strings.HasPrefix(content.Headers["Content-Type"][0], "multipart/")
}

func (content *Content) ParseMIMEBody() *MIMEBody {
	re := regexp.MustCompile("boundary=\"([^\"]+)\"")
	match := re.FindStringSubmatch(content.Headers["Content-Type"][0])
	log.Printf("Got boundary: %s", match[1])

	p := strings.Split(content.Body, "--" + match[1])
	parts := make([]*Content, 0)
	for m := range p {
		if len(p[m]) > 0 {
			parts = append(parts, ContentFromString(strings.Trim(p[m], "\r\n")))
		}
	}

	return &MIMEBody{
		Parts: parts,
	}
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
	log.Printf("Parsing Content from string: '%s'", data)
	x := strings.SplitN(data, "\r\n\r\n", 2)
	h := make(map[string][]string, 0)

	if len(x) == 2 {
		headers, body := x[0], x[1]
		hdrs := strings.Split(headers, "\r\n")
		var lastHdr = ""
		for _, hdr := range hdrs {
			if lastHdr != "" && strings.HasPrefix(hdr, " ") {
				h[lastHdr][len(h[lastHdr])-1] = h[lastHdr][len(h[lastHdr])-1] + hdr
			} else if strings.Contains(hdr, ": ") {
				y := strings.SplitN(hdr, ": ", 2)
				key, value := y[0], y[1]
				// TODO multiple header fields
				h[key] = []string{value}
				lastHdr = key
			} else {
				log.Printf("Found invalid header: '%s'", hdr)
			}
		}
		return &Content{
			Size: len(data),
			Headers: h,
			Body: body,
		}
	} else {
		return &Content{
			Size: len(data),
			Headers: h,
			Body: x[0],
		}
	}
}

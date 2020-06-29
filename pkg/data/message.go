package data

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io"
	"mime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// MessageID represents the ID of an SMTP message including the hostname part
type MessageID string

// NewMessageID generates a new message ID
func NewMessageID(hostname string) (MessageID, error) {
	size := 32

	rb := make([]byte, size)
	_, err := rand.Read(rb)

	if err != nil {
		return MessageID(""), err
	}

	rs := base64.URLEncoding.EncodeToString(rb)

	return MessageID(rs + "@" + hostname), nil
}

// Messages represents an array of Messages
// - TODO is this even required?
type Messages []Message

// Message represents a parsed SMTP message
type Message struct {
	ID      MessageID
	From    *Path
	To      []*Path
	Content *Content
	Created time.Time
	MIME    *MIMEBody // FIXME refactor to use Content.MIME
	Raw     *SMTPMessage
}

// Path represents an SMTP forward-path or return-path
type Path struct {
	Relays  []string
	Mailbox string
	Domain  string
	Params  string
}

// Content represents the body content of an SMTP message
type Content struct {
	Headers map[string][]string
	Body    string
	Size    int
	MIME    *MIMEBody
}

// SMTPMessage represents a raw SMTP message
type SMTPMessage struct {
	From string
	To   []string
	Data string
	Helo string
}

// MIMEBody represents a collection of MIME parts
type MIMEBody struct {
	Parts []*Content
}

// Parse converts a raw SMTP message to a parsed MIME message
func (m *SMTPMessage) Parse(hostname string) *Message {
	arr := make([]*Path, 0)
	for _, path := range m.To {
		arr = append(arr, PathFromString(path))
	}

	id, _ := NewMessageID(hostname)
	msg := &Message{
		ID:      id,
		From:    PathFromString(m.From),
		To:      arr,
		Content: ContentFromString(m.Data),
		Created: time.Now(),
		Raw:     m,
	}

	if msg.Content.IsMIME() {
		log.Trace("Parsing MIME body")
		msg.MIME = msg.Content.ParseMIMEBody()
	}

	// TODO: Don't inject this stuff as part of the parsing function, we need a proper Marshal/Unmarshal.

	// find headers
	var hasMessageID bool
	var receivedHeaderName string
	var returnPathHeaderName string

	for k := range msg.Content.Headers {
		if strings.ToLower(k) == "message-id" {
			hasMessageID = true
			continue
		}
		if strings.ToLower(k) == "received" {
			receivedHeaderName = k
			continue
		}
		if strings.ToLower(k) == "return-path" {
			returnPathHeaderName = k
			continue
		}
	}

	if !hasMessageID {
		msg.Content.Headers["Message-ID"] = []string{string(id)}
	}

	if len(receivedHeaderName) > 0 {
		msg.Content.Headers[receivedHeaderName] = append(msg.Content.Headers[receivedHeaderName], "from "+m.Helo+" by "+hostname+" (MailHog)\r\n          id "+string(id)+"; "+time.Now().Format(time.RFC1123Z))
	} else {
		msg.Content.Headers["Received"] = []string{"from " + m.Helo + " by " + hostname + " (MailHog)\r\n          id " + string(id) + "; " + time.Now().Format(time.RFC1123Z)}
	}

	if len(returnPathHeaderName) > 0 {
		msg.Content.Headers[returnPathHeaderName] = append(msg.Content.Headers[returnPathHeaderName], "<"+m.From+">")
	} else {
		msg.Content.Headers["Return-Path"] = []string{"<" + m.From + ">"}
	}

	return msg
}

// Bytes returns an io.Reader containing the raw message data
func (m *SMTPMessage) Bytes() io.Reader {
	var b = new(bytes.Buffer)

	b.WriteString("HELO:<" + m.Helo + ">\r\n")
	b.WriteString("FROM:<" + m.From + ">\r\n")
	for _, t := range m.To {
		b.WriteString("TO:<" + t + ">\r\n")
	}
	b.WriteString("\r\n")
	b.WriteString(m.Data)

	return b
}

// FromBytes returns a SMTPMessage from raw message bytes (as output by SMTPMessage.Bytes())
func FromBytes(b []byte) *SMTPMessage {
	msg := &SMTPMessage{}
	var headerDone bool
	r := strings.NewReader(string(b))
	rb := bufio.NewReader(r)

	done := false
	for !done {
		l, err := rb.ReadString('\n')
		if err == io.EOF {
			done = true
		} else if err != nil {
			panic(err)
		}

		if !headerDone {
			switch {
			case strings.HasPrefix(l, "HELO:<"):
				l = strings.TrimPrefix(l, "HELO:<")
				l = strings.TrimSuffix(l, ">\r\n")
				msg.Helo = l
				continue
			case strings.HasPrefix(l, "FROM:<"):
				l = strings.TrimPrefix(l, "FROM:<")
				l = strings.TrimSuffix(l, ">\r\n")
				msg.From = l
				continue
			case strings.HasPrefix(l, "TO:<"):
				l = strings.TrimPrefix(l, "TO:<")
				l = strings.TrimSuffix(l, ">\r\n")
				msg.To = append(msg.To, l)
				continue
			case l == "\r\n":
				headerDone = true
				continue
			}
		}
		msg.Data += l
	}
	return msg
}

// Bytes returns an io.Reader containing the raw message data
func (m *Message) Bytes() io.Reader {
	var b = new(bytes.Buffer)

	for k, vs := range m.Content.Headers {
		for _, v := range vs {
			b.WriteString(k + ": " + v + "\r\n")
		}
	}

	b.WriteString("\r\n")
	b.WriteString(m.Content.Body)

	return b
}

// IsMIME detects a valid MIME header
func (content *Content) IsMIME() bool {
	header, ok := content.Headers["Content-Type"]
	if !ok {
		return false
	}
	return strings.HasPrefix(header[0], "multipart/")
}

// ParseMIMEBody parses SMTP message content into multiple MIME parts
func (content *Content) ParseMIMEBody() *MIMEBody {
	var parts []*Content

	if hdr, ok := content.Headers["Content-Type"]; ok {
		if len(hdr) > 0 {
			boundary := extractBoundary(hdr[0])
			var p []string
			if len(boundary) > 0 {
				p = strings.Split(content.Body, "--"+boundary)
				log.Tracef("Got boundary: %s", boundary)
			} else {
				log.Tracef("Boundary not found: %s", hdr[0])
			}

			for _, s := range p {
				if len(s) > 0 {
					part := ContentFromString(strings.Trim(s, "\r\n"))
					if part.IsMIME() {
						log.Trace("Parsing inner MIME body")
						part.MIME = part.ParseMIMEBody()
					}
					parts = append(parts, part)
				}
			}
		}
	}

	return &MIMEBody{
		Parts: parts,
	}
}

// PathFromString parses a forward-path or reverse-path into its parts
func PathFromString(path string) *Path {
	var relays []string
	email := path
	if strings.Contains(path, ":") {
		x := strings.SplitN(path, ":", 2)
		r, e := x[0], x[1]
		email = e
		relays = strings.Split(r, ",")
	}
	mailbox, domain := "", ""
	if strings.Contains(email, "@") {
		x := strings.SplitN(email, "@", 2)
		mailbox, domain = x[0], x[1]
	} else {
		mailbox = email
	}

	return &Path{
		Relays:  relays,
		Mailbox: mailbox,
		Domain:  domain,
	}
}

// ContentFromString parses SMTP content into separate headers and body
func ContentFromString(data string) *Content {
	log.Tracef("Parsing Content from string: '%s'", data)
	x := strings.SplitN(data, "\r\n\r\n", 2)
	h := make(map[string][]string)

	// FIXME this fails if the message content has no headers - specifically,
	// if it doesn't contain \r\n\r\n

	if len(x) == 2 {
		headers, body := x[0], x[1]
		hdrs := strings.Split(headers, "\r\n")
		var lastHdr = ""
		for _, hdr := range hdrs {
			switch {
			case lastHdr != "" && (strings.HasPrefix(hdr, " ") || strings.HasPrefix(hdr, "\t")):
				h[lastHdr][len(h[lastHdr])-1] = h[lastHdr][len(h[lastHdr])-1] + hdr
			case strings.Contains(hdr, ": "):
				y := strings.SplitN(hdr, ": ", 2)
				key, value := y[0], y[1]
				// TODO multiple header fields
				h[key] = []string{value}
				lastHdr = key
			case len(hdr) > 0:
				log.Warnf("Found invalid header: '%s'", hdr)
			}
		}
		return &Content{
			Size:    len(data),
			Headers: h,
			Body:    body,
		}
	}
	return &Content{
		Size:    len(data),
		Headers: h,
		Body:    x[0],
	}
}

// extractBoundary extract boundary string in contentType.
// It returns empty string if no valid boundary found
func extractBoundary(contentType string) string {
	_, params, err := mime.ParseMediaType(contentType)
	if err == nil {
		return params["boundary"]
	}
	return ""
}

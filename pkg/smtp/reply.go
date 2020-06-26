package smtp

import "strconv"

// http://www.rfc-editor.org/rfc/rfc5321.txt

// Reply is a struct representing an SMTP reply (status code + lines)
type Reply struct {
	Status int
	lines  []string
	Done   func()
}

// Lines returns the formatted SMTP reply
func (r Reply) Lines() []string {
	var lines []string

	if len(r.lines) == 0 {
		l := strconv.Itoa(r.Status)
		lines = append(lines, l+"\n")
		return lines
	}

	for i, line := range r.lines {
		l := ""
		if i == len(r.lines)-1 {
			l = strconv.Itoa(r.Status) + " " + line + "\r\n"
		} else {
			l = strconv.Itoa(r.Status) + "-" + line + "\r\n"
		}
		lines = append(lines, l)
	}

	return lines
}

// ReplyIdent creates a 220 welcome reply
func ReplyIdent(ident string) *Reply { return &Reply{220, []string{ident}, nil} }

// ReplyReadyToStartTLS creates a 220 ready to start TLS reply
func ReplyReadyToStartTLS(callback func()) *Reply {
	return &Reply{220, []string{"Ready to start TLS"}, callback}
}

// ReplyBye creates a 221 Bye reply
func ReplyBye() *Reply { return &Reply{221, []string{"Bye"}, nil} }

// ReplyAuthOk creates a 235 authentication successful reply
func ReplyAuthOk() *Reply { return &Reply{235, []string{"Authentication successful"}, nil} }

// ReplyOk creates a 250 Ok reply
func ReplyOk(message ...string) *Reply {
	if len(message) == 0 {
		message = []string{"Ok"}
	}
	return &Reply{250, message, nil}
}

// ReplySenderOk creates a 250 Sender ok reply
func ReplySenderOk(sender string) *Reply {
	return &Reply{250, []string{"Sender " + sender + " ok"}, nil}
}

// ReplyRecipientOk creates a 250 Sender ok reply
func ReplyRecipientOk(recipient string) *Reply {
	return &Reply{250, []string{"Recipient " + recipient + " ok"}, nil}
}

// ReplyAuthResponse creates a 334 authentication reply
func ReplyAuthResponse(response string) *Reply { return &Reply{334, []string{response}, nil} }

// ReplyDataResponse creates a 354 data reply
func ReplyDataResponse() *Reply { return &Reply{354, []string{"End data with <CR><LF>.<CR><LF>"}, nil} }

// ReplyStorageFailed creates a 452 error reply
func ReplyStorageFailed(reason string) *Reply { return &Reply{452, []string{reason}, nil} }

// ReplyUnrecognisedCommand creates a 500 Unrecognised command reply
func ReplyUnrecognisedCommand() *Reply { return &Reply{500, []string{"Unrecognised command"}, nil} }

// ReplyLineTooLong creates a 500 Line too long reply
func ReplyLineTooLong() *Reply { return &Reply{500, []string{"Line too long"}, nil} }

// ReplySyntaxError creates a 501 Syntax error reply
func ReplySyntaxError(response string) *Reply {
	if len(response) > 0 {
		response = " (" + response + ")"
	}
	return &Reply{501, []string{"Syntax error" + response}, nil}
}

// ReplyUnsupportedAuth creates a 504 unsupported authentication reply
func ReplyUnsupportedAuth() *Reply {
	return &Reply{504, []string{"Unsupported authentication mechanism"}, nil}
}

// ReplyMustIssueSTARTTLSFirst creates a 530 reply for RFC3207
func ReplyMustIssueSTARTTLSFirst() *Reply {
	return &Reply{530, []string{"Must issue a STARTTLS command first"}, nil}
}

// ReplyInvalidAuth creates a 535 error reply
func ReplyInvalidAuth() *Reply {
	return &Reply{535, []string{"Authentication credentials invalid"}, nil}
}

// ReplyError creates a 500 error reply
func ReplyError(err error) *Reply { return &Reply{550, []string{err.Error()}, nil} }

// ReplyTooManyRecipients creates a 552 too many recipients reply
func ReplyTooManyRecipients() *Reply { return &Reply{552, []string{"Too many recipients"}, nil} }

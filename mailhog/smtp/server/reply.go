package smtp

// http://www.rfc-editor.org/rfc/rfc5321.txt

// Reply is a struct representing an SMTP reply (status code + lines)
type Reply struct {
	status int
	lines  []string
}

// ReplyIdent creates a 220 welcome reply
func ReplyIdent(ident string) *Reply { return &Reply{220, []string{ident}} }

// ReplyBye creates a 221 Bye reply
func ReplyBye() *Reply { return &Reply{221, []string{"Bye"}} }

// ReplyAuthOk creates a 235 authentication successful reply
func ReplyAuthOk() *Reply { return &Reply{235, []string{"Authentication successful"}} }

// ReplyOk creates a 250 Ok reply
func ReplyOk(message ...string) *Reply {
	if len(message) == 0 {
		message = []string{"Ok"}
	}
	return &Reply{250, message}
}

// ReplySenderOk creates a 250 Sender ok reply
func ReplySenderOk(sender string) *Reply { return &Reply{250, []string{"Sender " + sender + " ok"}} }

// ReplyRecipientOk creates a 250 Sender ok reply
func ReplyRecipientOk(recipient string) *Reply {
	return &Reply{250, []string{"Recipient " + recipient + " ok"}}
}

// ReplyAuthResponse creates a 334 authentication reply
func ReplyAuthResponse(response string) *Reply { return &Reply{334, []string{response}} }

// ReplyDataResponse creates a 354 data reply
func ReplyDataResponse() *Reply { return &Reply{354, []string{"End data with <CR><LF>.<CR><LF>"}} }

// ReplyStorageFailed creates a 452 error reply
func ReplyStorageFailed(reason string) *Reply { return &Reply{452, []string{reason}} }

// ReplyUnrecognisedCommand creates a 500 Unrecognised command reply
func ReplyUnrecognisedCommand() *Reply { return &Reply{500, []string{"Unrecognised command"}} }

// ReplyUnsupportedAuth creates a 504 unsupported authentication reply
func ReplyUnsupportedAuth() *Reply {
	return &Reply{504, []string{"Unsupported authentication mechanism"}}
}

// ReplyError creates a 500 error reply
func ReplyError(err error) *Reply { return &Reply{550, []string{err.Error()}} }

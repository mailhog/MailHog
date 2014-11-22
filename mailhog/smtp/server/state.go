package smtp

// http://www.rfc-editor.org/rfc/rfc5321.txt

// State represents the state of an SMTP conversation
type State int

// SMTP message conversation states
const (
	INVALID   = State(-1)
	ESTABLISH = State(iota)
	AUTH
	AUTHLOGIN
	MAIL
	RCPT
	DATA
	DONE
)

// StateMap provides string representations of SMTP conversation states
var StateMap = map[State]string{
	INVALID:   "INVALID",
	ESTABLISH: "ESTABLISH",
	AUTH:      "AUTH",
	AUTHLOGIN: "AUTHLOGIN",
	MAIL:      "MAIL",
	RCPT:      "RCPT",
	DATA:      "DATA",
	DONE:      "DONE",
}

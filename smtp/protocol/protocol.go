package protocol

// http://www.rfc-editor.org/rfc/rfc5321.txt

import (
	"errors"
	"log"
	"regexp"
	"strings"

	"github.com/ian-kent/Go-MailHog/data"
)

// Command is a struct representing an SMTP command (verb + arguments)
type Command struct {
	verb string
	args string
	orig string
}

// ParseCommand returns a Command from the line string
func ParseCommand(line string) *Command {
	words := strings.Split(line, " ")
	command := strings.ToUpper(words[0])
	args := strings.Join(words[1:len(words)], " ")

	return &Command{
		verb: command,
		args: args,
		orig: line,
	}
}

// Protocol is a state machine representing an SMTP session
type Protocol struct {
	state   State
	message *data.SMTPMessage

	lastCommand *Command

	Hostname string
	Ident    string

	// LogHandler is called for each log message. If nil, log messages will
	// be output using log.Printf instead.
	LogHandler func(message string, args ...interface{})
	// MessageReceivedHandler is called for each message accepted by the
	// SMTP protocol. It must return a MessageID or error. If nil, messages
	// will be rejected with an error.
	MessageReceivedHandler func(*data.Message) (string, error)
	// ValidateSenderHandler should return true if the sender is valid,
	// otherwise false. If nil, all senders will be accepted.
	ValidateSenderHandler func(from string) bool
	// ValidateRecipientHandler should return true if the recipient is valid,
	// otherwise false. If nil, all recipients will be accepted.
	ValidateRecipientHandler func(to string) bool
	// ValidateAuthenticationhandler should return true if the authentication
	// parameters are valid, otherwise false. If nil, all authentication
	// attempts will be accepted.
	ValidateAuthenticationHandler func(mechanism string, args ...string) (errorReply *Reply, ok bool)

	// RejectBrokenRCPTSyntax controls whether the protocol accepts technically
	// invalid syntax for the RCPT command. Set to true, the RCPT syntax requires
	// no space between `TO:` and the opening `<`
	RejectBrokenRCPTSyntax bool
	// RejectBrokenMAILSyntax controls whether the protocol accepts technically
	// invalid syntax for the MAIL command. Set to true, the MAIL syntax requires
	// no space between `FROM:` and the opening `<`
	RejectBrokenMAILSyntax bool
}

// NewProtocol returns a new SMTP state machine in INVALID state
// handler is called when a message is received and should return a message ID
func NewProtocol() *Protocol {
	return &Protocol{
		state:    INVALID,
		message:  &data.SMTPMessage{},
		Hostname: "mailhog.example",
		Ident:    "ESMTP Go-MailHog",
	}
}

func (proto *Protocol) logf(message string, args ...interface{}) {
	message = strings.Join([]string{"[PROTO: %s]", message}, " ")
	args = append([]interface{}{StateMap[proto.state]}, args...)

	if proto.LogHandler != nil {
		proto.LogHandler(message, args...)
	} else {
		log.Printf(message, args...)
	}
}

// Start begins an SMTP conversation with a 220 reply, placing the state
// machine in ESTABLISH state.
func (proto *Protocol) Start() *Reply {
	proto.logf("Started session, switching to ESTABLISH state")
	proto.state = ESTABLISH
	return ReplyIdent(proto.Hostname + " " + proto.Ident)
}

// Parse parses a line string and returns any remaining line string
// and a reply, if a command was found. Parse does nothing until a
// new line is found.
// - TODO decide whether to move this to a buffer inside Protocol
//   sort of like it this way, since it gives control back to the caller
func (proto *Protocol) Parse(line string) (string, *Reply) {
	var reply *Reply

	if !strings.Contains(line, "\n") {
		return line, reply
	}

	parts := strings.SplitN(line, "\n", 2)
	line = parts[1]

	// TODO collapse AUTH states into separate processing
	if proto.state == DATA {
		reply = proto.ProcessData(parts[0])
	} else {
		reply = proto.ProcessCommand(parts[0])
	}

	return line, reply
}

// ProcessData handles content received (with newlines stripped) while
// in the SMTP DATA state
func (proto *Protocol) ProcessData(line string) (reply *Reply) {
	proto.message.Data += line + "\n"

	if strings.HasSuffix(proto.message.Data, "\r\n.\r\n") {
		proto.message.Data = strings.Replace(proto.message.Data, "\r\n..", "\r\n.", -1)

		proto.logf("Got EOF, storing message and switching to MAIL state")
		proto.message.Data = strings.TrimSuffix(proto.message.Data, "\r\n.\r\n")
		proto.state = MAIL

		msg := proto.message.Parse(proto.Hostname)

		if proto.MessageReceivedHandler == nil {
			return ReplyStorageFailed("No storage backend")
		}

		id, err := proto.MessageReceivedHandler(msg)
		if err != nil {
			proto.logf("Error storing message: %s", err)
			return ReplyStorageFailed("Unable to store message")
		}
		return ReplyOk("Ok: queued as " + id)
	}

	return
}

// ProcessCommand processes a line of text as a command
// It expects the line string to be a properly formed SMTP verb and arguments
func (proto *Protocol) ProcessCommand(line string) (reply *Reply) {
	line = strings.Trim(line, "\r\n")
	proto.logf("Processing line: %s", line)

	words := strings.Split(line, " ")
	command := strings.ToUpper(words[0])
	args := strings.Join(words[1:len(words)], " ")
	proto.logf("In state %d, got command '%s', args '%s'", proto.state, command, args)

	cmd := ParseCommand(strings.TrimSuffix(line, "\r\n"))
	return proto.Command(cmd)
}

// Command applies an SMTP verb and arguments to the state machine
func (proto *Protocol) Command(command *Command) (reply *Reply) {
	defer func() {
		proto.lastCommand = command
	}()
	switch {
	case "RSET" == command.verb:
		proto.logf("Got RSET command, switching to MAIL state")
		proto.state = MAIL
		proto.message = &data.SMTPMessage{}
		return ReplyOk()
	case "NOOP" == command.verb:
		proto.logf("Got NOOP verb, staying in %s state", StateMap[proto.state])
		return ReplyOk()
	case "QUIT" == command.verb:
		proto.logf("Got QUIT verb, staying in %s state", StateMap[proto.state])
		proto.state = DONE
		return ReplyBye()
	case ESTABLISH == proto.state:
		switch command.verb {
		case "HELO":
			return proto.HELO(command.args)
		case "EHLO":
			return proto.EHLO(command.args)
		default:
			proto.logf("Got unknown command for ESTABLISH state: '%s'", command.verb)
			return ReplyUnrecognisedCommand()
		}
	case AUTHPLAIN == proto.state:
		proto.logf("Got PLAIN authentication response: '%s', switching to MAIL state", command.args)
		proto.state = MAIL
		if proto.ValidateAuthenticationHandler != nil {
			if reply, ok := proto.ValidateAuthenticationHandler("PLAIN", command.orig); !ok {
				return reply
			}
		}
		return ReplyAuthOk()
	case AUTHLOGIN == proto.state:
		proto.logf("Got LOGIN authentication response: '%s', switching to AUTHLOGIN2 state", command.args)
		proto.state = AUTHLOGIN2
		return ReplyAuthResponse("UGFzc3dvcmQ6")
	case AUTHLOGIN2 == proto.state:
		proto.logf("Got LOGIN authentication response: '%s', switching to MAIL state", command.args)
		proto.state = MAIL
		if proto.ValidateAuthenticationHandler != nil {
			if reply, ok := proto.ValidateAuthenticationHandler("LOGIN", proto.lastCommand.orig, command.orig); !ok {
				return reply
			}
		}
		return ReplyAuthOk()
	case AUTHCRAMMD5 == proto.state:
		proto.logf("Got CRAM-MD5 authentication response: '%s', switching to MAIL state", command.args)
		proto.state = MAIL
		if proto.ValidateAuthenticationHandler != nil {
			if reply, ok := proto.ValidateAuthenticationHandler("CRAM-MD5", command.orig); !ok {
				return reply
			}
		}
		return ReplyAuthOk()
	case MAIL == proto.state:
		switch command.verb {
		case "AUTH":
			proto.logf("Got AUTH command, staying in MAIL state")
			switch {
			case strings.HasPrefix(command.args, "PLAIN "):
				proto.logf("Got PLAIN authentication: %s", strings.TrimPrefix(command.args, "PLAIN "))
				if proto.ValidateAuthenticationHandler != nil {
					if reply, ok := proto.ValidateAuthenticationHandler("PLAIN", strings.TrimPrefix(command.args, "PLAIN ")); !ok {
						return reply
					}
				}
				return ReplyAuthOk()
			case "LOGIN" == command.args:
				proto.logf("Got LOGIN authentication, switching to AUTH state")
				proto.state = AUTHLOGIN
				return ReplyAuthResponse("VXNlcm5hbWU6")
			case "PLAIN" == command.args:
				proto.logf("Got PLAIN authentication (no args), switching to AUTH2 state")
				proto.state = AUTHPLAIN
				return ReplyAuthResponse("")
			case "CRAM-MD5" == command.args:
				proto.logf("Got CRAM-MD5 authentication, switching to AUTH state")
				proto.state = AUTHCRAMMD5
				return ReplyAuthResponse("PDQxOTI5NDIzNDEuMTI4Mjg0NzJAc291cmNlZm91ci5hbmRyZXcuY211LmVkdT4=")
			case strings.HasPrefix(command.args, "EXTERNAL "):
				proto.logf("Got EXTERNAL authentication: %s", strings.TrimPrefix(command.args, "EXTERNAL "))
				if proto.ValidateAuthenticationHandler != nil {
					if reply, ok := proto.ValidateAuthenticationHandler("EXTERNAL", strings.TrimPrefix(command.args, "EXTERNAL ")); !ok {
						return reply
					}
				}
				return ReplyAuthOk()
			default:
				return ReplyUnsupportedAuth()
			}
		case "MAIL":
			proto.logf("Got MAIL command, switching to RCPT state")
			from, err := proto.ParseMAIL(command.args)
			if err != nil {
				return ReplyError(err)
			}
			if proto.ValidateSenderHandler != nil {
				if !proto.ValidateSenderHandler(from) {
					// TODO correct sender error response
					return ReplyError(errors.New("Invalid sender " + from))
				}
			}
			proto.message.From = from
			proto.state = RCPT
			return ReplySenderOk(from)
		case "HELO":
			return proto.HELO(command.args)
		case "EHLO":
			return proto.EHLO(command.args)
		default:
			proto.logf("Got unknown command for MAIL state: '%s'", command)
			return ReplyUnrecognisedCommand()
		}
	case RCPT == proto.state:
		switch command.verb {
		case "RCPT":
			proto.logf("Got RCPT command")
			to, err := proto.ParseRCPT(command.args)
			if err != nil {
				return ReplyError(err)
			}
			if proto.ValidateRecipientHandler != nil {
				if !proto.ValidateRecipientHandler(to) {
					// TODO correct send error response
					return ReplyError(errors.New("Invalid recipient " + to))
				}
			}
			proto.message.To = append(proto.message.To, to)
			proto.state = RCPT
			return ReplyRecipientOk(to)
		case "HELO":
			return proto.HELO(command.args)
		case "EHLO":
			return proto.EHLO(command.args)
		case "DATA":
			proto.logf("Got DATA command, switching to DATA state")
			proto.state = DATA
			return ReplyDataResponse()
		default:
			proto.logf("Got unknown command for RCPT state: '%s'", command)
			return ReplyUnrecognisedCommand()
		}
	default:
		return ReplyUnrecognisedCommand()
	}
}

// HELO creates a reply to a HELO command
func (proto *Protocol) HELO(args string) (reply *Reply) {
	proto.logf("Got HELO command, switching to MAIL state")
	proto.state = MAIL
	proto.message.Helo = args
	return ReplyOk("Hello " + args)
}

// EHLO creates a reply to a EHLO command
func (proto *Protocol) EHLO(args string) (reply *Reply) {
	proto.logf("Got EHLO command, switching to MAIL state")
	proto.state = MAIL
	proto.message.Helo = args
	return ReplyOk("Hello "+args, "PIPELINING", "AUTH EXTERNAL CRAM-MD5 LOGIN PLAIN")
}

// ParseMAIL returns the forward-path from a MAIL command argument
func (proto *Protocol) ParseMAIL(mail string) (string, error) {
	var r *regexp.Regexp
	if proto.RejectBrokenMAILSyntax {
		r = regexp.MustCompile("(?i:From):<([^>]+)>")
	} else {
		r = regexp.MustCompile("(?i:From):\\s*<([^>]+)>")
	}
	match := r.FindStringSubmatch(mail)
	if len(match) != 2 {
		return "", errors.New("Invalid syntax in MAIL command")
	}
	return match[1], nil
}

// ParseRCPT returns the return-path from a RCPT command argument
func (proto *Protocol) ParseRCPT(rcpt string) (string, error) {
	var r *regexp.Regexp
	if proto.RejectBrokenRCPTSyntax {
		r = regexp.MustCompile("(?i:To):<([^>]+)>")
	} else {
		r = regexp.MustCompile("(?i:To):\\s*<([^>]+)>")
	}
	match := r.FindStringSubmatch(rcpt)
	if len(match) != 2 {
		return "", errors.New("Invalid syntax in RCPT command")
	}
	return match[1], nil
}

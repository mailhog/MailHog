package smtp

// http://www.rfc-editor.org/rfc/rfc5321.txt

import (
	"encoding/base64"
	"errors"
	"log"
	"regexp"
	"strings"

	"github.com/mailhog/data"
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
	lastCommand *Command

	TLSPending  bool
	TLSUpgraded bool

	State   State
	Message *data.SMTPMessage

	Hostname string
	Ident    string

	MaximumLineLength int
	MaximumRecipients int

	// LogHandler is called for each log message. If nil, log messages will
	// be output using log.Printf instead.
	LogHandler func(message string, args ...interface{})
	// MessageReceivedHandler is called for each message accepted by the
	// SMTP protocol. It must return a MessageID or error. If nil, messages
	// will be rejected with an error.
	MessageReceivedHandler func(*data.SMTPMessage) (string, error)
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
	// SMTPVerbFilter is called after each command is parsed, but before
	// any code is executed. This provides an opportunity to reject unwanted verbs,
	// e.g. to require AUTH before MAIL
	SMTPVerbFilter func(verb string, args ...string) (errorReply *Reply)
	// TLSHandler is called when a STARTTLS command is received.
	//
	// It should acknowledge the TLS request and set ok to true.
	// It should also return a callback which will be invoked after the reply is
	// sent. E.g., a TCP connection can only perform the upgrade after sending the reply
	//
	// Once the upgrade is complete, invoke the done function (e.g., from the returned callback)
	//
	// If TLS upgrade isn't possible, return an errorReply and set ok to false.
	TLSHandler func(done func(ok bool)) (errorReply *Reply, callback func(), ok bool)

	// GetAuthenticationMechanismsHandler should return an array of strings
	// listing accepted authentication mechanisms
	GetAuthenticationMechanismsHandler func() []string

	// RejectBrokenRCPTSyntax controls whether the protocol accepts technically
	// invalid syntax for the RCPT command. Set to true, the RCPT syntax requires
	// no space between `TO:` and the opening `<`
	RejectBrokenRCPTSyntax bool
	// RejectBrokenMAILSyntax controls whether the protocol accepts technically
	// invalid syntax for the MAIL command. Set to true, the MAIL syntax requires
	// no space between `FROM:` and the opening `<`
	RejectBrokenMAILSyntax bool
	// RequireTLS controls whether TLS is required for a connection before other
	// commands can be issued, applied at the protocol layer.
	RequireTLS bool
}

// NewProtocol returns a new SMTP state machine in INVALID state
// handler is called when a message is received and should return a message ID
func NewProtocol() *Protocol {
	p := &Protocol{
		Hostname:          "mailhog.example",
		Ident:             "ESMTP MailHog",
		State:             INVALID,
		MaximumLineLength: -1,
		MaximumRecipients: -1,
	}
	p.resetState()
	return p
}

func (proto *Protocol) resetState() {
	proto.Message = &data.SMTPMessage{}
}

func (proto *Protocol) logf(message string, args ...interface{}) {
	message = strings.Join([]string{"[PROTO: %s]", message}, " ")
	args = append([]interface{}{StateMap[proto.State]}, args...)

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
	proto.State = ESTABLISH
	return ReplyIdent(proto.Hostname + " " + proto.Ident)
}

// Parse parses a line string and returns any remaining line string
// and a reply, if a command was found. Parse does nothing until a
// new line is found.
// - TODO decide whether to move this to a buffer inside Protocol
//   sort of like it this way, since it gives control back to the caller
func (proto *Protocol) Parse(line string) (string, *Reply) {
	var reply *Reply

	if !strings.Contains(line, "\r\n") {
		return line, reply
	}

	parts := strings.SplitN(line, "\r\n", 2)
	line = parts[1]

	if proto.MaximumLineLength > -1 {
		if len(parts[0]) > proto.MaximumLineLength {
			return line, ReplyLineTooLong()
		}
	}

	// TODO collapse AUTH states into separate processing
	if proto.State == DATA {
		reply = proto.ProcessData(parts[0])
	} else {
		reply = proto.ProcessCommand(parts[0])
	}

	return line, reply
}

// ProcessData handles content received (with newlines stripped) while
// in the SMTP DATA state
func (proto *Protocol) ProcessData(line string) (reply *Reply) {
	proto.Message.Data += line + "\r\n"

	if strings.HasSuffix(proto.Message.Data, "\r\n.\r\n") {
		proto.Message.Data = strings.Replace(proto.Message.Data, "\r\n..", "\r\n.", -1)

		proto.logf("Got EOF, storing message and switching to MAIL state")
		proto.Message.Data = strings.TrimSuffix(proto.Message.Data, "\r\n.\r\n")
		proto.State = MAIL

		defer proto.resetState()

		if proto.MessageReceivedHandler == nil {
			return ReplyStorageFailed("No storage backend")
		}

		id, err := proto.MessageReceivedHandler(proto.Message)
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
	proto.logf("In state %d, got command '%s', args '%s'", proto.State, command, args)

	cmd := ParseCommand(strings.TrimSuffix(line, "\r\n"))
	return proto.Command(cmd)
}

// Command applies an SMTP verb and arguments to the state machine
func (proto *Protocol) Command(command *Command) (reply *Reply) {
	defer func() {
		proto.lastCommand = command
	}()
	if proto.SMTPVerbFilter != nil {
		proto.logf("sending to SMTP verb filter")
		r := proto.SMTPVerbFilter(command.verb)
		if r != nil {
			proto.logf("response returned by SMTP verb filter")
			return r
		}
	}
	switch {
	case proto.TLSPending && !proto.TLSUpgraded:
		proto.logf("Got command before TLS upgrade complete")
		// FIXME what to do?
		return ReplyBye()
	case "RSET" == command.verb:
		proto.logf("Got RSET command, switching to MAIL state")
		proto.State = MAIL
		proto.Message = &data.SMTPMessage{}
		return ReplyOk()
	case "NOOP" == command.verb:
		proto.logf("Got NOOP verb, staying in %s state", StateMap[proto.State])
		return ReplyOk()
	case "QUIT" == command.verb:
		proto.logf("Got QUIT verb, staying in %s state", StateMap[proto.State])
		proto.State = DONE
		return ReplyBye()
	case ESTABLISH == proto.State:
		proto.logf("In ESTABLISH state")
		switch command.verb {
		case "HELO":
			return proto.HELO(command.args)
		case "EHLO":
			return proto.EHLO(command.args)
		case "STARTTLS":
			return proto.STARTTLS(command.args)
		default:
			proto.logf("Got unknown command for ESTABLISH state: '%s'", command.verb)
			return ReplyUnrecognisedCommand()
		}
	case "STARTTLS" == command.verb:
		proto.logf("Got STARTTLS command outside ESTABLISH state")
		return proto.STARTTLS(command.args)
	case proto.RequireTLS && !proto.TLSUpgraded:
		proto.logf("RequireTLS set and not TLS not upgraded")
		return ReplyMustIssueSTARTTLSFirst()
	case AUTHPLAIN == proto.State:
		proto.logf("Got PLAIN authentication response: '%s', switching to MAIL state", command.args)
		proto.State = MAIL
		if proto.ValidateAuthenticationHandler != nil {
			// TODO error handling
			val, _ := base64.StdEncoding.DecodeString(command.orig)
			bits := strings.Split(string(val), string(rune(0)))

			if len(bits) < 3 {
				return ReplyError(errors.New("Badly formed parameter"))
			}

			user, pass := bits[1], bits[2]

			if reply, ok := proto.ValidateAuthenticationHandler("PLAIN", user, pass); !ok {
				return reply
			}
		}
		return ReplyAuthOk()
	case AUTHLOGIN == proto.State:
		proto.logf("Got LOGIN authentication response: '%s', switching to AUTHLOGIN2 state", command.args)
		proto.State = AUTHLOGIN2
		return ReplyAuthResponse("UGFzc3dvcmQ6")
	case AUTHLOGIN2 == proto.State:
		proto.logf("Got LOGIN authentication response: '%s', switching to MAIL state", command.args)
		proto.State = MAIL
		if proto.ValidateAuthenticationHandler != nil {
			if reply, ok := proto.ValidateAuthenticationHandler("LOGIN", proto.lastCommand.orig, command.orig); !ok {
				return reply
			}
		}
		return ReplyAuthOk()
	case AUTHCRAMMD5 == proto.State:
		proto.logf("Got CRAM-MD5 authentication response: '%s', switching to MAIL state", command.args)
		proto.State = MAIL
		if proto.ValidateAuthenticationHandler != nil {
			if reply, ok := proto.ValidateAuthenticationHandler("CRAM-MD5", command.orig); !ok {
				return reply
			}
		}
		return ReplyAuthOk()
	case MAIL == proto.State:
		proto.logf("In MAIL state")
		switch command.verb {
		case "AUTH":
			proto.logf("Got AUTH command, staying in MAIL state")
			switch {
			case strings.HasPrefix(command.args, "PLAIN "):
				proto.logf("Got PLAIN authentication: %s", strings.TrimPrefix(command.args, "PLAIN "))
				if proto.ValidateAuthenticationHandler != nil {
					val, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(command.args, "PLAIN "))
					bits := strings.Split(string(val), string(rune(0)))

					if len(bits) < 3 {
						return ReplyError(errors.New("Badly formed parameter"))
					}

					user, pass := bits[1], bits[2]

					if reply, ok := proto.ValidateAuthenticationHandler("PLAIN", user, pass); !ok {
						return reply
					}
				}
				return ReplyAuthOk()
			case "LOGIN" == command.args:
				proto.logf("Got LOGIN authentication, switching to AUTH state")
				proto.State = AUTHLOGIN
				return ReplyAuthResponse("VXNlcm5hbWU6")
			case "PLAIN" == command.args:
				proto.logf("Got PLAIN authentication (no args), switching to AUTH2 state")
				proto.State = AUTHPLAIN
				return ReplyAuthResponse("")
			case "CRAM-MD5" == command.args:
				proto.logf("Got CRAM-MD5 authentication, switching to AUTH state")
				proto.State = AUTHCRAMMD5
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
			proto.Message.From = from
			proto.State = RCPT
			return ReplySenderOk(from)
		case "HELO":
			return proto.HELO(command.args)
		case "EHLO":
			return proto.EHLO(command.args)
		default:
			proto.logf("Got unknown command for MAIL state: '%s'", command)
			return ReplyUnrecognisedCommand()
		}
	case RCPT == proto.State:
		proto.logf("In RCPT state")
		switch command.verb {
		case "RCPT":
			proto.logf("Got RCPT command")
			if proto.MaximumRecipients > -1 && len(proto.Message.To) >= proto.MaximumRecipients {
				return ReplyTooManyRecipients()
			}
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
			proto.Message.To = append(proto.Message.To, to)
			proto.State = RCPT
			return ReplyRecipientOk(to)
		case "HELO":
			return proto.HELO(command.args)
		case "EHLO":
			return proto.EHLO(command.args)
		case "DATA":
			proto.logf("Got DATA command, switching to DATA state")
			proto.State = DATA
			return ReplyDataResponse()
		default:
			proto.logf("Got unknown command for RCPT state: '%s'", command)
			return ReplyUnrecognisedCommand()
		}
	default:
		proto.logf("Command not recognised")
		return ReplyUnrecognisedCommand()
	}
}

// HELO creates a reply to a HELO command
func (proto *Protocol) HELO(args string) (reply *Reply) {
	proto.logf("Got HELO command, switching to MAIL state")
	proto.State = MAIL
	proto.Message.Helo = args
	return ReplyOk("Hello " + args)
}

// EHLO creates a reply to a EHLO command
func (proto *Protocol) EHLO(args string) (reply *Reply) {
	proto.logf("Got EHLO command, switching to MAIL state")
	proto.State = MAIL
	proto.Message.Helo = args
	replyArgs := []string{"Hello " + args, "PIPELINING"}

	if proto.TLSHandler != nil && !proto.TLSPending && !proto.TLSUpgraded {
		replyArgs = append(replyArgs, "STARTTLS")
	}

	if !proto.RequireTLS || proto.TLSUpgraded {
		if proto.GetAuthenticationMechanismsHandler != nil {
			mechanisms := proto.GetAuthenticationMechanismsHandler()
			if len(mechanisms) > 0 {
				replyArgs = append(replyArgs, "AUTH "+strings.Join(mechanisms, " "))
			}
		}
	}
	return ReplyOk(replyArgs...)
}

// STARTTLS creates a reply to a STARTTLS command
func (proto *Protocol) STARTTLS(args string) (reply *Reply) {
	if proto.TLSUpgraded {
		return ReplyUnrecognisedCommand()
	}

	if proto.TLSHandler == nil {
		proto.logf("tls handler not found")
		return ReplyUnrecognisedCommand()
	}

	if len(args) > 0 {
		return ReplySyntaxError("no parameters allowed")
	}

	r, callback, ok := proto.TLSHandler(func(ok bool) {
		proto.TLSUpgraded = ok
		proto.TLSPending = ok
		if ok {
			proto.resetState()
			proto.State = ESTABLISH
		}
	})
	if !ok {
		return r
	}

	proto.TLSPending = true
	return ReplyReadyToStartTLS(callback)
}

var parseMailBrokenRegexp = regexp.MustCompile("(?i:From):\\s*<([^>]+)>")
var parseMailRFCRegexp = regexp.MustCompile("(?i:From):<([^>]+)>")

// ParseMAIL returns the forward-path from a MAIL command argument
func (proto *Protocol) ParseMAIL(mail string) (string, error) {
	var match []string
	if proto.RejectBrokenMAILSyntax {
		match = parseMailRFCRegexp.FindStringSubmatch(mail)
	} else {
		match = parseMailBrokenRegexp.FindStringSubmatch(mail)
	}

	if len(match) != 2 {
		return "", errors.New("Invalid syntax in MAIL command")
	}
	return match[1], nil
}

var parseRcptBrokenRegexp = regexp.MustCompile("(?i:To):\\s*<([^>]+)>")
var parseRcptRFCRegexp = regexp.MustCompile("(?i:To):<([^>]+)>")

// ParseRCPT returns the return-path from a RCPT command argument
func (proto *Protocol) ParseRCPT(rcpt string) (string, error) {
	var match []string
	if proto.RejectBrokenRCPTSyntax {
		match = parseRcptRFCRegexp.FindStringSubmatch(rcpt)
	} else {
		match = parseRcptBrokenRegexp.FindStringSubmatch(rcpt)
	}
	if len(match) != 2 {
		return "", errors.New("Invalid syntax in RCPT command")
	}
	return match[1], nil
}

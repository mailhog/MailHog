MailHog SMTP Protocol [![GoDoc](https://godoc.org/github.com/mailhog/smtp?status.svg)](https://godoc.org/github.com/mailhog/smtp) [![Build Status](https://travis-ci.org/mailhog/smtp.svg?branch=master)](https://travis-ci.org/mailhog/smtp)
=========

`github.com/mailhog/smtp` implements an SMTP server state machine.

It attempts to encapsulate as much of the SMTP protocol (plus its extensions) as possible
without compromising configurability or requiring specific backend implementations.

  * ESMTP server implementing [RFC5321](http://tools.ietf.org/html/rfc5321)
  * Support for:
    * AUTH [RFC4954](http://tools.ietf.org/html/rfc4954)
    * PIPELINING [RFC2920](http://tools.ietf.org/html/rfc2920)
    * STARTTLS [RFC3207](http://tools.ietf.org/html/rfc3207)

```go
proto := NewProtocol()
reply := proto.Start()
reply = proto.ProcessCommand("EHLO localhost")
// ...
```

See [MailHog-Server](https://github.com/mailhog/MailHog-Server) and [MailHog-MTA](https://github.com/mailhog/MailHog-MTA) for example implementations.

### Commands and replies

Interaction with the state machine is via:
* the `Parse` function
* the `ProcessCommand` and `ProcessData` functions

You can mix the use of all three functions as necessary.

#### Parse

`Parse` should be used on a raw text stream. It looks for an end of line (`\r\n`), and if found, processes a single command. Any unprocessed data is returned.

If any unprocessed data is returned, `Parse` should be
called again to process then next command.

```go
text := "EHLO localhost\r\nMAIL FROM:<test>\r\nDATA\r\nTest\r\n.\r\n"

var reply *smtp.Reply
for {
  text, reply = proto.Parse(text)
  if len(text) == 0 {
    break
  }
}
```

#### ProcessCommand and ProcessData

`ProcessCommand` should be used for an already parsed command (i.e., a complete
SMTP "line" excluding the line ending).

`ProcessData` should be used if the protocol is in `DATA` state.

```go
reply = proto.ProcessCommand("EHLO localhost")
reply = proto.ProcessCommand("MAIL FROM:<test>")
reply = proto.ProcessCommand("DATA")
reply = proto.ProcessData("Test\r\n.\r\n")
```

### Hooks

The state machine provides hooks to manipulate its behaviour.

See [![GoDoc](https://godoc.org/github.com/mailhog/smtp?status.svg)](https://godoc.org/github.com/mailhog/smtp) for more information.

| Hook                               | Description
| ---------------------------------- | -----------
| LogHandler                         | Called for every log message
| MessageReceivedHandler             | Called for each message received
| ValidateSenderHandler              | Called after MAIL FROM
| ValidateRecipientHandler           | Called after RCPT TO
| ValidateAuthenticationHandler      | Called after AUTH
| SMTPVerbFilter                     | Called for every SMTP command processed
| TLSHandler                         | Callback mashup called after STARTTLS
| GetAuthenticationMechanismsHandler | Called for each EHLO command

### Behaviour flags

The state machine also exports variables to control its behaviour:

See [![GoDoc](https://godoc.org/github.com/mailhog/smtp?status.svg)](https://godoc.org/github.com/mailhog/smtp) for more information.

| Variable               | Description
| ---------------------- | -----------
| RejectBrokenRCPTSyntax | Reject non-conforming RCPT syntax
| RejectBrokenMAILSyntax | Reject non-conforming MAIL syntax
| RequireTLS             | Require STARTTLS before other commands
| MaximumRecipients      | Maximum recipients per message
| MaximumLineLength      | Maximum length of SMTP line

### Licence

Copyright ©‎ 2014-2015, Ian Kent (http://iankent.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.

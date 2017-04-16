package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"os/user"
)

import flag "github.com/spf13/pflag"

// Go runs the MailHog sendmail replacement.
func Go() {
	host, err := os.Hostname()
	if err != nil {
		host = "localhost"
	}

	username := "nobody"
	user, err := user.Current()
	if err == nil && user != nil && len(user.Username) > 0 {
		username = user.Username
	}

	fromAddr := username + "@" + host
	smtpAddr := "localhost:1025"
	var recip []string

	// defaults from envars if provided
	if len(os.Getenv("MH_SENDMAIL_SMTP_ADDR")) > 0 {
		smtpAddr = os.Getenv("MH_SENDMAIL_SMTP_ADDR")
	}
	if len(os.Getenv("MH_SENDMAIL_FROM")) > 0 {
		fromAddr = os.Getenv("MH_SENDMAIL_FROM")
	}

	var verbose bool

	// override defaults from cli flags
	flag.StringVar(&smtpAddr, "smtp-addr", smtpAddr, "SMTP server address")
	flag.StringVarP(&fromAddr, "from", "f", fromAddr, "SMTP sender")
	flag.BoolP("long-i", "i", true, "Ignored. This flag exists for sendmail compatibility.")
	flag.BoolP("long-t", "t", true, "Ignored. This flag exists for sendmail compatibility.")
	flag.BoolVarP(&verbose, "verbose", "v", false, "Verbose mode (sends debug output to stderr)")
	flag.Parse()

	// allow recipient to be passed as an argument
	recip = flag.Args()

	if verbose {
		fmt.Fprintln(os.Stderr, smtpAddr, fromAddr)
	}

	body, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading stdin")
		os.Exit(11)
	}

	msg, err := mail.ReadMessage(bytes.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error parsing message body")
		os.Exit(11)
	}

	if len(recip) == 0 {
		// We only need to parse the message to get a recipient if none where
		// provided on the command line.
		recip = append(recip, msg.Header.Get("To"))
	}

	err = smtp.SendMail(smtpAddr, nil, fromAddr, recip, body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error sending mail")
		log.Fatal(err)
	}

}

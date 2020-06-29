package main

import (
	"flag"
	"fmt"
	"os"

	"net/http"

	"github.com/gorilla/pat"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/doctolib/MailHog/generated/assets"
	"github.com/doctolib/MailHog/pkg/api"
	"github.com/doctolib/MailHog/pkg/config"
	"github.com/doctolib/MailHog/pkg/smtp"
	"github.com/doctolib/MailHog/pkg/web"
)

var conf *config.Config
var exitCh chan int
var version string

func configure() {
	config.RegisterFlags()
	flag.Parse()
	conf = config.Configure()
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-version" || os.Args[1] == "--version") {
		log.Infof("MailHog version: " + version)
		os.Exit(0)
	}
	if len(os.Args) > 1 && os.Args[1] == "bcrypt" {
		var pw string
		if len(os.Args) > 2 {
			pw = os.Args[2]
		} else {
			// TODO: read from stdin
			panic(fmt.Errorf("bcrypt command requires an argument"))
		}
		b, err := bcrypt.GenerateFromPassword([]byte(pw), 4)
		if err != nil {
			log.Fatalf("error bcrypting password: %s", err)
			os.Exit(1)
		}
		fmt.Println(string(b))
		os.Exit(0)
	}

	configure()

	if conf.AuthFile != "" {
		web.AuthFile(conf.AuthFile)
	}

	exitCh = make(chan int)
	cb := func(r http.Handler) {
		web.CreateWeb(conf, r.(*pat.Router), assets.Asset)
		api.CreateAPI(conf, r.(*pat.Router))
	}
	go web.Listen(conf.HTTPBindAddr, assets.Asset, exitCh, cb)
	go smtp.Listen(conf, exitCh)

	for range exitCh {
		log.Infof("Received exit signal")
		os.Exit(0)
	}
}

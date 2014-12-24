package main

import (
	"flag"
	"os"

	"github.com/ian-kent/go-log/log"
	gotcha "github.com/ian-kent/gotcha/app"
	"github.com/mailhog/MailHog-Server/api"
	"github.com/mailhog/MailHog-Server/config"
	"github.com/mailhog/MailHog-Server/smtp"
	"github.com/mailhog/MailHog-UI/assets"
	"github.com/mailhog/MailHog-UI/web"
	"github.com/mailhog/http"
)

var conf *config.Config
var exitCh chan int

func configure() {
	config.RegisterFlags()
	flag.Parse()
	conf = config.Configure()
}

func main() {
	configure()

	exitCh = make(chan int)
	cb := func(app *gotcha.App) {
		web.CreateWeb(conf, app)
		api.CreateAPIv1(conf, app)
	}
	go http.Listen(conf, assets.Asset, exitCh, cb)
	go smtp.Listen(conf, exitCh)

	for {
		select {
		case <-exitCh:
			log.Printf("Received exit signal")
			os.Exit(0)
		}
	}
}

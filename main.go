package main

import (
	"flag"
	"os"

	"github.com/ian-kent/go-log/log"
	gotcha "github.com/ian-kent/gotcha/app"
	"github.com/mailhog/MailHog-Server/api"
	cfgapi "github.com/mailhog/MailHog-Server/config"
	"github.com/mailhog/MailHog-Server/smtp"
	"github.com/mailhog/MailHog-UI/assets"
	cfgui "github.com/mailhog/MailHog-UI/config"
	"github.com/mailhog/MailHog-UI/web"
	"github.com/mailhog/http"
)

var apiconf *cfgapi.Config
var uiconf *cfgui.Config
var exitCh chan int

func configure() {
	cfgapi.RegisterFlags()
	cfgui.RegisterFlags()
	flag.Parse()
	apiconf = cfgapi.Configure()
	uiconf = cfgui.Configure()
}

func main() {
	configure()

	exitCh = make(chan int)
	cb := func(app *gotcha.App) {
		web.CreateWeb(uiconf, app)
		api.CreateAPIv1(apiconf, app)
	}
	go http.Listen(uiconf.HTTPBindAddr, assets.Asset, exitCh, cb)
	go smtp.Listen(apiconf, exitCh)

	for {
		select {
		case <-exitCh:
			log.Printf("Received exit signal")
			os.Exit(0)
		}
	}
}

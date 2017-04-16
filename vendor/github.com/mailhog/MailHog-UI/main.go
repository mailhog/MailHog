package main

import (
	"flag"
	"os"

	gohttp "net/http"

	"github.com/gorilla/pat"
	"github.com/ian-kent/go-log/log"
	"github.com/mailhog/MailHog-UI/assets"
	"github.com/mailhog/MailHog-UI/config"
	"github.com/mailhog/MailHog-UI/web"
	comcfg "github.com/mailhog/MailHog/config"
	"github.com/mailhog/http"
)

var conf *config.Config
var comconf *comcfg.Config
var exitCh chan int

func configure() {
	comcfg.RegisterFlags()
	config.RegisterFlags()
	flag.Parse()
	conf = config.Configure()
	comconf = comcfg.Configure()
	// FIXME hacky
	web.APIHost = conf.APIHost
}

func main() {
	configure()

	// FIXME need to make API URL configurable

	if comconf.AuthFile != "" {
		http.AuthFile(comconf.AuthFile)
	}

	exitCh = make(chan int)
	cb := func(r gohttp.Handler) {
		web.CreateWeb(conf, r.(*pat.Router), assets.Asset)
	}
	go http.Listen(conf.UIBindAddr, assets.Asset, exitCh, cb)

	for {
		select {
		case <-exitCh:
			log.Printf("Received exit signal")
			os.Exit(0)
		}
	}
}

package main

import (
	"flag"
	"fmt"
	"os"

	gohttp "net/http"

	"github.com/gorilla/pat"
	"github.com/ian-kent/go-log/log"
	"github.com/mailhog/MailHog-Server/api"
	cfgapi "github.com/mailhog/MailHog-Server/config"
	"github.com/mailhog/MailHog-Server/smtp"
	"github.com/mailhog/MailHog-UI/assets"
	cfgui "github.com/mailhog/MailHog-UI/config"
	"github.com/mailhog/MailHog-UI/web"
	cfgcom "github.com/mailhog/MailHog/config"
	"github.com/mailhog/http"
	"github.com/mailhog/mhsendmail/cmd"
	"golang.org/x/crypto/bcrypt"
)

var apiconf *cfgapi.Config
var uiconf *cfgui.Config
var comconf *cfgcom.Config
var exitCh chan int
var version string

func configure() {
	cfgcom.RegisterFlags()
	cfgapi.RegisterFlags()
	cfgui.RegisterFlags()
	flag.Parse()
	apiconf = cfgapi.Configure()
	uiconf = cfgui.Configure()
	comconf = cfgcom.Configure()

	apiconf.WebPath = comconf.WebPath
	uiconf.WebPath = comconf.WebPath
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-version" || os.Args[1] == "--version") {
		fmt.Println("MailHog version: " + version)
		os.Exit(0)
	}

	if len(os.Args) > 1 && os.Args[1] == "sendmail" {
		args := os.Args
		os.Args = []string{args[0]}
		if len(args) > 2 {
			os.Args = append(os.Args, args[2:]...)
		}
		cmd.Go()
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "bcrypt" {
		var pw string
		if len(os.Args) > 2 {
			pw = os.Args[2]
		} else {
			// TODO: read from stdin
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

	if comconf.AuthFile != "" {
		http.AuthFile(comconf.AuthFile)
	}

	exitCh = make(chan int)
	if uiconf.UIBindAddr == apiconf.APIBindAddr {
		cb := func(r gohttp.Handler) {
			web.CreateWeb(uiconf, r.(*pat.Router), assets.Asset)
			api.CreateAPI(apiconf, r.(*pat.Router))
		}
		go http.Listen(uiconf.UIBindAddr, assets.Asset, exitCh, cb)
	} else {
		cb1 := func(r gohttp.Handler) {
			api.CreateAPI(apiconf, r.(*pat.Router))
		}
		cb2 := func(r gohttp.Handler) {
			web.CreateWeb(uiconf, r.(*pat.Router), assets.Asset)
		}
		go http.Listen(apiconf.APIBindAddr, assets.Asset, exitCh, cb1)
		go http.Listen(uiconf.UIBindAddr, assets.Asset, exitCh, cb2)
	}
	go smtp.Listen(apiconf, exitCh)

	for {
		select {
		case <-exitCh:
			log.Printf("Received exit signal")
			os.Exit(0)
		}
	}
}

/*

Add some random content to the end of this file, hopefully tricking GitHub
into recognising this as a Go repo instead of Makefile.

A gopher, ASCII art style - borrowed from
https://gist.github.com/belbomemo/b5e7dad10fa567a5fe8a

          ,_---~~~~~----._
   _,,_,*^____      _____``*g*\"*,
  / __/ /'     ^.  /      \ ^@q   f
 [  @f | @))    |  | @))   l  0 _/
  \`/   \~____ / __ \_____/    \
   |           _l__l_           I
   }          [______]           I
   ]            | | |            |
   ]             ~ ~             |
   |                            |
    |                           |

*/

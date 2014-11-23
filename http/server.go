package http

import (
	"github.com/ian-kent/Go-MailHog/MailHog-Server/config"
	"github.com/ian-kent/go-log/log"
	gotcha "github.com/ian-kent/gotcha/app"
)

func Listen(cfg *config.Config, Asset func(string) ([]byte, error), exitCh chan int, registerCallback func(*gotcha.App)) {
	log.Info("[HTTP] Binding to address: %s", cfg.HTTPBindAddr)

	var app = gotcha.Create(Asset)
	app.Config.Listen = cfg.HTTPBindAddr

	registerCallback(app)

	app.Start()

	<-make(chan int)
	exitCh <- 1
}

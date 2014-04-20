package http

import (
	"regexp"
	"net/http"
	"github.com/ian-kent/MailHog/mailhog"
	"github.com/ian-kent/MailHog/mailhog/templates"
	"github.com/ian-kent/MailHog/mailhog/templates/images"
	"github.com/ian-kent/MailHog/mailhog/templates/js"
	"github.com/ian-kent/MailHog/mailhog/http/api"
	"github.com/ian-kent/MailHog/mailhog/http/handler"
)

var exitChannel chan int
var config *mailhog.Config

func web_exit(w http.ResponseWriter, r *http.Request, route *handler.Route) {
	web_headers(w)
	w.Write([]byte("Exiting MailHog!"))
	exitChannel <- 1
}

func web_index(w http.ResponseWriter, r *http.Request, route *handler.Route) {
	web_headers(w)
	w.Write([]byte(web_render(templates.Index())))
}

func web_jscontroller(w http.ResponseWriter, r *http.Request, route *handler.Route) {
	w.Header().Set("Content-Type", "text/javascript")
	w.Write([]byte(js.Controllers()))
}

func web_imgcontroller(w http.ResponseWriter, r *http.Request, route *handler.Route) {
	w.Header().Set("Content-Type", "image/png")
	w.Write(images.Hog())
}

func web_render(content string) string {
	return templates.Layout(content)
}

func web_headers(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
}

func Start(exitCh chan int, conf *mailhog.Config) {
	exitChannel = exitCh
	config = conf

	server := &http.Server{
		Addr: conf.HTTPBindAddr,
		Handler: &handler.RegexpHandler{},
	}

	server.Handler.(*handler.RegexpHandler).HandleFunc(regexp.MustCompile("^/exit/?$"), web_exit)
	server.Handler.(*handler.RegexpHandler).HandleFunc(regexp.MustCompile("^/js/controllers.js$"), web_jscontroller)
	server.Handler.(*handler.RegexpHandler).HandleFunc(regexp.MustCompile("^/images/hog.png$"), web_imgcontroller)
	server.Handler.(*handler.RegexpHandler).HandleFunc(regexp.MustCompile("^/$"), web_index)

	api.CreateAPIv1(exitCh, conf, server)

	server.ListenAndServe()
}

package http

import (
	"fmt"
	"net/http"
	"github.com/ian-kent/MailHog/mailhog"
	"github.com/ian-kent/MailHog/mailhog/templates"
	"github.com/ian-kent/MailHog/mailhog/templates/images"
	"github.com/ian-kent/MailHog/mailhog/templates/js"
)

var exitChannel chan int

func web_exit(w http.ResponseWriter, r *http.Request) {
	web_headers(w)
	fmt.Fprint(w, "Exiting MailHog!")
	exitChannel <- 1
}

func web_index(w http.ResponseWriter, r *http.Request) {
	web_headers(w)
	fmt.Fprint(w, web_render(templates.Index()))
}

func web_jscontroller(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")
	fmt.Fprint(w, js.Controllers())
}

func web_imgcontroller(w http.ResponseWriter, r *http.Request) {
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
	http.HandleFunc("/exit", web_exit)
	http.HandleFunc("/js/controllers.js", web_jscontroller)
	http.HandleFunc("/images/hog.png", web_imgcontroller)
	http.HandleFunc("/", web_index)
	http.ListenAndServe(conf.HTTPBindAddr, nil)
}

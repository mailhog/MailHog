package http

import (
	"net/http"
	"strings"
	"github.com/ian-kent/MailHog/mailhog/config"
	"github.com/ian-kent/MailHog/mailhog/http/api"
	"github.com/ian-kent/MailHog/mailhog/http/router"
)

var exitChannel chan int
var cfg *config.Config

// TODO clean this mess up

func web_exit(w http.ResponseWriter, r *http.Request, route *router.Route) {
	web_headers(w)
	w.Write([]byte("Exiting MailHog!"))
	exitChannel <- 1
}

func web_index(w http.ResponseWriter, r *http.Request, route *router.Route) {
	web_headers(w)
	data, _ := cfg.Assets("assets/templates/index.html")
	w.Write([]byte(web_render(string(data))))
}

func web_jscontroller(w http.ResponseWriter, r *http.Request, route *router.Route) {
	w.Header().Set("Content-Type", "text/javascript")
	data, _ := cfg.Assets("assets/js/controllers.js")
	w.Write(data)
}

func web_imgcontroller(w http.ResponseWriter, r *http.Request, route *router.Route) {
	w.Header().Set("Content-Type", "image/png")
	data, _ := cfg.Assets("assets/images/hog.png")
	w.Write(data)
}

func web_img_github(w http.ResponseWriter, r *http.Request, route *router.Route) {
	w.Header().Set("Content-Type", "image/png")
	data, _ := cfg.Assets("assets/images/github.png")
	w.Write(data)
}

func web_img_ajaxloader(w http.ResponseWriter, r *http.Request, route *router.Route) {
	w.Header().Set("Content-Type", "image/gif")
	data, _ := cfg.Assets("assets/images/ajax-loader.gif")
	w.Write(data)
}

func web_render(content string) string {
	data, _ := cfg.Assets("assets/templates/layout.html")
	layout := string(data)
	html := strings.Replace(layout, "<%= content %>", content, -1)
	// TODO clean this up
	html = strings.Replace(html, "<%= config[Hostname] %>", cfg.Hostname, -1)
	return html
}

func web_headers(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
}

func Start(exitCh chan int, conf *config.Config) {
	exitChannel = exitCh
	cfg = conf

	r := &router.Router{}
	server := &http.Server{
		Addr: conf.HTTPBindAddr,
		Handler: r,
	}

	r.Get("^/exit/?$", web_exit)
	r.Get("^/js/controllers.js$", web_jscontroller)
	r.Get("^/images/hog.png$", web_imgcontroller)
	r.Get("^/images/github.png$", web_img_github)
	r.Get("^/images/ajax-loader.gif$", web_img_ajaxloader)
	r.Get("^/$", web_index)

	api.CreateAPIv1(exitCh, conf, server)

	server.ListenAndServe()
}

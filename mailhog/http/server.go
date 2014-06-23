package http

import (
	"net/http"
	"strings"
	"log"
	"github.com/ian-kent/Go-MailHog/mailhog/config"
	"github.com/ian-kent/Go-MailHog/mailhog/http/api"
	"github.com/ian-kent/Go-MailHog/mailhog/http/router"
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

func web_static(w http.ResponseWriter, r *http.Request, route *router.Route) {
	match := route.Pattern.FindStringSubmatch(r.URL.Path)
	file := match[1]
	log.Printf("[HTTP] GET %s\n", file)	

	if strings.HasSuffix(file, ".gif") {
		w.Header().Set("Content-Type", "image/gif")
	} else if strings.HasSuffix(file, ".png") {
		w.Header().Set("Content-Type", "image/png")
	} else if strings.HasSuffix(file, ".js") {
		w.Header().Set("Content-Type", "text/javascript")
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}

	data, err := cfg.Assets("assets" + file)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	
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
	r.Get("^(/js/controllers.js)$", web_static)
	r.Get("^(/images/hog.png)$", web_static)
	r.Get("^(/images/github.png)$", web_static)
	r.Get("^(/images/ajax-loader.gif)$", web_static)
	r.Get("^/$", web_index)

	api.CreateAPIv1(exitCh, conf, server)

	server.ListenAndServe()
}

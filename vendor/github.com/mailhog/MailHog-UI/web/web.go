package web

import (
	"bytes"
	"html/template"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/pat"
	"github.com/mailhog/MailHog-UI/config"
)

var APIHost string
var WebPath string

type Web struct {
	config *config.Config
	asset  func(string) ([]byte, error)
}

func CreateWeb(cfg *config.Config, r http.Handler, asset func(string) ([]byte, error)) *Web {
	web := &Web{
		config: cfg,
		asset:  asset,
	}

	pat := r.(*pat.Router)

	WebPath = cfg.WebPath

	log.Printf("Serving under http://%s%s/", cfg.UIBindAddr, WebPath)

	pat.Path(WebPath + "/images/{file:.*}").Methods("GET").HandlerFunc(web.Static("assets/images/{{file}}"))
	pat.Path(WebPath + "/css/{file:.*}").Methods("GET").HandlerFunc(web.Static("assets/css/{{file}}"))
	pat.Path(WebPath + "/js/{file:.*}").Methods("GET").HandlerFunc(web.Static("assets/js/{{file}}"))
	pat.Path(WebPath + "/fonts/{file:.*}").Methods("GET").HandlerFunc(web.Static("assets/fonts/{{file}}"))
	pat.Path(WebPath + "/healthz").Methods("GET").HandlerFunc(web.HealthCheck())
	pat.StrictSlash(true).Path(WebPath + "/").Methods("GET").HandlerFunc(web.Index())

	return web
}

func (web Web) HealthCheck() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Println("[UI] GET /healthz")

		w.WriteHeader(200)
		w.Write([]byte("UP"))
	}
}

func (web Web) Static(pattern string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		fp := strings.TrimSuffix(pattern, "{{file}}") + req.URL.Query().Get(":file")
		if b, err := web.asset(fp); err == nil {
			ext := filepath.Ext(fp)

			w.Header().Set("Content-Type", mime.TypeByExtension(ext))
			w.WriteHeader(200)
			w.Write(b)
			return
		}
		log.Printf("[UI] File not found: %s", fp)
		w.WriteHeader(404)
	}
}

func (web Web) Index() func(http.ResponseWriter, *http.Request) {
	tmpl := template.New("index.html")
	tmpl.Delims("[:", ":]")

	asset, err := web.asset("assets/templates/index.html")
	if err != nil {
		log.Fatalf("[UI] Error loading index.html: %s", err)
	}

	tmpl, err = tmpl.Parse(string(asset))
	if err != nil {
		log.Fatalf("[UI] Error parsing index.html: %s", err)
	}

	layout := template.New("layout.html")
	layout.Delims("[:", ":]")

	asset, err = web.asset("assets/templates/layout.html")
	if err != nil {
		log.Fatalf("[UI] Error loading layout.html: %s", err)
	}

	layout, err = layout.Parse(string(asset))
	if err != nil {
		log.Fatalf("[UI] Error parsing layout.html: %s", err)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		data := map[string]interface{}{
			"config":  web.config,
			"Page":    "Browse",
			"APIHost": APIHost,
		}

		b := new(bytes.Buffer)
		err := tmpl.Execute(b, data)

		if err != nil {
			log.Printf("[UI] Error executing template: %s", err)
			w.WriteHeader(500)
			return
		}

		data["Content"] = template.HTML(b.String())

		b = new(bytes.Buffer)
		err = layout.Execute(b, data)

		if err != nil {
			log.Printf("[UI] Error executing template: %s", err)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		w.Write(b.Bytes())
	}
}

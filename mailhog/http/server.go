package http

import (
	"encoding/json"
	"net/http"
	"regexp"
	"github.com/ian-kent/MailHog/mailhog"
	"github.com/ian-kent/MailHog/mailhog/templates"
	"github.com/ian-kent/MailHog/mailhog/storage"
	"github.com/ian-kent/MailHog/mailhog/templates/images"
	"github.com/ian-kent/MailHog/mailhog/templates/js"
)

var exitChannel chan int
var config *mailhog.Config

func web_exit(w http.ResponseWriter, r *http.Request) {
	web_headers(w)
	w.Write([]byte("Exiting MailHog!"))
	exitChannel <- 1
}

func web_index(w http.ResponseWriter, r *http.Request) {
	web_headers(w)
	w.Write([]byte(web_render(templates.Index())))
}

func web_jscontroller(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")
	w.Write([]byte(js.Controllers()))
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

func api_messages(w http.ResponseWriter, r *http.Request) {
	re, _ := regexp.Compile("/api/v1/messages/([0-9a-f]+)/delete")
	match := re.FindStringSubmatch(r.URL.Path)
	if len(match) > 0 {
		api_delete_one(w, r, match[1])
		return
	}

	// TODO start, limit
	messages, _ := storage.List(config, 0, 1000)
	bytes, _ := json.Marshal(messages)
	w.Header().Set("Content-Type", "text/json")
	w.Write(bytes)
}

func api_delete_all(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/json")
	storage.DeleteAll(config)
}

func api_delete_one(w http.ResponseWriter, r *http.Request, id string) {
	w.Header().Set("Content-Type", "text/json")
	storage.DeleteOne(config, id)
}

func Start(exitCh chan int, conf *mailhog.Config) {
	exitChannel = exitCh
	config = conf

	http.HandleFunc("/exit", web_exit)
	http.HandleFunc("/js/controllers.js", web_jscontroller)
	http.HandleFunc("/images/hog.png", web_imgcontroller)
	http.HandleFunc("/", web_index)
	http.HandleFunc("/api/v1/messages/", api_messages)
	http.HandleFunc("/api/v1/messages/delete", api_delete_all)
	http.ListenAndServe(conf.HTTPBindAddr, nil)
}

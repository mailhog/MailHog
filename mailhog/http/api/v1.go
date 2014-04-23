package api

import (
	"log"
	"encoding/json"
	"net/http"
	"regexp"
	"github.com/ian-kent/MailHog/mailhog/config"
	"github.com/ian-kent/MailHog/mailhog/storage"
	"github.com/ian-kent/MailHog/mailhog/http/handler"
)

type APIv1 struct {
	config *config.Config
	exitChannel chan int
	server *http.Server
}

func CreateAPIv1(exitCh chan int, conf *config.Config, server *http.Server) *APIv1 {
	log.Println("Creating API v1")
	apiv1 := &APIv1{
		config: conf,
		exitChannel: exitCh,
		server: server,
	}

	server.Handler.(*handler.RegexpHandler).HandleFunc(regexp.MustCompile("^/api/v1/messages/?$"), apiv1.messages)
	server.Handler.(*handler.RegexpHandler).HandleFunc(regexp.MustCompile("^/api/v1/messages/delete/?$"), apiv1.delete_all)
	server.Handler.(*handler.RegexpHandler).HandleFunc(regexp.MustCompile("^/api/v1/messages/([0-9a-f]+)/?$"), apiv1.message)
	server.Handler.(*handler.RegexpHandler).HandleFunc(regexp.MustCompile("^/api/v1/messages/([0-9a-f]+)/delete/?$"), apiv1.delete_one)

	return apiv1
}

func (apiv1 *APIv1) messages(w http.ResponseWriter, r *http.Request, route *handler.Route) {
	log.Println("[APIv1] GET /api/v1/messages")

	// TODO start, limit
	switch apiv1.config.Storage.(type) {
		case *storage.MongoDB:
			messages, _ := apiv1.config.Storage.(*storage.MongoDB).List(0, 1000)
			bytes, _ := json.Marshal(messages)
			w.Header().Set("Content-Type", "text/json")
			w.Write(bytes)
		case *storage.Memory:
			messages, _ := apiv1.config.Storage.(*storage.Memory).List(0, 1000)
			bytes, _ := json.Marshal(messages)
			w.Header().Set("Content-Type", "text/json")
			w.Write(bytes)
		default:
			w.Header().Set("Content-Type", "text/json")
			w.Write([]byte("[]"))
	}
}

func (apiv1 *APIv1) message(w http.ResponseWriter, r *http.Request, route *handler.Route) {
	match := route.Pattern.FindStringSubmatch(r.URL.Path)
	id := match[1]
	log.Printf("[APIv1] GET /api/v1/messages/%s\n", id)

	switch apiv1.config.Storage.(type) {
		case *storage.MongoDB:
			message, _ := apiv1.config.Storage.(*storage.MongoDB).Load(id)
			bytes, _ := json.Marshal(message)
			w.Header().Set("Content-Type", "text/json")
			w.Write(bytes)
		case *storage.Memory:
			message, _ := apiv1.config.Storage.(*storage.Memory).Load(id)
			bytes, _ := json.Marshal(message)
			w.Header().Set("Content-Type", "text/json")
			w.Write(bytes)
		default:
			w.Header().Set("Content-Type", "text/json")
			w.Write([]byte("[]"))
	}
}

func (apiv1 *APIv1) delete_all(w http.ResponseWriter, r *http.Request, route *handler.Route) {
	log.Println("[APIv1] POST /api/v1/messages/delete")

	w.Header().Set("Content-Type", "text/json")
	switch apiv1.config.Storage.(type) {
		case *storage.MongoDB:
			apiv1.config.Storage.(*storage.MongoDB).DeleteAll()
		case *storage.Memory:
			apiv1.config.Storage.(*storage.Memory).DeleteAll()
	}
}

func (apiv1 *APIv1) delete_one(w http.ResponseWriter, r *http.Request, route *handler.Route) {
	match := route.Pattern.FindStringSubmatch(r.URL.Path)
	id := match[1]
	log.Printf("[APIv1] POST /api/v1/messages/%s/delete\n", id)

	w.Header().Set("Content-Type", "text/json")
	switch apiv1.config.Storage.(type) {
		case *storage.MongoDB:
			apiv1.config.Storage.(*storage.MongoDB).DeleteOne(id)
		case *storage.Memory:
			apiv1.config.Storage.(*storage.Memory).DeleteOne(id)
	}
}

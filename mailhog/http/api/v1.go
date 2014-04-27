package api

import (
	"log"
	"encoding/json"
	"net/http"
	"net/smtp"
	"strconv"
	"github.com/ian-kent/MailHog/mailhog/data"
	"github.com/ian-kent/MailHog/mailhog/config"
	"github.com/ian-kent/MailHog/mailhog/storage"
	"github.com/ian-kent/MailHog/mailhog/http/router"
)

type APIv1 struct {
	config *config.Config
	exitChannel chan int
	server *http.Server
}

type ReleaseConfig struct {
	Email string
	Host string
	Port string
}

func CreateAPIv1(exitCh chan int, conf *config.Config, server *http.Server) *APIv1 {
	log.Println("Creating API v1")
	apiv1 := &APIv1{
		config: conf,
		exitChannel: exitCh,
		server: server,
	}

	r := server.Handler.(*router.Router)

	r.Get("^/api/v1/messages/?$", apiv1.messages)
	r.Delete("^/api/v1/messages/?$", apiv1.delete_all)
	r.Get("^/api/v1/messages/([0-9a-f]+)/?$", apiv1.message)
	r.Delete("^/api/v1/messages/([0-9a-f]+)/?$", apiv1.delete_one)
	r.Get("^/api/v1/messages/([0-9a-f]+)/download/?$", apiv1.download)
	r.Get("^/api/v1/messages/([0-9a-f]+)/mime/part/(\\d+)/download/?$", apiv1.download_part)
	r.Post("^/api/v1/messages/([0-9a-f]+)/release/?$", apiv1.release_one)

	return apiv1
}

func (apiv1 *APIv1) messages(w http.ResponseWriter, r *http.Request, route *router.Route) {
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
			w.WriteHeader(500)
	}
}

func (apiv1 *APIv1) message(w http.ResponseWriter, r *http.Request, route *router.Route) {

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
			w.WriteHeader(500)
	}
}

func (apiv1 *APIv1) download(w http.ResponseWriter, r *http.Request, route *router.Route) {
	match := route.Pattern.FindStringSubmatch(r.URL.Path)
	id := match[1]
	log.Printf("[APIv1] GET /api/v1/messages/%s/download\n", id)

	w.Header().Set("Content-Type", "message/rfc822")
	w.Header().Set("Content-Disposition", "attachment; filename=\"" + id + ".eml\"")

	switch apiv1.config.Storage.(type) {
		case *storage.MongoDB:
			message, _ := apiv1.config.Storage.(*storage.MongoDB).Load(id)
			for h, l := range message.Content.Headers {
				for _, v := range l {
					w.Write([]byte(h + ": " + v + "\r\n"))
				}
			}
			w.Write([]byte("\r\n" + message.Content.Body))
		case *storage.Memory:
			message, _ := apiv1.config.Storage.(*storage.Memory).Load(id)
			for h, l := range message.Content.Headers {
				for _, v := range l {
					w.Write([]byte(h + ": " + v + "\r\n"))
				}
			}
			w.Write([]byte("\r\n" + message.Content.Body))
		default:
			w.WriteHeader(500)
	}
}

func (apiv1 *APIv1) download_part(w http.ResponseWriter, r *http.Request, route *router.Route) {
	match := route.Pattern.FindStringSubmatch(r.URL.Path)
	id := match[1]
	part, _ := strconv.Atoi(match[2])
	log.Printf("[APIv1] GET /api/v1/messages/%s/mime/part/%d/download\n", id, part)

	// TODO extension from content-type?

	w.Header().Set("Content-Disposition", "attachment; filename=\"" + id + "-part-" + match[2] + "\"")

	switch apiv1.config.Storage.(type) {
		case *storage.MongoDB:
			message, _ := apiv1.config.Storage.(*storage.MongoDB).Load(id)
			for h, l := range message.MIME.Parts[part].Headers {
				for _, v := range l {
					w.Header().Set(h, v)
				}
			}
			w.Write([]byte("\r\n" + message.MIME.Parts[part].Body))
		case *storage.Memory:
			message, _ := apiv1.config.Storage.(*storage.Memory).Load(id)
			for h, l := range message.MIME.Parts[part].Headers {
				for _, v := range l {
					w.Header().Set(h, v)
				}
			}
			w.Write([]byte("\r\n" + message.MIME.Parts[part].Body))
		default:
			w.WriteHeader(500)
	}
}

func (apiv1 *APIv1) delete_all(w http.ResponseWriter, r *http.Request, route *router.Route) {
	log.Println("[APIv1] POST /api/v1/messages")

	w.Header().Set("Content-Type", "text/json")
	switch apiv1.config.Storage.(type) {
		case *storage.MongoDB:
			apiv1.config.Storage.(*storage.MongoDB).DeleteAll()
		case *storage.Memory:
			apiv1.config.Storage.(*storage.Memory).DeleteAll()
		default:
			w.WriteHeader(500)
	}
}

func (apiv1 *APIv1) release_one(w http.ResponseWriter, r *http.Request, route *router.Route) {
	match := route.Pattern.FindStringSubmatch(r.URL.Path)
	id := match[1]
	log.Printf("[APIv1] POST /api/v1/messages/%s/release\n", id)

	w.Header().Set("Content-Type", "text/json")
	var msg = &data.Message{}
	switch apiv1.config.Storage.(type) {
		case *storage.MongoDB:
			msg, _ = apiv1.config.Storage.(*storage.MongoDB).Load(id)
		case *storage.Memory:
			msg, _ = apiv1.config.Storage.(*storage.Memory).Load(id)
		default:
			w.WriteHeader(500)
	}

	decoder := json.NewDecoder(r.Body)
	var cfg ReleaseConfig
	err := decoder.Decode(&cfg)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		w.WriteHeader(500)
		w.Write([]byte("Error decoding request body"))
		return
	}

	log.Printf("Releasing to %s (via %s:%s)", cfg.Email, cfg.Host, cfg.Port)
	log.Printf("Got message: %s", msg.Id)

	bytes := make([]byte, 0)
	for h, l := range msg.Content.Headers {
		for _, v := range l {
			bytes = append(bytes, []byte(h + ": " + v + "\r\n")...)
		}
	}
	bytes = append(bytes, []byte("\r\n" + msg.Content.Body)...)

	err = smtp.SendMail(cfg.Host + ":" + cfg.Port, nil, "nobody@" + apiv1.config.Hostname, []string{cfg.Email}, bytes)
	if err != nil {
		log.Printf("Failed to release message: %s", err)
		w.WriteHeader(500)
		return
	}
	log.Printf("Message released successfully")
}

func (apiv1 *APIv1) delete_one(w http.ResponseWriter, r *http.Request, route *router.Route) {
	match := route.Pattern.FindStringSubmatch(r.URL.Path)
	id := match[1]
	log.Printf("[APIv1] POST /api/v1/messages/%s/delete\n", id)

	w.Header().Set("Content-Type", "text/json")
	switch apiv1.config.Storage.(type) {
		case *storage.MongoDB:
			apiv1.config.Storage.(*storage.MongoDB).DeleteOne(id)
		case *storage.Memory:
			apiv1.config.Storage.(*storage.Memory).DeleteOne(id)
		default:
			w.WriteHeader(500)
	}
}

package api

import (
	"encoding/json"
	"net/smtp"
	"strconv"

	"github.com/ian-kent/Go-MailHog/config"
	"github.com/ian-kent/Go-MailHog/data"
	"github.com/ian-kent/Go-MailHog/storage"
	"github.com/ian-kent/go-log/log"
	gotcha "github.com/ian-kent/gotcha/app"
	"github.com/ian-kent/gotcha/http"
)

type APIv1 struct {
	config         *config.Config
	eventlisteners []*EventListener
	app            *gotcha.App
}

type EventListener struct {
	session *http.Session
	ch      chan []byte
}

type ReleaseConfig struct {
	Email string
	Host  string
	Port  string
}

func CreateAPIv1(conf *config.Config, app *gotcha.App) *APIv1 {
	log.Println("Creating API v1")
	apiv1 := &APIv1{
		config:         conf,
		eventlisteners: make([]*EventListener, 0),
		app:            app,
	}

	r := app.Router

	r.Get("/api/v1/messages/?", apiv1.messages)
	r.Delete("/api/v1/messages/?", apiv1.delete_all)
	r.Get("/api/v1/messages/(?P<id>[0-9a-f]+)/?", apiv1.message)
	r.Delete("/api/v1/messages/(?P<id>[0-9a-f]+)/?", apiv1.delete_one)
	r.Get("/api/v1/messages/(?P<id>[0-9a-f]+)/download/?", apiv1.download)
	r.Get("/api/v1/messages/(?P<id>[0-9a-f]+)/mime/part/(\\d+)/download/?", apiv1.download_part)
	r.Post("/api/v1/messages/(?P<id>[0-9a-f]+)/release/?", apiv1.release_one)
	r.Get("/api/v1/events/?", apiv1.eventstream)

	go func() {
		for {
			select {
			case msg := <-apiv1.config.MessageChan:
				log.Println("Got message in APIv1 event stream")
				bytes, _ := json.MarshalIndent(msg, "", "  ")
				json := string(bytes)
				log.Printf("Sending content: %s\n", json)
				apiv1.broadcast(json)
			}
		}
	}()

	return apiv1
}

func (apiv1 *APIv1) broadcast(json string) {
	log.Println("[APIv1] BROADCAST /api/v1/events")
	b := []byte(json)
	for _, l := range apiv1.eventlisteners {
		log.Printf("Sending to connection: %s\n", l.session.Request.RemoteAddr)
		l.ch <- b
	}
}

func (apiv1 *APIv1) eventstream(session *http.Session) {
	log.Println("[APIv1] GET /api/v1/events")

	apiv1.eventlisteners = append(apiv1.eventlisteners, &EventListener{
		session,
		session.Response.EventStream(),
	})
}

func (apiv1 *APIv1) messages(session *http.Session) {
	log.Println("[APIv1] GET /api/v1/messages")

	// TODO start, limit
	switch apiv1.config.Storage.(type) {
	case *storage.MongoDB:
		messages, _ := apiv1.config.Storage.(*storage.MongoDB).List(0, 1000)
		bytes, _ := json.Marshal(messages)
		session.Response.Headers.Add("Content-Type", "text/json")
		session.Response.Write(bytes)
	case *storage.InMemory:
		messages, _ := apiv1.config.Storage.(*storage.InMemory).List(0, 1000)
		bytes, _ := json.Marshal(messages)
		session.Response.Headers.Add("Content-Type", "text/json")
		session.Response.Write(bytes)
	default:
		session.Response.Status = 500
	}
}

func (apiv1 *APIv1) message(session *http.Session) {
	id := session.Stash["id"].(string)
	log.Printf("[APIv1] GET /api/v1/messages/%s\n", id)

	switch apiv1.config.Storage.(type) {
	case *storage.MongoDB:
		message, _ := apiv1.config.Storage.(*storage.MongoDB).Load(id)
		bytes, _ := json.Marshal(message)
		session.Response.Headers.Add("Content-Type", "text/json")
		session.Response.Write(bytes)
	case *storage.InMemory:
		message, _ := apiv1.config.Storage.(*storage.InMemory).Load(id)
		bytes, _ := json.Marshal(message)
		session.Response.Headers.Add("Content-Type", "text/json")
		session.Response.Write(bytes)
	default:
		session.Response.Status = 500
	}
}

func (apiv1 *APIv1) download(session *http.Session) {
	id := session.Stash["id"].(string)
	log.Printf("[APIv1] GET /api/v1/messages/%s\n", id)

	session.Response.Headers.Add("Content-Type", "message/rfc822")
	session.Response.Headers.Add("Content-Disposition", "attachment; filename=\""+id+".eml\"")

	switch apiv1.config.Storage.(type) {
	case *storage.MongoDB:
		message, _ := apiv1.config.Storage.(*storage.MongoDB).Load(id)
		for h, l := range message.Content.Headers {
			for _, v := range l {
				session.Response.Write([]byte(h + ": " + v + "\r\n"))
			}
		}
		session.Response.Write([]byte("\r\n" + message.Content.Body))
	case *storage.InMemory:
		message, _ := apiv1.config.Storage.(*storage.InMemory).Load(id)
		for h, l := range message.Content.Headers {
			for _, v := range l {
				session.Response.Write([]byte(h + ": " + v + "\r\n"))
			}
		}
		session.Response.Write([]byte("\r\n" + message.Content.Body))
	default:
		session.Response.Status = 500
	}
}

func (apiv1 *APIv1) download_part(session *http.Session) {
	id := session.Stash["id"].(string)
	part, _ := strconv.Atoi(session.Stash["part"].(string))
	log.Printf("[APIv1] GET /api/v1/messages/%s/mime/part/%d/download\n", id, part)

	// TODO extension from content-type?

	session.Response.Headers.Add("Content-Disposition", "attachment; filename=\""+id+"-part-"+strconv.Itoa(part)+"\"")

	switch apiv1.config.Storage.(type) {
	case *storage.MongoDB:
		message, _ := apiv1.config.Storage.(*storage.MongoDB).Load(id)
		for h, l := range message.MIME.Parts[part].Headers {
			for _, v := range l {
				session.Response.Headers.Add(h, v)
			}
		}
		session.Response.Write([]byte("\r\n" + message.MIME.Parts[part].Body))
	case *storage.InMemory:
		message, _ := apiv1.config.Storage.(*storage.InMemory).Load(id)
		for h, l := range message.MIME.Parts[part].Headers {
			for _, v := range l {
				session.Response.Headers.Add(h, v)
			}
		}
		session.Response.Write([]byte("\r\n" + message.MIME.Parts[part].Body))
	default:
		session.Response.Status = 500
	}
}

func (apiv1 *APIv1) delete_all(session *http.Session) {
	log.Println("[APIv1] POST /api/v1/messages")

	session.Response.Headers.Add("Content-Type", "text/json")
	switch apiv1.config.Storage.(type) {
	case *storage.MongoDB:
		apiv1.config.Storage.(*storage.MongoDB).DeleteAll()
	case *storage.InMemory:
		apiv1.config.Storage.(*storage.InMemory).DeleteAll()
	default:
		session.Response.Status = 500
		return
	}
}

func (apiv1 *APIv1) release_one(session *http.Session) {
	id := session.Stash["id"].(string)
	log.Printf("[APIv1] POST /api/v1/messages/%s/release\n", id)

	session.Response.Headers.Add("Content-Type", "text/json")
	var msg = &data.Message{}
	switch apiv1.config.Storage.(type) {
	case *storage.MongoDB:
		msg, _ = apiv1.config.Storage.(*storage.MongoDB).Load(id)
	case *storage.InMemory:
		msg, _ = apiv1.config.Storage.(*storage.InMemory).Load(id)
	default:
		session.Response.Status = 500
		return
	}

	decoder := json.NewDecoder(session.Request.Body())
	var cfg ReleaseConfig
	err := decoder.Decode(&cfg)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		session.Response.Status = 500
		session.Response.Write([]byte("Error decoding request body"))
		return
	}

	log.Printf("Releasing to %s (via %s:%s)", cfg.Email, cfg.Host, cfg.Port)
	log.Printf("Got message: %s", msg.ID)

	bytes := make([]byte, 0)
	for h, l := range msg.Content.Headers {
		for _, v := range l {
			bytes = append(bytes, []byte(h+": "+v+"\r\n")...)
		}
	}
	bytes = append(bytes, []byte("\r\n"+msg.Content.Body)...)

	err = smtp.SendMail(cfg.Host+":"+cfg.Port, nil, "nobody@"+apiv1.config.Hostname, []string{cfg.Email}, bytes)
	if err != nil {
		log.Printf("Failed to release message: %s", err)
		session.Response.Status = 500
		return
	}
	log.Printf("Message released successfully")
}

func (apiv1 *APIv1) delete_one(session *http.Session) {
	id := session.Stash["id"].(string)
	log.Printf("[APIv1] POST /api/v1/messages/%s/delete\n", id)

	session.Response.Headers.Add("Content-Type", "text/json")
	switch apiv1.config.Storage.(type) {
	case *storage.MongoDB:
		apiv1.config.Storage.(*storage.MongoDB).DeleteOne(id)
	case *storage.InMemory:
		apiv1.config.Storage.(*storage.InMemory).DeleteOne(id)
	default:
		session.Response.Status = 500
	}
}

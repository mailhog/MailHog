package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/pat"
	"github.com/ian-kent/go-log/log"
	"github.com/mailhog/MailHog-Server/config"
	"github.com/mailhog/data"
	"github.com/mailhog/storage"

	"github.com/ian-kent/goose"
)

// APIv1 implements version 1 of the MailHog API
//
// The specification has been frozen and will eventually be deprecated.
// Only bug fixes and non-breaking changes will be applied here.
//
// Any changes/additions should be added in APIv2.
type APIv1 struct {
	config      *config.Config
	messageChan chan *data.Message
}

// FIXME should probably move this into APIv1 struct
var stream *goose.EventStream

// ReleaseConfig is an alias to preserve go package API
type ReleaseConfig config.OutgoingSMTP

func createAPIv1(conf *config.Config, r *pat.Router) *APIv1 {
	log.Println("Creating API v1 with WebPath: " + conf.WebPath)
	apiv1 := &APIv1{
		config:      conf,
		messageChan: make(chan *data.Message),
	}

	stream = goose.NewEventStream()

	r.Path(conf.WebPath + "/api/v1/messages").Methods("GET").HandlerFunc(apiv1.messages)
	r.Path(conf.WebPath + "/api/v1/messages").Methods("DELETE").HandlerFunc(apiv1.delete_all)
	r.Path(conf.WebPath + "/api/v1/messages").Methods("OPTIONS").HandlerFunc(apiv1.defaultOptions)

	r.Path(conf.WebPath + "/api/v1/messages/{id}").Methods("GET").HandlerFunc(apiv1.message)
	r.Path(conf.WebPath + "/api/v1/messages/{id}").Methods("DELETE").HandlerFunc(apiv1.delete_one)
	r.Path(conf.WebPath + "/api/v1/messages/{id}").Methods("OPTIONS").HandlerFunc(apiv1.defaultOptions)

	r.Path(conf.WebPath + "/api/v1/messages/{id}/download").Methods("GET").HandlerFunc(apiv1.download)
	r.Path(conf.WebPath + "/api/v1/messages/{id}/download").Methods("OPTIONS").HandlerFunc(apiv1.defaultOptions)

	r.Path(conf.WebPath + "/api/v1/messages/{id}/mime/part/{part}/download").Methods("GET").HandlerFunc(apiv1.download_part)
	r.Path(conf.WebPath + "/api/v1/messages/{id}/mime/part/{part}/download").Methods("OPTIONS").HandlerFunc(apiv1.defaultOptions)

	r.Path(conf.WebPath + "/api/v1/messages/{id}/release").Methods("POST").HandlerFunc(apiv1.release_one)
	r.Path(conf.WebPath + "/api/v1/messages/{id}/release").Methods("OPTIONS").HandlerFunc(apiv1.defaultOptions)

	r.Path(conf.WebPath + "/api/v1/events").Methods("GET").HandlerFunc(apiv1.eventstream)
	r.Path(conf.WebPath + "/api/v1/events").Methods("OPTIONS").HandlerFunc(apiv1.defaultOptions)

	go func() {
		keepaliveTicker := time.Tick(time.Minute)
		for {
			select {
			case msg := <-apiv1.messageChan:
				log.Println("Got message in APIv1 event stream")
				bytes, _ := json.MarshalIndent(msg, "", "  ")
				json := string(bytes)
				log.Printf("Sending content: %s\n", json)
				apiv1.broadcast(json)
			case <-keepaliveTicker:
				apiv1.keepalive()
			}
		}
	}()

	return apiv1
}

func (apiv1 *APIv1) defaultOptions(w http.ResponseWriter, req *http.Request) {
	if len(apiv1.config.CORSOrigin) > 0 {
		w.Header().Add("Access-Control-Allow-Origin", apiv1.config.CORSOrigin)
		w.Header().Add("Access-Control-Allow-Methods", "OPTIONS,GET,POST,DELETE")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	}
}

func (apiv1 *APIv1) broadcast(json string) {
	log.Println("[APIv1] BROADCAST /api/v1/events")
	b := []byte(json)
	stream.Notify("data", b)
}

// keepalive sends an empty keep alive message.
//
// This not only can keep connections alive, but also will detect broken
// connections. Without this it is possible for the server to become
// unresponsive due to too many open files.
func (apiv1 *APIv1) keepalive() {
	log.Println("[APIv1] KEEPALIVE /api/v1/events")
	stream.Notify("keepalive", []byte{})
}

func (apiv1 *APIv1) eventstream(w http.ResponseWriter, req *http.Request) {
	log.Println("[APIv1] GET /api/v1/events")

	//apiv1.defaultOptions(session)
	if len(apiv1.config.CORSOrigin) > 0 {
		w.Header().Add("Access-Control-Allow-Origin", apiv1.config.CORSOrigin)
		w.Header().Add("Access-Control-Allow-Methods", "OPTIONS,GET,POST,DELETE")
	}

	stream.AddReceiver(w)
}

func (apiv1 *APIv1) messages(w http.ResponseWriter, req *http.Request) {
	log.Println("[APIv1] GET /api/v1/messages")

	apiv1.defaultOptions(w, req)

	// TODO start, limit
	switch apiv1.config.Storage.(type) {
	case *storage.MongoDB:
		messages, _ := apiv1.config.Storage.(*storage.MongoDB).List(0, 1000)
		bytes, _ := json.Marshal(messages)
		w.Header().Add("Content-Type", "text/json")
		w.Write(bytes)
	case *storage.InMemory:
		messages, _ := apiv1.config.Storage.(*storage.InMemory).List(0, 1000)
		bytes, _ := json.Marshal(messages)
		w.Header().Add("Content-Type", "text/json")
		w.Write(bytes)
	default:
		w.WriteHeader(500)
	}
}

func (apiv1 *APIv1) message(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get(":id")
	log.Printf("[APIv1] GET /api/v1/messages/%s\n", id)

	apiv1.defaultOptions(w, req)

	message, err := apiv1.config.Storage.Load(id)
	if err != nil {
		log.Printf("- Error: %s", err)
		w.WriteHeader(500)
		return
	}

	bytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("- Error: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	w.Write(bytes)
}

func (apiv1 *APIv1) download(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get(":id")
	log.Printf("[APIv1] GET /api/v1/messages/%s\n", id)

	apiv1.defaultOptions(w, req)

	w.Header().Set("Content-Type", "message/rfc822")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+id+".eml\"")

	switch apiv1.config.Storage.(type) {
	case *storage.MongoDB:
		message, _ := apiv1.config.Storage.(*storage.MongoDB).Load(id)
		for h, l := range message.Content.Headers {
			for _, v := range l {
				w.Write([]byte(h + ": " + v + "\r\n"))
			}
		}
		w.Write([]byte("\r\n" + message.Content.Body))
	case *storage.InMemory:
		message, _ := apiv1.config.Storage.(*storage.InMemory).Load(id)
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

func (apiv1 *APIv1) download_part(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get(":id")
	part := req.URL.Query().Get(":part")
	log.Printf("[APIv1] GET /api/v1/messages/%s/mime/part/%s/download\n", id, part)

	// TODO extension from content-type?
	apiv1.defaultOptions(w, req)

	w.Header().Set("Content-Disposition", "attachment; filename=\""+id+"-part-"+part+"\"")

	message, _ := apiv1.config.Storage.Load(id)
	contentTransferEncoding := ""
	pid, _ := strconv.Atoi(part)
	for h, l := range message.MIME.Parts[pid].Headers {
		for _, v := range l {
			switch strings.ToLower(h) {
			case "content-disposition":
				// Prevent duplicate "content-disposition"
				w.Header().Set(h, v)
			case "content-transfer-encoding":
				if contentTransferEncoding == "" {
					contentTransferEncoding = v
				}
				fallthrough
			default:
				w.Header().Add(h, v)
			}
		}
	}
	body := []byte(message.MIME.Parts[pid].Body)
	if strings.ToLower(contentTransferEncoding) == "base64" {
		var e error
		body, e = base64.StdEncoding.DecodeString(message.MIME.Parts[pid].Body)
		if e != nil {
			log.Printf("[APIv1] Decoding base64 encoded body failed: %s", e)
		}
	}
	w.Write(body)
}

func (apiv1 *APIv1) delete_all(w http.ResponseWriter, req *http.Request) {
	log.Println("[APIv1] POST /api/v1/messages")

	apiv1.defaultOptions(w, req)

	w.Header().Add("Content-Type", "text/json")

	err := apiv1.config.Storage.DeleteAll()
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
}

func (apiv1 *APIv1) release_one(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get(":id")
	log.Printf("[APIv1] POST /api/v1/messages/%s/release\n", id)

	apiv1.defaultOptions(w, req)

	w.Header().Add("Content-Type", "text/json")
	msg, _ := apiv1.config.Storage.Load(id)

	decoder := json.NewDecoder(req.Body)
	var cfg ReleaseConfig
	err := decoder.Decode(&cfg)
	if err != nil {
		log.Printf("Error decoding request body: %s", err)
		w.WriteHeader(500)
		w.Write([]byte("Error decoding request body"))
		return
	}

	log.Printf("%+v", cfg)

	log.Printf("Got message: %s", msg.ID)

	if cfg.Save {
		if _, ok := apiv1.config.OutgoingSMTP[cfg.Name]; ok {
			log.Printf("Server already exists named %s", cfg.Name)
			w.WriteHeader(400)
			return
		}
		cf := config.OutgoingSMTP(cfg)
		apiv1.config.OutgoingSMTP[cfg.Name] = &cf
		log.Printf("Saved server with name %s", cfg.Name)
	}

	if len(cfg.Name) > 0 {
		if c, ok := apiv1.config.OutgoingSMTP[cfg.Name]; ok {
			log.Printf("Using server with name: %s", cfg.Name)
			cfg.Name = c.Name
			if len(cfg.Email) == 0 {
				cfg.Email = c.Email
			}
			cfg.Host = c.Host
			cfg.Port = c.Port
			cfg.Username = c.Username
			cfg.Password = c.Password
			cfg.Mechanism = c.Mechanism
		} else {
			log.Printf("Server not found: %s", cfg.Name)
			w.WriteHeader(400)
			return
		}
	}

	log.Printf("Releasing to %s (via %s:%s)", cfg.Email, cfg.Host, cfg.Port)

	bytes := make([]byte, 0)
	for h, l := range msg.Content.Headers {
		for _, v := range l {
			bytes = append(bytes, []byte(h+": "+v+"\r\n")...)
		}
	}
	bytes = append(bytes, []byte("\r\n"+msg.Content.Body)...)

	var auth smtp.Auth

	if len(cfg.Username) > 0 || len(cfg.Password) > 0 {
		log.Printf("Found username/password, using auth mechanism: [%s]", cfg.Mechanism)
		switch cfg.Mechanism {
		case "CRAMMD5":
			auth = smtp.CRAMMD5Auth(cfg.Username, cfg.Password)
		case "PLAIN":
			auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
		default:
			log.Printf("Error - invalid authentication mechanism")
			w.WriteHeader(400)
			return
		}
	}

	err = smtp.SendMail(cfg.Host+":"+cfg.Port, auth, "nobody@"+apiv1.config.Hostname, []string{cfg.Email}, bytes)
	if err != nil {
		log.Printf("Failed to release message: %s", err)
		w.WriteHeader(500)
		return
	}
	log.Printf("Message released successfully")
}

func (apiv1 *APIv1) delete_one(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get(":id")

	log.Printf("[APIv1] POST /api/v1/messages/%s/delete\n", id)

	apiv1.defaultOptions(w, req)

	w.Header().Add("Content-Type", "text/json")
	err := apiv1.config.Storage.DeleteOne(id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}

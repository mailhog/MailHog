package api

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/pat"
	"github.com/ian-kent/go-log/log"
	"github.com/mailhog/MailHog-Server/config"
	"github.com/mailhog/MailHog-Server/websockets"
	"github.com/mailhog/data"
	"github.com/mailhog/storage"
)

// APIv3 is optimized for large Maildir-backed message stores.
//
// List and search endpoints return compact message metadata. Full message
// payloads, bounded body previews and downloads are fetched by message ID.
type APIv3 struct {
	config      *config.Config
	messageChan chan *data.Message
	wsHub       *websockets.Hub
}

func createAPIv3(conf *config.Config, r *pat.Router) *APIv3 {
	log.Println("Creating API v3 with WebPath: " + conf.WebPath)
	apiv3 := &APIv3{
		config:      conf,
		messageChan: make(chan *data.Message),
		wsHub:       websockets.NewHub(),
	}

	r.Path(conf.WebPath + "/api/v3/messages").Methods("GET").HandlerFunc(apiv3.messages)
	r.Path(conf.WebPath + "/api/v3/messages").Methods("DELETE").HandlerFunc(apiv3.deleteMessages)
	r.Path(conf.WebPath + "/api/v3/messages").Methods("OPTIONS").HandlerFunc(apiv3.defaultOptions)

	r.Path(conf.WebPath + "/api/v3/messages/{id}").Methods("GET").HandlerFunc(apiv3.message)
	r.Path(conf.WebPath + "/api/v3/messages/{id}").Methods("OPTIONS").HandlerFunc(apiv3.defaultOptions)

	r.Path(conf.WebPath + "/api/v3/messages/{id}/body").Methods("GET").HandlerFunc(apiv3.messageBody)
	r.Path(conf.WebPath + "/api/v3/messages/{id}/body").Methods("OPTIONS").HandlerFunc(apiv3.defaultOptions)

	r.Path(conf.WebPath + "/api/v3/messages/{id}/download").Methods("GET").HandlerFunc(apiv3.download)
	r.Path(conf.WebPath + "/api/v3/messages/{id}/download").Methods("OPTIONS").HandlerFunc(apiv3.defaultOptions)

	r.Path(conf.WebPath + "/api/v3/search").Methods("GET").HandlerFunc(apiv3.search)
	r.Path(conf.WebPath + "/api/v3/search").Methods("OPTIONS").HandlerFunc(apiv3.defaultOptions)

	r.Path(conf.WebPath + "/api/v3/websocket").Methods("GET").HandlerFunc(apiv3.websocket)

	go func() {
		for {
			select {
			case msg := <-apiv3.messageChan:
				log.Println("Got message in APIv3 websocket channel")
				apiv3.broadcast(msg)
			}
		}
	}()

	return apiv3
}

func (apiv3 *APIv3) defaultOptions(w http.ResponseWriter, req *http.Request) {
	if len(apiv3.config.CORSOrigin) > 0 {
		w.Header().Add("Access-Control-Allow-Origin", apiv3.config.CORSOrigin)
		w.Header().Add("Access-Control-Allow-Methods", "OPTIONS,GET,POST,DELETE")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	}
}

func (apiv3 *APIv3) messages(w http.ResponseWriter, req *http.Request) {
	log.Println("[APIv3] GET /api/v3/messages")

	apiv3.defaultOptions(w, req)

	start, limit := getStartLimit(req)

	var res messagesResult

	messages, err := apiv3.config.Storage.List(start, limit)
	if err != nil {
		panic(err)
	}

	res.Count = len([]data.Message(*messages))
	res.Start = start
	res.Items = compactMessages(*messages)
	res.Total = apiv3.config.Storage.Count()

	writeJSON(w, res)
}

func (apiv3 *APIv3) search(w http.ResponseWriter, req *http.Request) {
	log.Println("[APIv3] GET /api/v3/search")

	apiv3.defaultOptions(w, req)

	start, limit := getStartLimit(req)

	kind := req.URL.Query().Get("kind")
	if kind != "from" && kind != "to" && kind != "containing" {
		w.WriteHeader(400)
		return
	}

	query := req.URL.Query().Get("query")
	if len(query) == 0 {
		w.WriteHeader(400)
		return
	}

	var res messagesResult

	messages, total, _ := apiv3.config.Storage.Search(kind, query, start, limit)

	res.Count = len([]data.Message(*messages))
	res.Start = start
	res.Items = compactMessages(*messages)
	res.Total = total

	writeJSON(w, res)
}

func (apiv3 *APIv3) message(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get(":id")
	log.Printf("[APIv3] GET /api/v3/messages/%s\n", id)

	apiv3.defaultOptions(w, req)

	var message *data.Message
	var err error
	if previewStorage, ok := apiv3.config.Storage.(storage.PreviewStorage); ok {
		message, err = previewStorage.LoadPreview(id, previewMaxMessageSize)
	} else {
		message, err = apiv3.config.Storage.Load(id)
	}
	if err != nil {
		log.Printf("- Error: %s", err)
		if os.IsNotExist(err) {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	writeJSON(w, message)
}

func (apiv3 *APIv3) messageBody(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get(":id")
	log.Printf("[APIv3] GET /api/v3/messages/%s/body\n", id)

	apiv3.defaultOptions(w, req)

	offset := parseInt64Query(req, "offset", 0)
	limit := parseInt64Query(req, "limit", previewBodyChunkSize)
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = previewBodyChunkSize
	}
	if limit > previewBodyMaxChunkSize {
		limit = previewBodyMaxChunkSize
	}

	var chunk *storage.MessageBodyChunk
	var err error
	if previewStorage, ok := apiv3.config.Storage.(storage.BodyPreviewStorage); ok {
		chunk, err = previewStorage.LoadBodyChunk(id, offset, limit, previewMaxMessageSize)
	} else {
		chunk, err = loadMessageBodyChunk(apiv3.config.Storage, id, offset, limit, previewMaxMessageSize)
	}
	if err != nil {
		log.Printf("- Error: %s", err)
		if os.IsNotExist(err) {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	writeJSON(w, chunk)
}

func (apiv3 *APIv3) download(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get(":id")
	log.Printf("[APIv3] GET /api/v3/messages/%s/download\n", id)

	apiv3.defaultOptions(w, req)

	w.Header().Set("Content-Type", "message/rfc822")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+id+".eml\"")

	if downloader, ok := apiv3.config.Storage.(storage.DownloadStorage); ok {
		if err := downloader.WriteMessageTo(id, w); err != nil {
			log.Printf("- Error: %s", err)
			if os.IsNotExist(err) {
				w.WriteHeader(404)
			} else {
				w.WriteHeader(500)
			}
		}
		return
	}

	message, err := apiv3.config.Storage.Load(id)
	if err != nil {
		log.Printf("- Error: %s", err)
		w.WriteHeader(500)
		return
	}
	for h, l := range message.Content.Headers {
		for _, v := range l {
			w.Write([]byte(h + ": " + v + "\r\n"))
		}
	}
	w.Write([]byte("\r\n" + message.Content.Body))
}

func (apiv3 *APIv3) deleteMessages(w http.ResponseWriter, req *http.Request) {
	log.Println("[APIv3] DELETE /api/v3/messages")

	apiv3.defaultOptions(w, req)

	status, err := deleteMessages(apiv3.config.Storage, req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(status)
		return
	}

	w.WriteHeader(status)
}

func (apiv3 *APIv3) websocket(w http.ResponseWriter, req *http.Request) {
	log.Println("[APIv3] GET /api/v3/websocket")

	apiv3.wsHub.Serve(w, req)
}

func (apiv3 *APIv3) broadcast(msg *data.Message) {
	log.Println("[APIv3] BROADCAST /api/v3/websocket")

	apiv3.wsHub.Broadcast(compactMessage(*msg))
}

func writeJSON(w http.ResponseWriter, value interface{}) {
	bytes, _ := json.Marshal(value)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

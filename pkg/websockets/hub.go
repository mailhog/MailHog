package websockets

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/ian-kent/go-log/log"
)

type Hub struct {
	upgrader       websocket.Upgrader
	connections    map[*connection]bool
	messages       chan interface{}
	registerChan   chan *connection
	unregisterChan chan *connection
}

func NewHub() *Hub {
	hub := &Hub{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  256,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		connections:    make(map[*connection]bool),
		messages:       make(chan interface{}),
		registerChan:   make(chan *connection),
		unregisterChan: make(chan *connection),
	}
	go hub.run()
	return hub
}

func (h *Hub) run() {
	for {
		select {
		case c := <-h.registerChan:
			h.connections[c] = true
		case c := <-h.unregisterChan:
			h.unregister(c)
		case m := <-h.messages:
			for c := range h.connections {
				select {
				case c.send <- m:
				default:
					h.unregister(c)
				}
			}
		}
	}
}

func (h *Hub) unregister(c *connection) {
	if _, ok := h.connections[c]; ok {
		close(c.send)
		delete(h.connections, c)
	}
}

func (h *Hub) Serve(w http.ResponseWriter, r *http.Request) {
	ws, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	c := &connection{hub: h, ws: ws, send: make(chan interface{}, 256)}
	h.registerChan <- c
	go c.writeLoop()
	go c.readLoop()
}

func (h *Hub) Broadcast(data interface{}) {
	h.messages <- data
}

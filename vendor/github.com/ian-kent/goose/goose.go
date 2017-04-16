package goose

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

var (
	// ErrUnableToHijackRequest is returned by AddReceiver if the type
	// conversion to http.Hijacker is unsuccessful
	ErrUnableToHijackRequest = errors.New("Unable to hijack request")
)

// EventStream represents a collection of receivers
type EventStream struct {
	mutex     *sync.Mutex
	receivers map[net.Conn]*EventReceiver
}

// NewEventStream creates a new event stream
func NewEventStream() *EventStream {
	return &EventStream{
		mutex:     new(sync.Mutex),
		receivers: make(map[net.Conn]*EventReceiver),
	}
}

// EventReceiver represents a hijacked HTTP connection
type EventReceiver struct {
	stream *EventStream
	conn   net.Conn
	bufrw  *bufio.ReadWriter
}

// Notify sends the event to all event stream receivers
func (es *EventStream) Notify(event string, bytes []byte) {
	// TODO reader?

	lines := strings.Split(string(bytes), "\n")

	data := ""
	for _, l := range lines {
		data += event + ": " + l + "\n"
	}

	sz := len(data) + 1
	size := fmt.Sprintf("%X", sz)

	for _, er := range es.receivers {
		go er.send(size, data)
	}
}

func (er *EventReceiver) send(size, data string) {
	_, err := er.write([]byte(size + "\r\n"))
	if err != nil {
		return
	}

	lines := strings.Split(data, "\n")
	for _, ln := range lines {
		_, err = er.write([]byte(ln + "\n"))
		if err != nil {
			return
		}
	}
	er.write([]byte("\r\n"))
}

func (er *EventReceiver) write(bytes []byte) (int, error) {
	n, err := er.bufrw.Write(bytes)

	if err != nil {
		er.stream.mutex.Lock()
		delete(er.stream.receivers, er.conn)
		er.stream.mutex.Unlock()
		er.conn.Close()
		return n, err
	}

	err = er.bufrw.Flush()
	if err != nil {
		er.stream.mutex.Lock()
		delete(er.stream.receivers, er.conn)
		er.stream.mutex.Unlock()
		er.conn.Close()
	}

	return n, err
}

// AddReceiver hijacks a http.ResponseWriter and attaches it to the event stream
func (es *EventStream) AddReceiver(w http.ResponseWriter) (*EventReceiver, error) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	w.WriteHeader(200)

	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, ErrUnableToHijackRequest
	}

	hjConn, hjBufrw, err := hj.Hijack()
	if err != nil {
		return nil, err
	}

	rec := &EventReceiver{es, hjConn, hjBufrw}

	es.mutex.Lock()
	es.receivers[hjConn] = rec
	es.mutex.Unlock()

	return rec, nil
}

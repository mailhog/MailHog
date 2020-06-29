package web

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/pat"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// Authorised should be given a function to enable HTTP Basic Authentication
var Authorised func(string, string) bool
var users map[string]string

// AuthFile sets Authorised to a function which validates against file
func AuthFile(file string) {
	users = make(map[string]string)

	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("[HTTP] Error reading auth-file: %s", err)
	}

	buf := bytes.NewBuffer(b)

	for {
		l, err := buf.ReadString('\n')
		l = strings.TrimSpace(l)
		if len(l) > 0 {
			p := strings.SplitN(l, ":", 2)
			if len(p) < 2 {
				panic(fmt.Errorf("[HTTP] Error reading auth-file, invalid line: %s", l))
			}
			users[p[0]] = p[1]
		}
		switch {
		case err == io.EOF:
			break
		case err != nil:
			panic(fmt.Errorf("[HTTP] Error reading auth-file: %s", err))
		}
		if err == io.EOF {
			break
		}
	}

	log.Infof("[HTTP] Loaded %d users from %s", len(users), file)

	Authorised = func(u, pw string) bool {
		hpw, ok := users[u]
		if !ok {
			return false
		}

		err := bcrypt.CompareHashAndPassword([]byte(hpw), []byte(pw))
		return err == nil
	}
}

// BasicAuthHandler is middleware to check HTTP Basic Authentication
// if an authorisation function is defined.
func BasicAuthHandler(h http.Handler) http.Handler {
	f := func(w http.ResponseWriter, req *http.Request) {
		if Authorised == nil {
			h.ServeHTTP(w, req)
			return
		}

		u, pw, ok := req.BasicAuth()
		if !ok || !Authorised(u, pw) {
			w.Header().Set("WWW-Authenticate", "Basic")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, req)
	}

	return http.HandlerFunc(f)
}

// Listen binds to httpBindAddr
func Listen(httpBindAddr string, _ func(string) ([]byte, error), exitCh chan int, registerCallback func(http.Handler)) {
	log.Infof("[HTTP] Binding to address: %s", httpBindAddr)

	pat := pat.New()
	registerCallback(pat)

	//compress := handlers.CompressHandler(pat)
	auth := BasicAuthHandler(pat) //compress)

	err := http.ListenAndServe(httpBindAddr, auth)
	if err != nil {
		log.Errorf("[HTTP] Error binding to address %s: %s", httpBindAddr, err)
	}
}

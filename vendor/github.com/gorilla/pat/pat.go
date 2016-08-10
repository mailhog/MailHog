// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pat

import (
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

// New returns a new router.
func New() *Router {
	return &Router{}
}

// Router is a request router that implements a pat-like API.
//
// pat docs: http://godoc.org/github.com/bmizerany/pat
type Router struct {
	mux.Router
}

// Add registers a pattern with a handler for the given request method.
func (r *Router) Add(meth, pat string, h http.Handler) *mux.Route {
	return r.NewRoute().PathPrefix(pat).Handler(h).Methods(meth)
}

// Options registers a pattern with a handler for OPTIONS requests.
func (r *Router) Options(pat string, h http.HandlerFunc) *mux.Route {
	return r.Add("OPTIONS", pat, h)
}

// Delete registers a pattern with a handler for DELETE requests.
func (r *Router) Delete(pat string, h http.HandlerFunc) *mux.Route {
	return r.Add("DELETE", pat, h)
}

// Head registers a pattern with a handler for HEAD requests.
func (r *Router) Head(pat string, h http.HandlerFunc) *mux.Route {
	return r.Add("HEAD", pat, h)
}

// Get registers a pattern with a handler for GET requests.
func (r *Router) Get(pat string, h http.HandlerFunc) *mux.Route {
	return r.Add("GET", pat, h)
}

// Post registers a pattern with a handler for POST requests.
func (r *Router) Post(pat string, h http.HandlerFunc) *mux.Route {
	return r.Add("POST", pat, h)
}

// Put registers a pattern with a handler for PUT requests.
func (r *Router) Put(pat string, h http.HandlerFunc) *mux.Route {
	return r.Add("PUT", pat, h)
}

// Patch registers a pattern with a handler for PATCH requests.
func (r *Router) Patch(pat string, h http.HandlerFunc) *mux.Route {
	return r.Add("PATCH", pat, h)
}

// ServeHTTP dispatches the handler registered in the matched route.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Clean path to canonical form and redirect.
	if p := cleanPath(req.URL.Path); p != req.URL.Path {
		w.Header().Set("Location", p)
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}
	var match mux.RouteMatch
	var handler http.Handler
	if matched := r.Match(req, &match); matched {
		handler = match.Handler
		registerVars(req, match.Vars)
	}
	if handler == nil {
		if r.NotFoundHandler == nil {
			r.NotFoundHandler = http.NotFoundHandler()
		}
		handler = r.NotFoundHandler
	}
	if !r.KeepContext {
		defer context.Clear(req)
	}
	handler.ServeHTTP(w, req)
}

// registerVars adds the matched route variables to the URL query.
func registerVars(r *http.Request, vars map[string]string) {
	parts, i := make([]string, len(vars)), 0
	for key, value := range vars {
		parts[i] = url.QueryEscape(":"+key) + "=" + url.QueryEscape(value)
		i++
	}
	q := strings.Join(parts, "&")
	if r.URL.RawQuery == "" {
		r.URL.RawQuery = q
	} else {
		r.URL.RawQuery += "&" + q
	}
}

// cleanPath returns the canonical path for p, eliminating . and .. elements.
// Borrowed from the net/http package.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

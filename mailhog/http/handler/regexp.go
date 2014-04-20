package handler;

import (
    "regexp"
    "net/http"
)

// http://stackoverflow.com/questions/6564558/wildcards-in-the-pattern-for-http-handlefunc

type Route struct {
    Pattern *regexp.Regexp
    Handler HandlerFunc
}

type HandlerFunc func(http.ResponseWriter, *http.Request, *Route)

//type Handler http.Handler 
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, route *Route) {
    f(w, r, route)
}

type RegexpHandler struct {
    routes []*Route
}

func (h *RegexpHandler) Handler(pattern *regexp.Regexp, handler HandlerFunc) {
    h.routes = append(h.routes, &Route{pattern, handler})
}

func (h *RegexpHandler) HandleFunc(pattern *regexp.Regexp, handler func(http.ResponseWriter, *http.Request, *Route)) {
    h.routes = append(h.routes, &Route{pattern, HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    for _, route := range h.routes {
        if route.Pattern.MatchString(r.URL.Path) {
            route.Handler.ServeHTTP(w, r, route)
            return
        }
    }
    // no pattern matched; send 404 response
    http.NotFound(w, r)
}

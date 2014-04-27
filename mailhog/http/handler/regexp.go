package handler;

import (
    "regexp"
    "net/http"
)

// http://stackoverflow.com/questions/6564558/wildcards-in-the-pattern-for-http-handlefunc

type Route struct {
    Methods map[string]int
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

func (h *RegexpHandler) Handler(methods []string, pattern *regexp.Regexp, handler HandlerFunc) {
    m := make(map[string]int,0)
    for _, v := range methods {
        m[v] = 1
    }
    h.routes = append(h.routes, &Route{m, pattern, handler})
}

func (h *RegexpHandler) HandleFunc(methods []string, pattern *regexp.Regexp, handler func(http.ResponseWriter, *http.Request, *Route)) {
    m := make(map[string]int,0)
    for _, v := range methods {
        m[v] = 1
    }
    h.routes = append(h.routes, &Route{m, pattern, HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    for _, route := range h.routes {
        if route.Pattern.MatchString(r.URL.Path) {
            _, ok := route.Methods[r.Method]
            if ok {
                route.Handler.ServeHTTP(w, r, route)
                return
            }
        }
    }
    // no pattern matched; send 404 response
    http.NotFound(w, r)
}

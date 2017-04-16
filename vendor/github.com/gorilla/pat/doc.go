// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gorilla/pat is a request router and dispatcher with a pat-like
interface. It is an alternative to gorilla/mux that showcases how it can
be used as a base for different API flavors. Package pat is documented at:

	http://godoc.org/github.com/bmizerany/pat

Let's start registering a couple of URL paths and handlers:

	func main() {
		r := pat.New()
		r.Get("/products", ProductsHandler)
		r.Get("/articles", ArticlesHandler)
		r.Get("/", HomeHandler)
		http.Handle("/", r)
	}

Here we register three routes mapping URL paths to handlers. This is
equivalent to how http.HandleFunc() works: if an incoming GET request matches
one of the paths, the corresponding handler is called passing
(http.ResponseWriter, *http.Request) as parameters.

Note: gorilla/pat matches path prefixes, so you must register the most
specific paths first.

Note: differently from pat, these methods accept a handler function, and not an
http.Handler. We think this is shorter and more convenient. To set an
http.Handler, use the Add() method.

Paths can have variables. They are defined using the format {name} or
{name:pattern}. If a regular expression pattern is not defined, the matched
variable will be anything until the next slash. For example:

	r := pat.New()
	r.Get("/articles/{category}/{id:[0-9]+}", ArticleHandler)
	r.Get("/articles/{category}/", ArticlesCategoryHandler)
	r.Get("/products/{key}", ProductHandler)

The names are used to create a map of route variables which are stored in the
URL query, prefixed by a colon:

	category := req.URL.Query().Get(":category")

As in the gorilla/mux package, other matchers can be added to the registered
routes and URLs can be reversed as well. To build a URL for a route, first
add a name to it:

	r.Get("/products/{key}", ProductHandler).Name("product")

Then you can get it using the name and generate a URL:

	url, err := r.GetRoute("product").URL("key", "transmogrifier")

...and the result will be a url.URL with the following path:

	"/products/transmogrifier"

Check the mux documentation for more details about URL building and extra
matchers:

	http://gorilla-web.appspot.com/pkg/mux/
*/
package pat

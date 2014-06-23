package http

import (
	"github.com/ian-kent/gotcha/http"
	"html/template"
)

func Index(session *http.Session) {
	html, _ := session.RenderTemplate("index.html")

	session.Stash["Page"] = "Browse"
	session.Stash["Content"] = template.HTML(html)
	session.Render("layout.html")
}

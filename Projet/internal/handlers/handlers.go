package handlers

import (
	"database/sql"
	"html/template"
)

type Handler struct {
	DB        *sql.DB
	Templates map[string]*template.Template
}

func New(db *sql.DB) *Handler {
	funcMap := template.FuncMap{
		"min": func(a, b int) int {
			if a < b {
				return a
			}
			return b
		},
	}

	templates := make(map[string]*template.Template)
	pages := []string{"index", "login", "register", "post", "create_post", "error", "admin", "profile", "edit_profile"}

	for _, page := range pages {
		tmpl := template.Must(
			template.New("base").Funcs(funcMap).ParseFiles(
				"web/templates/base.html",
				"web/templates/"+page+".html",
			),
		)
		templates[page] = tmpl
	}

	return &Handler{DB: db, Templates: templates}
}

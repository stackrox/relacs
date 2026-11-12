package serve

import (
	"bytes"
	"html/template"
	"log"
	"net/http"

	"github.com/stackrox/relacs/pkg/client"
	"github.com/stackrox/relacs/pkg/config"
	"github.com/stackrox/relacs/pkg/serve"
)

func renderTemplate(w http.ResponseWriter, tmpl *template.Template, data any) {
	var buf bytes.Buffer
	// Start rendering from the outermost template; it pulls in child-defined blocks.
	if err := tmpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err := buf.WriteTo(w)
	if err != nil {
		log.Printf("Failed to write template to response: %v", err)
	}
}

func handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, homeTmpl, nil)
	}
}

func handleTag(c *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tag := r.URL.Query().Get("tag")
		if tag == "" {
			http.Error(w, "Missing required query parameter: tag", http.StatusBadRequest)
			return
		}
		repo := r.URL.Query().Get("repo")
		if repo == "" {
			repo = config.MainGitHubRepository
		}

		data := serve.FetchTagData(r.Context(), c, repo, tag)
		renderTemplate(w, tagTmpl, data)
	}
}

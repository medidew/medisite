package web

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/medidew/medisite/internal/content"
	"go.uber.org/zap"
)

// pageNames are the templates/{name}.html.tmpl files parsed alongside base.html.tmpl.
var pageNames = []string{"home", "blog_index", "article", "portfolio", "404"}

// Server holds the loaded site content and pre-parsed templates, and
// implements http.Handler via its embedded mux.
type Server struct {
	site   *content.Site
	pages  map[string]*template.Template
	mux    *http.ServeMux
	logger *zap.Logger
}

// New builds a Server, parsing all templates and registering routes.
// templatesDir and staticDir must exist; a missing or unparseable template
// is a startup error since it indicates broken wiring, not a content typo.
func New(site *content.Site, templatesDir, staticDir string, logger *zap.Logger) (*Server, error) {
	if _, err := os.Stat(templatesDir); err != nil {
		return nil, fmt.Errorf("templates directory: %w", err)
	}
	if _, err := os.Stat(staticDir); err != nil {
		return nil, fmt.Errorf("static directory: %w", err)
	}

	base, err := template.ParseFiles(filepath.Join(templatesDir, "base.html.tmpl"))
	if err != nil {
		return nil, fmt.Errorf("parsing base template: %w", err)
	}

	pages := make(map[string]*template.Template, len(pageNames))
	for _, name := range pageNames {
		clone, err := base.Clone()
		if err != nil {
			return nil, fmt.Errorf("cloning base template for %q: %w", name, err)
		}
		page, err := clone.ParseFiles(filepath.Join(templatesDir, name+".html.tmpl"))
		if err != nil {
			return nil, fmt.Errorf("parsing template %q: %w", name, err)
		}
		pages[name] = page
	}

	s := &Server{site: site, pages: pages, logger: logger}
	s.mux = newMux(s, staticDir)
	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

package web

import "net/http"

// newMux builds the route table for s. "GET /{$}" matches the root path
// exactly (a bare "/" would be a subtree match and swallow the catch-all
// below); the trailing "/" pattern is the standard stdlib idiom for a
// catch-all 404 since ServeMux has no built-in NotFound hook.
func newMux(s *Server, staticDir string) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", s.handleHome)
	mux.HandleFunc("GET /blog", s.handleBlogIndex)
	mux.HandleFunc("GET /blog/{slug}", s.handleArticle)
	mux.HandleFunc("GET /portfolio", s.handlePortfolio)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	mux.HandleFunc("/", s.handleNotFound)

	return mux
}

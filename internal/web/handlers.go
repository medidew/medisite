package web

import (
	"net/http"

	"github.com/medidew/medisite/internal/content"
	"go.uber.org/zap"
)

const siteName = "medidew.name"

type homeData struct {
	Title  string
	Recent []*content.Post
}

type blogIndexData struct {
	Title string
	Posts []*content.Post
}

type articleData struct {
	Title string
	Post  *content.Post
}

type portfolioData struct {
	Title    string
	Projects []content.Project
}

type notFoundData struct {
	Title string
}

func (s *Server) render(w http.ResponseWriter, page string, status int, data any) {
	tmpl, ok := s.pages[page]
	if !ok {
		s.logger.Error("no such page template", zap.String("page", page))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		s.logger.Error("rendering template", zap.String("page", page), zap.Error(err))
	}
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	recent := s.site.Posts
	if len(recent) > 3 {
		recent = recent[:3]
	}
	s.render(w, "home", http.StatusOK, homeData{
		Title:  siteName,
		Recent: recent,
	})
}

func (s *Server) handleBlogIndex(w http.ResponseWriter, r *http.Request) {
	s.render(w, "blog_index", http.StatusOK, blogIndexData{
		Title: "Blog — " + siteName,
		Posts: s.site.Posts,
	})
}

func (s *Server) handleArticle(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	post, ok := s.site.PostsBySlug[slug]
	if !ok {
		s.handleNotFound(w, r)
		return
	}
	s.render(w, "article", http.StatusOK, articleData{
		Title: post.Title + " — " + siteName,
		Post:  post,
	})
}

func (s *Server) handlePortfolio(w http.ResponseWriter, r *http.Request) {
	s.render(w, "portfolio", http.StatusOK, portfolioData{
		Title:    "Portfolio — " + siteName,
		Projects: s.site.Projects,
	})
}

func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	s.render(w, "404", http.StatusNotFound, notFoundData{
		Title: "Not Found — " + siteName,
	})
}

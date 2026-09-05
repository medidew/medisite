package content

import (
	"html/template"
	"time"
)

// Post is a single parsed blog post.
type Post struct {
	Slug    string
	Title   string
	Date    time.Time
	Tags    []string
	Summary string
	Content template.HTML
}

// Project is a single portfolio entry.
type Project struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Link        string   `yaml:"link"`
	Tags        []string `yaml:"tags"`
}

// Site is the whole site's content, loaded once at startup.
type Site struct {
	Posts       []*Post
	PostsBySlug map[string]*Post
	Projects    []Project
}

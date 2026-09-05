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
	// Body is the raw Markdown source (post-frontmatter), kept around for
	// search — matching against this avoids false hits on HTML markup that
	// searching the rendered Content would produce.
	Body string
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

package content

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const dateLayout = "2006-01-02"

type frontmatter struct {
	Title   string   `yaml:"title"`
	Date    string   `yaml:"date"`
	Tags    []string `yaml:"tags"`
	Summary string   `yaml:"summary"`
	Slug    string   `yaml:"slug"`
}

// LoadPosts reads every *.md file in dir, parsing its YAML frontmatter and
// Markdown body. A single malformed post is logged and skipped rather than
// failing the whole load.
func LoadPosts(dir string, logger *zap.Logger) ([]*Post, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading posts directory %q: %w", dir, err)
	}

	var posts []*Post
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		post, err := parsePost(path, entry.Name())
		if err != nil {
			logger.Warn("skipping post", zap.String("path", path), zap.Error(err))
			continue
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func parsePost(path, filename string) (*Post, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	const delim = "---\n"
	if !bytes.HasPrefix(raw, []byte(delim)) {
		return nil, fmt.Errorf("missing leading %q frontmatter delimiter", "---")
	}
	rest := raw[len(delim):]
	end := bytes.Index(rest, []byte("\n---\n"))
	if end == -1 {
		return nil, fmt.Errorf("missing closing %q frontmatter delimiter", "---")
	}
	fmBlock := rest[:end]
	body := rest[end+len("\n---\n"):]

	var fm frontmatter
	if err := yaml.Unmarshal(fmBlock, &fm); err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}
	if fm.Title == "" {
		return nil, fmt.Errorf("frontmatter missing required field %q", "title")
	}
	if fm.Date == "" {
		return nil, fmt.Errorf("frontmatter missing required field %q", "date")
	}
	date, err := time.Parse(dateLayout, fm.Date)
	if err != nil {
		return nil, fmt.Errorf("parsing date %q: %w", fm.Date, err)
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(body, &buf); err != nil {
		return nil, fmt.Errorf("rendering markdown: %w", err)
	}

	slug := fm.Slug
	if slug == "" {
		slug = strings.TrimSuffix(filename, ".md")
	}

	return &Post{
		Slug:    slug,
		Title:   fm.Title,
		Date:    date,
		Tags:    fm.Tags,
		Summary: fm.Summary,
		Content: template.HTML(buf.String()),
	}, nil
}

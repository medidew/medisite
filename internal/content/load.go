package content

import (
	"fmt"
	"path/filepath"
	"sort"

	"go.uber.org/zap"
)

// Load reads all posts and portfolio data under contentDir and builds a Site.
func Load(contentDir string, logger *zap.Logger) (*Site, error) {
	posts, err := LoadPosts(filepath.Join(contentDir, "posts"), logger)
	if err != nil {
		return nil, err
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	postsBySlug := make(map[string]*Post, len(posts))
	for _, p := range posts {
		if existing, ok := postsBySlug[p.Slug]; ok {
			return nil, fmt.Errorf("duplicate post slug %q (titles %q and %q)", p.Slug, existing.Title, p.Title)
		}
		postsBySlug[p.Slug] = p
	}

	projects, err := LoadPortfolio(filepath.Join(contentDir, "portfolio.yaml"))
	if err != nil {
		return nil, err
	}

	return &Site{
		Posts:       posts,
		PostsBySlug: postsBySlug,
		Projects:    projects,
	}, nil
}

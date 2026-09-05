package content

import "strings"

// Search returns every post whose title, summary, tags, or body contains
// query, case-insensitively. Posts are returned in Site.Posts order
// (newest first). An empty (post-trim) query matches nothing.
func (s *Site) Search(query string) []*Post {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	query = strings.ToLower(query)

	var results []*Post
	for _, p := range s.Posts {
		if postMatches(p, query) {
			results = append(results, p)
		}
	}
	return results
}

func postMatches(p *Post, lowerQuery string) bool {
	if strings.Contains(strings.ToLower(p.Title), lowerQuery) {
		return true
	}
	if strings.Contains(strings.ToLower(p.Summary), lowerQuery) {
		return true
	}
	if strings.Contains(strings.ToLower(p.Body), lowerQuery) {
		return true
	}
	for _, tag := range p.Tags {
		if strings.Contains(strings.ToLower(tag), lowerQuery) {
			return true
		}
	}
	return false
}

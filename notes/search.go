package notes

import (
	"strings"
	"time"
)

// SearchResult represents a single search hit.
type SearchResult struct {
	Date    time.Time
	Line    string
	LineNum int
}

// Search finds all lines matching the query across all day files.
func (s *Store) Search(query string) ([]SearchResult, error) {
	days, err := s.ListDays()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []SearchResult

	for _, d := range days {
		content, err := s.ReadDay(d)
		if err != nil {
			continue
		}

		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), query) {
				// Skip header lines from results
				if strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") {
					continue
				}
				trimmed := strings.TrimSpace(line)
				if trimmed == "" {
					continue
				}
				results = append(results, SearchResult{
					Date:    d,
					Line:    trimmed,
					LineNum: i + 1,
				})
			}
		}
	}

	return results, nil
}

// GetOpenTasks returns all unchecked tasks across all days.
func (s *Store) GetOpenTasks() ([]SearchResult, error) {
	days, err := s.ListDays()
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	for _, d := range days {
		content, err := s.ReadDay(d)
		if err != nil {
			continue
		}

		lines := strings.Split(content, "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- [ ] ") {
				results = append(results, SearchResult{
					Date:    d,
					Line:    trimmed,
					LineNum: i + 1,
				})
			}
		}
	}

	return results, nil
}

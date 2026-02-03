package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Entry represents a single timestamped note entry.
type Entry struct {
	Time    time.Time
	Content string
	IsTask  bool
	Done    bool
}

// DayNote represents all entries for a single day.
type DayNote struct {
	Date    time.Time
	Entries []Entry
}

// Store manages reading and writing day-files.
type Store struct {
	Dir string
}

// NewStore creates a new Store with the given notes directory.
func NewStore(dir string) *Store {
	return &Store{Dir: dir}
}

// EnsureDir creates the notes directory if it doesn't exist.
func (s *Store) EnsureDir() error {
	return os.MkdirAll(s.Dir, 0755)
}

// PathForDate returns the file path for a given date.
func (s *Store) PathForDate(d time.Time) string {
	return filepath.Join(s.Dir, d.Format("2006-01-02")+".md")
}

// Today returns today's date at midnight.
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// AppendEntry adds a new timestamped entry to the given day's file.
func (s *Store) AppendEntry(d time.Time, content string, isTask bool) error {
	if err := s.EnsureDir(); err != nil {
		return err
	}

	path := s.PathForDate(d)
	existed := true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		existed = false
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write day header if new file
	if !existed {
		header := fmt.Sprintf("# %s\n\n", d.Format("Monday, January 2, 2006"))
		if _, err := f.WriteString(header); err != nil {
			return err
		}
	}

	var line string
	if isTask {
		line = fmt.Sprintf("- [ ] %s\n", content)
	} else {
		line = fmt.Sprintf("%s\n", content)
	}

	_, err = f.WriteString(line)
	return err
}

// ReadDay reads all content for a given date. Returns empty string if no file exists.
func (s *Store) ReadDay(d time.Time) (string, error) {
	path := s.PathForDate(d)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ListDays returns all available dates (sorted newest first).
func (s *Store) ListDays() ([]time.Time, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var days []time.Time
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		t, err := time.Parse("2006-01-02", name)
		if err != nil {
			continue
		}
		days = append(days, t)
	}

	sort.Slice(days, func(i, j int) bool {
		return days[i].After(days[j])
	})

	return days, nil
}

// LoadStream loads the last N days of notes stitched together (oldest first).
// Today is always included even if it has no notes yet.
func (s *Store) LoadStream(numDays int) ([]DayNote, error) {
	days, err := s.ListDays()
	if err != nil {
		return nil, err
	}

	limit := numDays
	if limit > len(days) {
		limit = len(days)
	}

	// Take the most recent N days
	recent := days[:limit]

	// Check if today is already in the list
	today := Today()
	todayStr := today.Format("2006-01-02")
	todayIncluded := false
	for _, d := range recent {
		if d.Format("2006-01-02") == todayStr {
			todayIncluded = true
			break
		}
	}

	// Reverse to oldest-first for display
	var result []DayNote
	for i := len(recent) - 1; i >= 0; i-- {
		content, err := s.ReadDay(recent[i])
		if err != nil {
			return nil, err
		}
		if content == "" {
			continue
		}
		result = append(result, DayNote{
			Date:    recent[i],
			Entries: parseEntries(content),
		})
	}

	// Always include today so the user can select and edit it
	if !todayIncluded {
		result = append(result, DayNote{
			Date:    today,
			Entries: nil,
		})
	}

	return result, nil
}

// LoadMoreDays loads older days beyond the given offset.
func (s *Store) LoadMoreDays(offset, count int) ([]DayNote, error) {
	days, err := s.ListDays()
	if err != nil {
		return nil, err
	}

	if offset >= len(days) {
		return nil, nil
	}

	end := offset + count
	if end > len(days) {
		end = len(days)
	}

	slice := days[offset:end]

	var result []DayNote
	for i := len(slice) - 1; i >= 0; i-- {
		content, err := s.ReadDay(slice[i])
		if err != nil {
			return nil, err
		}
		if content == "" {
			continue
		}
		result = append(result, DayNote{
			Date:    slice[i],
			Entries: parseEntries(content),
		})
	}

	return result, nil
}

// parseEntries extracts entries from a day's markdown content.
func parseEntries(content string) []Entry {
	var entries []Entry
	lines := strings.Split(content, "\n")

	var current *Entry
	var contentLines []string

	for _, line := range lines {
		// Check for timestamp header (## 10:32 AM)
		if strings.HasPrefix(line, "## ") {
			// Save previous entry
			if current != nil {
				current.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))
				entries = append(entries, *current)
			}

			timeStr := strings.TrimPrefix(line, "## ")
			t, err := time.Parse("3:04 PM", timeStr)
			if err != nil {
				// Not a time header, skip
				current = nil
				contentLines = nil
				continue
			}

			current = &Entry{Time: t}
			contentLines = nil
			continue
		}

		// Skip day header
		if strings.HasPrefix(line, "# ") {
			continue
		}

		if current != nil {
			// Check if this is a task line
			if strings.HasPrefix(line, "- [ ] ") {
				current.IsTask = true
				current.Done = false
			} else if strings.HasPrefix(line, "- [x] ") || strings.HasPrefix(line, "- [X] ") {
				current.IsTask = true
				current.Done = true
			}
			contentLines = append(contentLines, line)
		}
	}

	// Don't forget last entry
	if current != nil {
		current.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))
		entries = append(entries, *current)
	}

	return entries
}

// GetRawDay returns the raw markdown for a day file (for opening in editor).
func (s *Store) GetRawDay(d time.Time) (string, error) {
	return s.ReadDay(d)
}

// WriteDayRaw overwrites a day file with raw content (for after editor changes).
func (s *Store) WriteDayRaw(d time.Time, content string) error {
	if err := s.EnsureDir(); err != nil {
		return err
	}
	path := s.PathForDate(d)
	return os.WriteFile(path, []byte(content), 0644)
}

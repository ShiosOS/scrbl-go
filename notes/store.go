package notes

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	dayLayout       = "2006-01-02"
	dayHeaderLayout = "2006.01.02"
)

type DayNote struct {
	Date    time.Time
	Content string
}

type Store struct {
	Dir string
}

func NewStore(dir string) *Store {
	return &Store{Dir: dir}
}

func (s *Store) EnsureDir() error {
	return os.MkdirAll(s.Dir, 0o755)
}

func (s *Store) PathForDate(day time.Time) string {
	return filepath.Join(s.Dir, day.Format(dayLayout)+".md")
}

func (s *Store) ReadDay(day time.Time) (string, error) {
	b, err := os.ReadFile(s.PathForDate(day))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(b), nil
}

func (s *Store) WriteDay(day time.Time, content string) error {
	if err := s.EnsureDir(); err != nil {
		return err
	}
	return os.WriteFile(s.PathForDate(day), []byte(content), 0o644)
}

func (s *Store) AppendEntry(day time.Time, content string) error {
	if err := s.EnsureDir(); err != nil {
		return err
	}

	entry := strings.TrimSpace(content)
	if entry == "" {
		return nil
	}

	path := s.PathForDate(day)
	currentBytes, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	current := strings.ReplaceAll(string(currentBytes), "\r\n", "\n")

	if strings.TrimSpace(current) == "" {
		initial := fmt.Sprintf("# %s\n\n%s\n", day.Format(dayHeaderLayout), entry)
		return os.WriteFile(path, []byte(initial), 0o644)
	}

	current = strings.TrimRight(current, "\n")
	updated := current + "\n\n" + entry + "\n"
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return err
	}
	return nil
}

func (s *Store) LoadRecent(limit int) ([]DayNote, bool, error) {
	if err := s.EnsureDir(); err != nil {
		return nil, false, err
	}

	dates, err := s.listDates()
	if err != nil {
		return nil, false, err
	}

	hasMore := false

	if limit > 0 && len(dates) > limit {
		hasMore = true
		dates = dates[len(dates)-limit:]
	}

	notes := make([]DayNote, 0, len(dates))
	for _, day := range dates {
		content, err := s.ReadDay(day)
		if err != nil {
			return nil, false, err
		}
		notes = append(notes, DayNote{Date: day, Content: content})
	}

	if len(notes) == 0 {
		today := Today()
		notes = append(notes, DayNote{
			Date:    today,
			Content: fmt.Sprintf("# %s\n\n", today.Format(dayHeaderLayout)),
		})
		return notes, false, nil
	}

	today := Today()
	todayKey := today.Format(dayLayout)
	hasToday := false
	for _, day := range notes {
		if day.Date.Format(dayLayout) == todayKey {
			hasToday = true
			break
		}
	}
	if !hasToday {
		notes = append(notes, DayNote{
			Date:    today,
			Content: fmt.Sprintf("# %s\n\n", today.Format(dayHeaderLayout)),
		})
	}

	return notes, hasMore, nil
}

func (s *Store) listDates() ([]time.Time, error) {
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		return nil, err
	}

	dates := make([]time.Time, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		day, err := time.Parse(dayLayout, strings.TrimSuffix(name, ".md"))
		if err != nil {
			continue
		}
		dates = append(dates, day)
	}

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	return dates, nil
}

func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

package dayfiles

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	DateLayout      = "2006-01-02"
	DayHeaderLayout = "2006.01.02"
)

func ParseDateOrToday(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	}

	day, err := time.Parse(DateLayout, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q (expected YYYY-MM-DD)", raw)
	}
	return day, nil
}

func Path(notesDir string, day time.Time) string {
	return filepath.Join(notesDir, day.Format(DateLayout)+".md")
}

func Read(notesDir string, day time.Time) (string, error) {
	b, err := os.ReadFile(Path(notesDir, day))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func Write(notesDir string, day time.Time, content string) error {
	if err := os.MkdirAll(notesDir, 0o755); err != nil {
		return fmt.Errorf("create notes dir: %w", err)
	}
	if err := os.WriteFile(Path(notesDir, day), []byte(content), 0o644); err != nil {
		return fmt.Errorf("write day file: %w", err)
	}
	return nil
}

func ListDates(notesDir string) ([]time.Time, error) {
	entries, err := os.ReadDir(notesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read notes dir: %w", err)
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

		day, err := time.Parse(DateLayout, strings.TrimSuffix(name, ".md"))
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

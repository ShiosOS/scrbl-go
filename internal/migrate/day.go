package migrate

import (
	"regexp"
	"strings"
	"time"

	"github.com/juliuswalton/scrbl/internal/dayfiles"
)

var legacyTimestampHeader = regexp.MustCompile(`(?i)^##\s+\d{1,2}:\d{2}\s*(am|pm)\s*$`)

func NormalizeDayContent(day time.Time, content string) (string, bool) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	header := "# " + day.Format(dayfiles.DayHeaderLayout)
	out := []string{header, ""}
	changed := false

	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}

	if i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "# ") {
		if strings.TrimSpace(lines[i]) != header {
			changed = true
		}
		i++
	} else {
		changed = true
	}

	seenBody := false
	for ; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if legacyTimestampHeader.MatchString(trimmed) {
			changed = true
			continue
		}
		if !seenBody && trimmed == "" {
			continue
		}
		if trimmed != "" {
			seenBody = true
		}
		out = append(out, lines[i])
	}

	migrated := strings.Join(out, "\n")
	if !strings.HasSuffix(migrated, "\n") {
		migrated += "\n"
	}

	original := normalized
	if !strings.HasSuffix(original, "\n") {
		original += "\n"
	}

	if migrated != original {
		changed = true
	}

	return migrated, changed
}

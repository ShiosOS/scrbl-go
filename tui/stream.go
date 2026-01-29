package tui

import (
	"strings"
	"time"

	"github.com/juliuswalton/scrbl/notes"
)

// dayLine tracks where each day starts in the rendered content.
type dayLine struct {
	date    time.Time
	lineNum int
}

// buildStreamContent renders all loaded day notes into a single string
// for the viewport, and returns line-number markers for each day boundary.
// focusedIdx indicates which day should be visually highlighted (-1 for none).
func buildStreamContent(days []notes.DayNote, width int, store *notes.Store, focusedIdx int) (string, []dayLine) {
	if len(days) == 0 {
		return helpStyle.Render("\n  No notes yet. Start typing below to create your first entry.\n"), nil
	}

	var sb strings.Builder
	var markers []dayLine
	today := notes.Today()
	lineCount := 0

	for i, day := range days {
		isToday := day.Date.Equal(today)
		isFocused := i == focusedIdx
		dateLabel := day.Date.Format("Monday, January 2, 2006")

		// Record where this day starts
		markers = append(markers, dayLine{date: day.Date, lineNum: lineCount})

		// Day divider
		divider := DayDivider(dateLabel, isToday, isFocused, width)
		sb.WriteString(divider)
		sb.WriteString("\n")
		lineCount += strings.Count(divider, "\n") + 1

		// Read raw file content and strip the # Day Header line,
		// then render through Glamour directly.
		raw, err := store.ReadDay(day.Date)
		if err != nil || raw == "" {
			continue
		}

		stripped := stripDayHeader(raw)
		if stripped == "" {
			continue
		}

		rendered := RenderMarkdown(stripped, width)
		sb.WriteString(rendered)
		sb.WriteString("\n")
		lineCount += strings.Count(rendered, "\n") + 1
	}

	return sb.String(), markers
}

// stripDayHeader removes the first "# ..." line from raw day content.
func stripDayHeader(raw string) string {
	lines := strings.Split(raw, "\n")
	var out []string
	skippedHeader := false

	for _, line := range lines {
		if !skippedHeader && strings.HasPrefix(line, "# ") {
			skippedHeader = true
			continue
		}
		out = append(out, line)
	}

	return strings.TrimSpace(strings.Join(out, "\n"))
}

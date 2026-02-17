package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/juliuswalton/scrbl/notes"
)

func (m *Model) resizeViewport() {
	streamHeight := m.height - m.composeHeight() - 3
	if streamHeight < 3 {
		streamHeight = 3
	}

	if !m.ready {
		m.viewport = viewport.New(m.width, streamHeight)
		m.ready = true
		return
	}

	m.viewport.Width = m.width
	m.viewport.Height = streamHeight
}

func (m *Model) refreshStream(jumpToFocused bool) {
	if !m.ready {
		return
	}
	if len(m.days) == 0 {
		m.streamLines = nil
		m.lineDayIndex = nil
		m.dayStartLine = nil
		m.guideLine = 0
		m.viewport.SetContent("No notes yet.")
		m.focusedDay = -1
		return
	}
	if m.focusedDay < 0 || m.focusedDay >= len(m.days) {
		m.focusedDay = len(m.days) - 1
	}

	oldY := m.viewport.YOffset
	if oldY < 0 {
		oldY = 0
	}
	oldGuide := m.guideLine

	lines, lineDayIndex, dayStartLine := buildStreamData(m.days, m.width)
	m.streamLines = lines
	m.lineDayIndex = lineDayIndex
	m.dayStartLine = dayStartLine

	if len(m.streamLines) == 0 {
		m.viewport.SetContent("No notes yet.")
		m.focusedDay = -1
		return
	}

	maxY := max(0, len(m.streamLines)-m.viewport.Height)
	if jumpToFocused {
		if m.focusedDay >= 0 && m.focusedDay < len(m.dayStartLine) {
			m.guideLine = m.dayStartLine[m.focusedDay]
		}
	} else {
		if oldY > maxY {
			oldY = maxY
		}
		m.viewport.SetYOffset(oldY)

		if oldGuide < 0 {
			oldGuide = 0
		}
		if oldGuide >= len(m.streamLines) {
			oldGuide = len(m.streamLines) - 1
		}
		m.guideLine = oldGuide
	}

	m.ensureGuideVisible()
	if m.guideLine >= 0 && m.guideLine < len(m.lineDayIndex) {
		m.focusedDay = m.lineDayIndex[m.guideLine]
	}
	m.renderStreamOverlay()
}

func (m *Model) jumpToFocusedDay() {
	if len(m.dayStartLine) == 0 || m.focusedDay < 0 || m.focusedDay >= len(m.dayStartLine) {
		return
	}
	m.setGuideLine(m.dayStartLine[m.focusedDay])
}

func (m *Model) moveGuide(delta int) {
	if len(m.streamLines) == 0 {
		return
	}
	m.setGuideLine(m.guideLine + delta)
}

func (m *Model) loadMoreCmd() tea.Cmd {
	if !m.hasMoreDays || m.loadingMore {
		return nil
	}
	if m.guideLine != 0 {
		return nil
	}

	anchorDate := ""
	if m.focusedDay >= 0 && m.focusedDay < len(m.days) {
		anchorDate = m.days[m.focusedDay].Date.Format("2006-01-02")
	}

	m.loadingMore = true
	m.loadLimit += loadMoreDaysStep
	m.status = "loading older notes"

	return m.loadStreamCmd(anchorDate)
}

func (m *Model) setGuideLine(line int) {
	if len(m.streamLines) == 0 {
		m.guideLine = 0
		m.focusedDay = -1
		return
	}
	if line < 0 {
		line = 0
	}
	if line >= len(m.streamLines) {
		line = len(m.streamLines) - 1
	}

	m.guideLine = line
	m.ensureGuideVisible()

	if m.guideLine < len(m.lineDayIndex) {
		m.focusedDay = m.lineDayIndex[m.guideLine]
	}
	if m.focusedDay < 0 && len(m.days) > 0 {
		m.focusedDay = 0
	}

	m.renderStreamOverlay()
}

func (m *Model) ensureGuideVisible() {
	if len(m.streamLines) == 0 || m.viewport.Height <= 0 {
		return
	}

	y := m.viewport.YOffset
	if m.guideLine < y {
		y = m.guideLine
	} else if m.guideLine >= y+m.viewport.Height {
		y = m.guideLine - m.viewport.Height + 1
	}

	maxY := max(0, len(m.streamLines)-m.viewport.Height)
	if y < 0 {
		y = 0
	}
	if y > maxY {
		y = maxY
	}
	m.viewport.SetYOffset(y)
}

func (m *Model) renderStreamOverlay() {
	if len(m.streamLines) == 0 {
		m.viewport.SetContent("No notes yet.")
		return
	}

	overlayWidth := max(1, m.viewport.Width)
	var sb strings.Builder

	for i, line := range m.streamLines {
		prefix := guideRailStyle.Render("│")
		display := prefix + " " + line

		if i == m.guideLine {
			plain := ansi.Strip(line)
			display = guideLineStyle.Width(overlayWidth).Render("▸ " + plain)
		}

		sb.WriteString(display)
		if i < len(m.streamLines)-1 {
			sb.WriteString("\n")
		}
	}

	m.viewport.SetContent(sb.String())
}

func buildStreamData(days []notes.DayNote, width int) ([]string, []int, []int) {
	if len(days) == 0 {
		return nil, nil, nil
	}

	renderWidth := width - 8
	if renderWidth < 24 {
		renderWidth = 24
	}

	lines := make([]string, 0, 512)
	lineDayIndex := make([]int, 0, 512)
	dayStartLine := make([]int, 0, len(days))

	for i, day := range days {
		dayStartLine = append(dayStartLine, len(lines))

		head := dayHeaderStyle.Render(centeredDateBanner(day.Date, renderWidth))
		lines = append(lines, head)
		lineDayIndex = append(lineDayIndex, i)

		body := stripDayHeader(day.Content)
		if strings.TrimSpace(body) == "" {
			body = "_No notes for this day yet._"
		}

		rendered := RenderMarkdown(body, renderWidth)
		for _, ln := range strings.Split(rendered, "\n") {
			lines = append(lines, ln)
			lineDayIndex = append(lineDayIndex, i)
		}

		if i < len(days)-1 {
			lines = append(lines, "")
			lineDayIndex = append(lineDayIndex, i)
		}
	}

	return lines, lineDayIndex, dayStartLine
}

func stripDayHeader(content string) string {
	lines := strings.Split(content, "\n")
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	if start < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[start]), "# ") {
		start++
	}
	return strings.TrimSpace(strings.Join(lines[start:], "\n"))
}

func dayFileTemplate(day time.Time) string {
	return fmt.Sprintf("# %s\n\n", day.Format("2006.01.02"))
}

func centeredDateBanner(day time.Time, width int) string {
	label := day.Format("2006.01.02")
	if width <= len(label)+2 {
		return label
	}

	dashCount := width - len(label) - 2
	left := dashCount / 2
	right := dashCount - left

	if left < 2 {
		left = 2
	}
	if right < 2 {
		right = 2
	}

	return strings.Repeat("-", left) + " " + label + " " + strings.Repeat("-", right)
}

func isSessionClosedError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "session closed")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func pollComposerCmd() tea.Cmd {
	return tea.Tick(60*time.Millisecond, func(time.Time) tea.Msg {
		return composerPollMsg{}
	})
}

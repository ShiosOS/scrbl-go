package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.err != nil {
		return "\n" + errorStyle.Render("error: "+m.err.Error()) + "\n"
	}
	if !m.ready {
		return "\n" + loadingStyle.Render("loading...") + "\n"
	}

	var out strings.Builder
	out.WriteString(m.streamHeader())
	out.WriteString("\n")
	out.WriteString(streamRuleStyle.Width(m.width).Render(strings.Repeat("─", max(1, m.width))))
	out.WriteString("\n")
	out.WriteString(m.viewport.View())
	out.WriteString("\n")
	out.WriteString(m.composePane())
	out.WriteString("\n")
	out.WriteString(m.statusBar())

	return out.String()
}

func (m Model) streamHeader() string {
	title := streamTitleStyle.Render(" _scrbl ")

	meta := ""
	if m.focusedDay >= 0 && m.focusedDay < len(m.days) {
		meta = m.days[m.focusedDay].Date.Format("2006.01.02")
	}
	right := streamMetaStyle.Render(meta)

	gap := m.width - lipgloss.Width(title) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return title + strings.Repeat(" ", gap) + right
}

func (m Model) composePane() string {
	h := m.composeHeight()
	if m.mode != modeCompose {
		line := strings.Repeat("─", max(1, m.width))
		return streamRuleStyle.Width(m.width).Render(line)
	}

	body := m.composer.ViewSnapshot(m.snapshot, m.width-4, h-2)
	return activeBoxStyle.Width(m.width).Height(h).Render(body)
}

func (m Model) statusBar() string {
	modeText := "STREAM"
	modeBadge := modeStreamStyle
	if m.mode == modeCompose {
		modeText = "COMPOSE"
		modeBadge = modeComposeStyle
		if m.snapshot.Mode != "" {
			modeText += "(" + strings.ToUpper(m.snapshot.Mode) + ")"
		}
	}

	focused := ""
	if m.mode == modeStream && m.focusedDay >= 0 && m.focusedDay < len(m.days) {
		focused = " " + m.days[m.focusedDay].Date.Format("2006-01-02")
	}

	help := "[j/k] move [e] edit [i] new [r] reload [q] quit"
	if m.mode == modeCompose {
		help = "[:w] save [:q] back [:wq/:x] save+back [Ctrl+C] quit"
	}

	left := modeBadge.Render(" "+modeText+" ") + subtleStyle.Render(" "+m.status+focused)
	right := subtleStyle.Render(help)
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + right
}

func (m Model) composeHeight() int {
	if m.mode != modeCompose {
		return composePanelMinRows
	}
	if m.height < 24 {
		return composePanelEditRows
	}
	return composePanelMaxRows
}

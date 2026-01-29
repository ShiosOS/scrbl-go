package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// ── Pastel Pink / Blue / White palette ──
	pastelPink   = lipgloss.Color("#F9A8D4") // Pink
	pastelBlue   = lipgloss.Color("#93C5FD") // Blue
	pastelLav    = lipgloss.Color("#C4B5FD") // Lavender
	white        = lipgloss.Color("#F8FAFC") // Bright white
	softWhite    = lipgloss.Color("#E2E8F0") // Soft white
	dimWhite     = lipgloss.Color("#94A3B8") // Dim text
	mutedGray    = lipgloss.Color("#64748B") // Muted
	darkGray     = lipgloss.Color("#475569") // Darker muted
	successGreen = lipgloss.Color("#86EFAC") // Pastel green
	warnAmber    = lipgloss.Color("#FCD34D") // Pastel yellow
	errorRed     = lipgloss.Color("#FCA5A5") // Pastel red

	// Day divider for past days
	dayDividerStyle = lipgloss.NewStyle().
			Foreground(dimWhite).
			Bold(true).
			PaddingTop(1).
			PaddingBottom(0)

	// Day divider for today
	todayDividerStyle = lipgloss.NewStyle().
				Foreground(pastelPink).
				Bold(true).
				PaddingTop(1).
				PaddingBottom(0)

	// Focused day divider
	focusedDayDividerStyle = lipgloss.NewStyle().
				Foreground(white).
				Background(lipgloss.Color("#3B1D4E")).
				Bold(true).
				PaddingTop(1).
				PaddingBottom(0)

	// Timestamp style
	timestampStyle = lipgloss.NewStyle().
			Foreground(pastelBlue).
			Bold(true)

	// Input box border
	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(darkGray).
			Padding(0, 1)

	// Input box border when focused
	inputBoxFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(pastelPink).
				Padding(0, 1)

	// Status bar at the bottom
	statusBarStyle = lipgloss.NewStyle().
			Foreground(dimWhite).
			PaddingTop(0)

	// Mode indicator
	insertModeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0F172A")).
			Background(pastelPink).
			Bold(true).
			Padding(0, 1)

	normalModeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0F172A")).
			Background(pastelBlue).
			Bold(true).
			Padding(0, 1)

	// Search overlay
	searchBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(pastelBlue).
			Padding(0, 1).
			Width(60)

	searchResultStyle = lipgloss.NewStyle().
				Foreground(softWhite)

	searchResultDateStyle = lipgloss.NewStyle().
				Foreground(pastelBlue).
				Bold(true)

	searchResultSelectedStyle = lipgloss.NewStyle().
					Foreground(white).
					Background(lipgloss.Color("#1E293B")).
					Bold(true)

	// Help text
	helpStyle = lipgloss.NewStyle().
			Foreground(darkGray)

	// Sync indicator
	syncingStyle = lipgloss.NewStyle().
			Foreground(warnAmber)

	syncedStyle = lipgloss.NewStyle().
			Foreground(successGreen)

	// Edit mode
	editModeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0F172A")).
			Background(warnAmber).
			Bold(true).
			Padding(0, 1)

	editInsertModeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0F172A")).
				Background(pastelPink).
				Bold(true).
				Padding(0, 1)

	editNormalModeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0F172A")).
				Background(pastelBlue).
				Bold(true).
				Padding(0, 1)

	editCommandModeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0F172A")).
				Background(warnAmber).
				Bold(true).
				Padding(0, 1)

	editHeaderStyle = lipgloss.NewStyle().
			Foreground(white).
			Background(lipgloss.Color("#3B1D4E")).
			Bold(true).
			Padding(0, 1)

	editBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(pastelLav).
			Padding(0, 1)

	// Command line style
	commandLineStyle = lipgloss.NewStyle().
				Foreground(white).
				Bold(true)
)

// DayDivider renders a day divider line.
func DayDivider(date string, isToday bool, focused bool, width int) string {
	style := dayDividerStyle
	if focused {
		style = focusedDayDividerStyle
	} else if isToday {
		style = todayDividerStyle
	}

	prefix := "╌╌╌╌╌ "
	if focused {
		prefix = "▶ "
	}
	suffix := " "

	label := date
	if isToday {
		label = date + " (Today)"
	}

	remaining := width - lipgloss.Width(prefix) - lipgloss.Width(suffix) - lipgloss.Width(label)
	if remaining < 4 {
		remaining = 4
	}
	for i := 0; i < remaining; i++ {
		suffix += "╌"
	}

	return style.Render(prefix + label + suffix)
}

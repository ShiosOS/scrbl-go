package tui

import "github.com/charmbracelet/lipgloss"

const (
	dracBg           = "#2A212C"
	dracFg           = "#F8F8F2"
	dracSelection    = "#544158"
	dracComment      = "#624C67"
	dracCursor       = "#9F70A9"
	dracBlue         = "#9580FF"
	dracBlueBright   = "#AA99FF"
	dracGreen        = "#8AFF80"
	dracPink         = "#FF80BF"
	dracPinkBright   = "#FF99CC"
	dracYellow       = "#FFFF80"
	dracYellowBright = "#FFFF99"
	dracRed          = "#FF9580"
)

var (
	streamTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(dracFg)).
				Background(lipgloss.Color(dracBg)).
				Bold(true).
				Padding(0, 1)

	streamMetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(dracBlueBright)).
			Background(lipgloss.Color(dracBg)).
			Padding(0, 1)

	streamRuleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(dracComment))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(dracComment)).
			Foreground(lipgloss.Color(dracFg)).
			Background(lipgloss.Color(dracBg)).
			Padding(0, 1)

	activeBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(dracBlue)).
			Foreground(lipgloss.Color(dracFg)).
			Background(lipgloss.Color(dracBg)).
			Padding(0, 1)

	composeTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(dracPinkBright)).
				Bold(true)

	dayHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(dracYellowBright)).
			Bold(true)

	guideRailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(dracComment))

	guidePointerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(dracPink)).
				Bold(true)

	guideLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(dracFg)).
			Background(lipgloss.Color(dracSelection))

	modeStreamStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(dracBg)).
			Background(lipgloss.Color(dracGreen)).
			Bold(true)

	modeComposeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(dracBg)).
				Background(lipgloss.Color(dracPink)).
				Bold(true)

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(dracCursor))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(dracRed)).
			Bold(true)

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(dracBlueBright))
)

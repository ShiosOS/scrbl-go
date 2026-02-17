package tui

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

const markdownLineWidth = 80

func RenderMarkdown(content string, width int) string {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(effectiveMarkdownWidth(width)),
	)
	if err != nil {
		return content
	}

	out, err := renderer.Render(content)
	if err != nil {
		return content
	}

	return strings.TrimRight(out, "\n")
}

func effectiveMarkdownWidth(width int) int {
	if width <= 0 {
		return markdownLineWidth
	}
	if width < 24 {
		return 24
	}
	if width > markdownLineWidth {
		return markdownLineWidth
	}
	return width
}

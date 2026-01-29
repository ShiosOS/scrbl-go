package tui

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// ── Markdown rendering styles (pastel pink/blue/white) ──

	// Headers — no ## prefix shown, distinct sizes by color intensity
	h2Style = lipgloss.NewStyle().
		Foreground(pastelPink).
		Bold(true).
		Underline(true)

	h3Style = lipgloss.NewStyle().
		Foreground(pastelBlue).
		Bold(true)

	h4Style = lipgloss.NewStyle().
		Foreground(pastelLav).
		Bold(true)

	h5Style = lipgloss.NewStyle().
		Foreground(dimWhite).
		Bold(true).
		Italic(true)

	// Bullet points
	bulletPointStyle = lipgloss.NewStyle().
				Foreground(pastelPink)

	// Checkboxes
	checkboxUnchecked = lipgloss.NewStyle().
				Foreground(warnAmber).
				Bold(true)

	checkboxChecked = lipgloss.NewStyle().
			Foreground(successGreen).
			Strikethrough(true)

	// Inline code
	codeStyle = lipgloss.NewStyle().
			Foreground(pastelBlue).
			Background(lipgloss.Color("#1E293B"))

	// Code block
	codeBlockStyle = lipgloss.NewStyle().
			Foreground(pastelBlue).
			Background(lipgloss.Color("#1E293B")).
			PaddingLeft(2).
			PaddingRight(2)

	// Bold text
	boldStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(white)

	// Normal text
	mdTextStyle = lipgloss.NewStyle().
			Foreground(softWhite)

	// Blockquote
	quoteStyle = lipgloss.NewStyle().
			Foreground(dimWhite).
			Italic(true).
			PaddingLeft(2).
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(pastelLav)

	// Regex patterns
	boldRegex       = regexp.MustCompile(`\*\*(.+?)\*\*`)
	inlineCodeRegex = regexp.MustCompile("`([^`]+)`")
)

// RenderMarkdown renders markdown content to styled terminal output.
func RenderMarkdown(content string, width int) string {
	lines := strings.Split(content, "\n")
	var result []string
	inCodeBlock := false
	var codeLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Code block handling
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				block := strings.Join(codeLines, "\n")
				rendered := codeBlockStyle.Render(block)
				result = append(result, rendered)
				codeLines = nil
				inCodeBlock = false
			} else {
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			codeLines = append(codeLines, line)
			continue
		}

		result = append(result, renderLine(line))
	}

	if inCodeBlock && len(codeLines) > 0 {
		block := strings.Join(codeLines, "\n")
		result = append(result, codeBlockStyle.Render(block))
	}

	return strings.Join(result, "\n")
}

func renderLine(line string) string {
	trimmed := strings.TrimSpace(line)

	if trimmed == "" {
		return ""
	}

	// ── Headers (strip prefix, apply style) ──
	if strings.HasPrefix(trimmed, "##### ") {
		return h5Style.Render(strings.TrimPrefix(trimmed, "##### "))
	}
	if strings.HasPrefix(trimmed, "#### ") {
		return h4Style.Render(strings.TrimPrefix(trimmed, "#### "))
	}
	if strings.HasPrefix(trimmed, "### ") {
		return h3Style.Render(strings.TrimPrefix(trimmed, "### "))
	}
	if strings.HasPrefix(trimmed, "## ") {
		return h2Style.Render(strings.TrimPrefix(trimmed, "## "))
	}

	// ── Blockquote ──
	if strings.HasPrefix(trimmed, "> ") {
		text := strings.TrimPrefix(trimmed, "> ")
		return quoteStyle.Render(renderInline(text))
	}

	// ── Task checkboxes ──
	if strings.HasPrefix(trimmed, "- [x] ") || strings.HasPrefix(trimmed, "- [X] ") {
		text := trimmed[6:]
		indent := leadingSpaces(line)
		return strings.Repeat(" ", indent) + checkboxChecked.Render("☑ "+text)
	}
	if strings.HasPrefix(trimmed, "- [ ] ") {
		text := trimmed[6:]
		indent := leadingSpaces(line)
		return strings.Repeat(" ", indent) + checkboxUnchecked.Render("☐ "+text)
	}

	// ── Bullet points (* or -) ──
	if strings.HasPrefix(trimmed, "* ") {
		text := strings.TrimPrefix(trimmed, "* ")
		indent := leadingSpaces(line)
		return strings.Repeat(" ", indent) + bulletPointStyle.Render("• ") + renderInline(text)
	}
	if strings.HasPrefix(trimmed, "- ") {
		text := strings.TrimPrefix(trimmed, "- ")
		indent := leadingSpaces(line)
		return strings.Repeat(" ", indent) + bulletPointStyle.Render("• ") + renderInline(text)
	}

	// ── Horizontal rule ──
	if trimmed == "---" || trimmed == "***" || trimmed == "___" {
		return lipgloss.NewStyle().Foreground(darkGray).Render("─────────────────────────────────────")
	}

	// ── Regular text ──
	indent := leadingSpaces(line)
	return strings.Repeat(" ", indent) + renderInline(trimmed)
}

func renderInline(text string) string {
	// Replace **bold**
	text = boldRegex.ReplaceAllStringFunc(text, func(match string) string {
		inner := match[2 : len(match)-2]
		return boldStyle.Render(inner)
	})

	// Replace `code`
	text = inlineCodeRegex.ReplaceAllStringFunc(text, func(match string) string {
		inner := match[1 : len(match)-1]
		return codeStyle.Render(inner)
	})

	// Apply text style if no ANSI codes present
	if !strings.Contains(text, "\033[") {
		return mdTextStyle.Render(text)
	}

	return text
}

func leadingSpaces(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4
		} else {
			break
		}
	}
	return count
}

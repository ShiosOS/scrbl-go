package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/juliuswalton/scrbl/config"
	"github.com/juliuswalton/scrbl/notes"
	"github.com/spf13/cobra"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Copy today's summary to clipboard as a Slack-formatted message",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		store := notes.NewStore(cfg.NotesDir)
		today := notes.Today()

		content, err := store.ReadDay(today)
		if err != nil {
			return fmt.Errorf("failed to read today's notes: %w", err)
		}
		if content == "" {
			fmt.Println("  No notes for today.")
			return nil
		}

		summary := extractSummary(content)
		if summary == "" {
			fmt.Println("  No summary section found in today's notes.")
			fmt.Println("  Add a ### Summary or ### Daily Summary header to your notes.")
			return nil
		}

		slack := toSlackFormat(summary)

		if err := copyToClipboard(slack); err != nil {
			// If clipboard fails, just print it
			fmt.Println("  Could not copy to clipboard:", err)
			fmt.Println()
			fmt.Println(slack)
			return nil
		}

		fmt.Println("  Summary copied to clipboard!")
		fmt.Println()
		fmt.Println(slack)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(summaryCmd)
}

// extractSummary pulls the content under ### Summary or ### Daily Summary.
// It reads until the next header of equal or higher level, or end of file.
func extractSummary(content string) string {
	lines := strings.Split(content, "\n")
	var summaryLines []string
	inSummary := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for summary header start
		if !inSummary {
			lower := strings.ToLower(trimmed)
			if lower == "### summary" || lower == "### daily summary" {
				inSummary = true
				continue
			}
			continue
		}

		// Stop if we hit another header of equal or higher level
		if strings.HasPrefix(trimmed, "# ") ||
			strings.HasPrefix(trimmed, "## ") ||
			strings.HasPrefix(trimmed, "### ") {
			break
		}

		summaryLines = append(summaryLines, line)
	}

	return strings.TrimSpace(strings.Join(summaryLines, "\n"))
}

// toSlackFormat converts markdown-style summary to Slack mrkdwn format.
func toSlackFormat(summary string) string {
	lines := strings.Split(summary, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result = append(result, "")
			continue
		}

		indent := countIndent(line)
		slackIndent := strings.Repeat("    ", indent)

		// Convert checkboxes (before generic bullets)
		if strings.HasPrefix(trimmed, "- [x] ") || strings.HasPrefix(trimmed, "- [X] ") {
			text := trimmed[6:]
			text = convertInlineFormatting(text)
			result = append(result, slackIndent+"- [x] ~"+text+"~")
			continue
		}
		if strings.HasPrefix(trimmed, "- [ ] ") {
			text := trimmed[6:]
			text = convertInlineFormatting(text)
			result = append(result, slackIndent+"- [ ] "+text)
			continue
		}

		// Convert bullet points
		if strings.HasPrefix(trimmed, "* ") {
			text := strings.TrimPrefix(trimmed, "* ")
			text = convertInlineFormatting(text)
			result = append(result, slackIndent+"- "+text)
			continue
		}
		if strings.HasPrefix(trimmed, "- ") {
			text := strings.TrimPrefix(trimmed, "- ")
			text = convertInlineFormatting(text)
			result = append(result, slackIndent+"- "+text)
			continue
		}

		// Regular text (like "Today:")
		text := convertInlineFormatting(trimmed)
		result = append(result, slackIndent+text)
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
}

// countIndent returns the nesting level based on leading spaces.
func countIndent(line string) int {
	spaces := 0
	for _, ch := range line {
		if ch == ' ' {
			spaces++
		} else if ch == '\t' {
			spaces += 4
		} else {
			break
		}
	}
	return spaces / 2 // 2-space indent per level
}

// convertInlineFormatting converts markdown inline to Slack format.
func convertInlineFormatting(text string) string {
	// **bold** → *bold* (Slack uses single asterisks)
	text = strings.ReplaceAll(text, "**", "*")

	// `code` stays the same in Slack

	return text
}

// copyToClipboard copies text to the system clipboard.
func copyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("clip")
	case "darwin":
		cmd = exec.Command("pbcopy")
	default:
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard tool found (install xclip or xsel)")
		}
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

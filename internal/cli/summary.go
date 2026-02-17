package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/juliuswalton/scrbl/internal/config"
	"github.com/juliuswalton/scrbl/internal/dayfiles"
)

var (
	summaryHeadingRegex  = regexp.MustCompile(`(?i)^##\s+(summary|daily\s+summary)\s*:?[ \t#]*$`)
	headingBoundaryRegex = regexp.MustCompile(`^#{1,2}\s+`)
	markdownHeadingRegex = regexp.MustCompile(`^\s*#{1,6}\s+(.+?)\s*$`)
	taskDoneRegex        = regexp.MustCompile(`^(\s*)-\s+\[(x|X)\]\s+(.*)$`)
	taskTodoRegex        = regexp.MustCompile(`^(\s*)-\s+\[\s\]\s+(.*)$`)
	unorderedListRegex   = regexp.MustCompile(`^(\s*)[-*+]\s+(.*)$`)
	markdownBoldRegex    = regexp.MustCompile(`\*\*(.+?)\*\*`)
)

func runSummary(args []string) error {
	if len(args) > 0 && args[0] == "slack" {
		args = args[1:]
	}

	fs := flag.NewFlagSet("summary", flag.ContinueOnError)
	dateRaw := fs.String("date", "", "date to export (YYYY-MM-DD), defaults to most recent day")
	stdout := fs.Bool("stdout", false, "also print generated Slack markdown")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("summary does not take positional arguments")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	day, err := resolveSummaryDate(cfg.NotesDir, strings.TrimSpace(*dateRaw))
	if err != nil {
		return err
	}

	content, err := dayfiles.Read(cfg.NotesDir, day)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("local note not found for %s", day.Format(dayfiles.DateLayout))
		}
		return err
	}

	summaryMarkdown, err := extractSummarySection(content)
	if err != nil {
		return fmt.Errorf("%s: %w", day.Format(dayfiles.DateLayout), err)
	}

	slackMarkdown := formatForSlack(summaryMarkdown)
	if err := copyToClipboard(slackMarkdown); err != nil {
		return err
	}

	fmt.Printf("copied summary for %s to clipboard\n", day.Format(dayfiles.DateLayout))
	if *stdout {
		fmt.Println()
		fmt.Println(slackMarkdown)
	}

	return nil
}

func resolveSummaryDate(notesDir string, raw string) (time.Time, error) {
	if raw != "" {
		day, parseErr := dayfiles.ParseDateOrToday(raw)
		if parseErr != nil {
			return time.Time{}, parseErr
		}
		return day, nil
	}

	dates, err := dayfiles.ListDates(notesDir)
	if err != nil {
		return time.Time{}, err
	}
	if len(dates) == 0 {
		return time.Time{}, fmt.Errorf("no local notes found")
	}

	return dates[len(dates)-1], nil
}

func extractSummarySection(content string) (string, error) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	found := false
	inCode := false
	collected := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !found {
			if summaryHeadingRegex.MatchString(trimmed) {
				found = true
			}
			continue
		}

		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			collected = append(collected, line)
			continue
		}

		if !inCode && headingBoundaryRegex.MatchString(trimmed) && !summaryHeadingRegex.MatchString(trimmed) {
			break
		}

		collected = append(collected, line)
	}

	if !found {
		return "", fmt.Errorf("## Summary section not found")
	}

	section := strings.TrimSpace(strings.Join(collected, "\n"))
	if section == "" {
		return "", fmt.Errorf("## Summary section is empty")
	}

	return section, nil
}

func formatForSlack(markdown string) string {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	inCode := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			continue
		}

		if m := markdownHeadingRegex.FindStringSubmatch(line); len(m) == 2 {
			lines[i] = "*" + strings.TrimSpace(m[1]) + "*"
			continue
		}

		if m := taskDoneRegex.FindStringSubmatch(line); len(m) == 4 {
			lines[i] = m[1] + "- [x] " + strings.TrimSpace(m[3])
			continue
		}

		if m := taskTodoRegex.FindStringSubmatch(line); len(m) == 3 {
			lines[i] = m[1] + "- [ ] " + strings.TrimSpace(m[2])
			continue
		}

		if m := unorderedListRegex.FindStringSubmatch(line); len(m) == 3 {
			lines[i] = m[1] + "• " + strings.TrimSpace(m[2])
			continue
		}

		lines[i] = markdownBoldRegex.ReplaceAllString(line, "*$1*")
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func copyToClipboard(text string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("nothing to copy")
	}

	switch runtime.GOOS {
	case "windows":
		powershellCmd := exec.Command(
			"powershell",
			"-NoProfile",
			"-NonInteractive",
			"-Command",
			"[Console]::InputEncoding=[System.Text.Encoding]::UTF8; Set-Clipboard -Value ([Console]::In.ReadToEnd())",
		)
		if err := runClipboardCommand(powershellCmd, text); err == nil {
			return nil
		}
		return runClipboardCommand(exec.Command("cmd", "/c", "clip"), text)
	case "darwin":
		return runClipboardCommand(exec.Command("pbcopy"), text)
	default:
		candidates := []struct {
			name string
			args []string
		}{
			{name: "wl-copy"},
			{name: "xclip", args: []string{"-selection", "clipboard"}},
			{name: "xsel", args: []string{"--clipboard", "--input"}},
		}

		var firstErr error
		for _, c := range candidates {
			if _, err := exec.LookPath(c.name); err != nil {
				continue
			}
			if err := runClipboardCommand(exec.Command(c.name, c.args...), text); err == nil {
				return nil
			} else if firstErr == nil {
				firstErr = err
			}
		}

		if firstErr != nil {
			return firstErr
		}
		return errors.New("no clipboard utility found (install wl-copy, xclip, or xsel)")
	}
}

func runClipboardCommand(cmd *exec.Cmd, text string) error {
	cmd.Stdin = strings.NewReader(text)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("copy to clipboard failed: %v (%s)", err, strings.TrimSpace(stderr.String()))
		}
		return fmt.Errorf("copy to clipboard failed: %w", err)
	}

	return nil
}

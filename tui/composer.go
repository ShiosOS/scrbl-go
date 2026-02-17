package tui

import (
	"fmt"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neovim/go-client/nvim"
)

type ComposerSnapshot struct {
	Content string
	Mode    string
	Cursor  [2]int
}

type Composer struct {
	editor string
	nv     *nvim.Nvim

	mu                   sync.Mutex
	requestSave          bool
	requestQuit          bool
	requestQuitAfterSave bool
}

func NewComposer(editor string) *Composer {
	if strings.TrimSpace(editor) == "" {
		editor = "nvim"
	}
	return &Composer{editor: editor}
}

func (c *Composer) Start() error {
	if c.nv != nil {
		if _, err := c.nv.Mode(); err == nil {
			return nil
		}
		c.Close()
	}

	nv, err := nvim.NewChildProcess(
		nvim.ChildProcessCommand(c.editor),
		nvim.ChildProcessArgs("--embed", "--headless", "-u", "NONE", "-n", "-i", "NONE"),
	)
	if err != nil {
		return fmt.Errorf("start embedded neovim: %w", err)
	}

	c.nv = nv
	c.clearRequests()

	if err := c.nv.RegisterHandler("scrbl_write", func() {
		c.markSave(false)
	}); err != nil {
		c.Close()
		return fmt.Errorf("register scrbl_write handler: %w", err)
	}
	if err := c.nv.RegisterHandler("scrbl_write_quit", func() {
		c.markSave(true)
	}); err != nil {
		c.Close()
		return fmt.Errorf("register scrbl_write_quit handler: %w", err)
	}
	if err := c.nv.RegisterHandler("scrbl_quit", func() {
		c.markQuit()
	}); err != nil {
		c.Close()
		return fmt.Errorf("register scrbl_quit handler: %w", err)
	}

	channelID := c.nv.ChannelID()
	setup := []string{
		"enew",
		"setlocal buftype=nofile bufhidden=wipe noswapfile",
		"setlocal filetype=markdown",
		"setlocal nowrap",
		"set noshowmode noruler noshowcmd",
		fmt.Sprintf("command! ScrblWrite call rpcnotify(%d, 'scrbl_write')", channelID),
		fmt.Sprintf("command! ScrblWriteQuit call rpcnotify(%d, 'scrbl_write_quit')", channelID),
		fmt.Sprintf("command! ScrblQuit call rpcnotify(%d, 'scrbl_quit')", channelID),
		"cnoreabbrev <expr> w (getcmdtype() == ':' && getcmdline() ==# 'w') ? 'ScrblWrite' : 'w'",
		"cnoreabbrev <expr> wq (getcmdtype() == ':' && getcmdline() ==# 'wq') ? 'ScrblWriteQuit' : 'wq'",
		"cnoreabbrev <expr> x (getcmdtype() == ':' && getcmdline() ==# 'x') ? 'ScrblWriteQuit' : 'x'",
		"cnoreabbrev <expr> q (getcmdtype() == ':' && getcmdline() ==# 'q') ? 'ScrblQuit' : 'q'",
		"cnoreabbrev <expr> q! (getcmdtype() == ':' && getcmdline() ==# 'q!') ? 'ScrblQuit' : 'q!'",
		"cnoreabbrev <expr> qa (getcmdtype() == ':' && getcmdline() ==# 'qa') ? 'ScrblQuit' : 'qa'",
		"cnoreabbrev <expr> qa! (getcmdtype() == ':' && getcmdline() ==# 'qa!') ? 'ScrblQuit' : 'qa!'",
		"cnoreabbrev <expr> quit (getcmdtype() == ':' && getcmdline() ==# 'quit') ? 'ScrblQuit' : 'quit'",
		"cnoreabbrev <expr> quit! (getcmdtype() == ':' && getcmdline() ==# 'quit!') ? 'ScrblQuit' : 'quit!'",
		"startinsert",
	}
	for _, cmd := range setup {
		if err := c.nv.Command(cmd); err != nil {
			c.Close()
			return fmt.Errorf("configure neovim: %w", err)
		}
	}

	return nil
}

func (c *Composer) Close() {
	if c.nv == nil {
		return
	}
	_ = c.nv.Close()
	c.nv = nil
	c.clearRequests()
}

func (c *Composer) Input(msg tea.KeyMsg) error {
	if c.nv == nil {
		return fmt.Errorf("composer is not started")
	}

	keys := keyMsgToNvim(msg)
	if keys == "" {
		return nil
	}

	_, err := c.nv.Input(keys)
	return err
}

func (c *Composer) Snapshot() ComposerSnapshot {
	if c.nv == nil {
		return ComposerSnapshot{}
	}

	buf, err := c.nv.CurrentBuffer()
	if err != nil {
		return ComposerSnapshot{}
	}

	rawLines, err := c.nv.BufferLines(buf, 0, -1, true)
	if err != nil {
		return ComposerSnapshot{}
	}

	lines := make([]string, 0, len(rawLines))
	for _, raw := range rawLines {
		lines = append(lines, string(raw))
	}

	mode := ""
	if m, err := c.nv.Mode(); err == nil {
		mode = m.Mode
	}

	cursor := [2]int{1, 0}
	if win, err := c.nv.CurrentWindow(); err == nil {
		if pos, err := c.nv.WindowCursor(win); err == nil {
			cursor = pos
		}
	}

	return ComposerSnapshot{
		Content: strings.Join(lines, "\n"),
		Mode:    mode,
		Cursor:  cursor,
	}
}

func (c *Composer) Clear() error {
	return c.SetContent("", true)
}

func (c *Composer) SetContent(content string, insertMode bool) error {
	if c.nv == nil {
		return fmt.Errorf("composer is not started")
	}

	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	buf, err := c.nv.CurrentBuffer()
	if err != nil {
		return err
	}

	replacement := make([][]byte, 0, len(lines))
	for _, line := range lines {
		replacement = append(replacement, []byte(line))
	}

	if err := c.nv.SetBufferLines(buf, 0, -1, true, replacement); err != nil {
		return err
	}

	if win, err := c.nv.CurrentWindow(); err == nil {
		_ = c.nv.SetWindowCursor(win, [2]int{1, 0})
	}

	if insertMode {
		return c.nv.Command("startinsert")
	}
	return c.nv.Command("stopinsert")
}

func (c *Composer) ConsumeRequests() (save bool, quit bool, quitAfterSave bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	save = c.requestSave
	quit = c.requestQuit
	quitAfterSave = c.requestQuitAfterSave
	c.requestSave = false
	c.requestQuit = false
	c.requestQuitAfterSave = false
	return save, quit, quitAfterSave
}

func (c *Composer) markSave(quit bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.requestSave = true
	if quit {
		c.requestQuitAfterSave = true
	}
}

func (c *Composer) markQuit() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.requestQuit = true
}

func (c *Composer) clearRequests() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.requestSave = false
	c.requestQuit = false
	c.requestQuitAfterSave = false
}

func (c *Composer) ViewSnapshot(s ComposerSnapshot, width int, height int) string {
	lines := strings.Split(s.Content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	row := s.Cursor[0] - 1
	if row < 0 {
		row = 0
	}
	if row >= len(lines) {
		row = len(lines) - 1
	}
	col := s.Cursor[1]
	if col < 0 {
		col = 0
	}

	start := 0
	if len(lines) > height {
		start = row - (height / 2)
		if start < 0 {
			start = 0
		}
		if start > len(lines)-height {
			start = len(lines) - height
		}
		lines = lines[start : start+height]
	}

	visibleRow := row - start
	if visibleRow >= 0 && visibleRow < len(lines) {
		lines[visibleRow] = renderCursorLine(lines[visibleRow], col)
	}

	if len(lines) < height {
		padding := make([]string, height-len(lines))
		lines = append(lines, padding...)
	}

	body := strings.Join(lines, "\n")
	return lipgloss.NewStyle().Width(width).Height(height).Render(body)
}

func renderCursorLine(line string, col int) string {
	r := []rune(line)
	if len(r) == 0 {
		return lipgloss.NewStyle().Reverse(true).Render(" ")
	}
	if col < 0 {
		col = 0
	}
	if col >= len(r) {
		return string(r) + lipgloss.NewStyle().Reverse(true).Render(" ")
	}

	left := string(r[:col])
	cur := lipgloss.NewStyle().Reverse(true).Render(string(r[col]))
	right := string(r[col+1:])
	return left + cur + right
}

func keyMsgToNvim(msg tea.KeyMsg) string {
	s := msg.String()

	switch s {
	case "enter":
		return "\r"
	case "tab":
		return "\t"
	case "backspace":
		return "\b"
	case "esc":
		return "\x1b"
	case "space", " ":
		return " "
	case "up":
		return "<Up>"
	case "down":
		return "<Down>"
	case "left":
		return "<Left>"
	case "right":
		return "<Right>"
	case "home":
		return "<Home>"
	case "end":
		return "<End>"
	case "pgup":
		return "<PageUp>"
	case "pgdown":
		return "<PageDown>"
	case "delete":
		return "<Del>"
	case "insert":
		return "<Insert>"
	}

	if strings.HasPrefix(s, "ctrl+") {
		k := strings.TrimPrefix(s, "ctrl+")
		if len(k) == 1 {
			return "<C-" + k + ">"
		}
	}

	if strings.HasPrefix(s, "alt+") {
		k := strings.TrimPrefix(s, "alt+")
		if len(k) == 1 {
			return "<M-" + k + ">"
		}
	}

	if len(msg.Runes) > 0 {
		return string(msg.Runes)
	}

	return ""
}

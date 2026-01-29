package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/juliuswalton/scrbl/notes"
	"github.com/juliuswalton/scrbl/sync"
)

// Mode represents the current input mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
)

// AppModel is the root Bubble Tea model.
type AppModel struct {
	store      *notes.Store
	syncer     *sync.Client
	editor     string
	viewport   viewport.Model
	input      textarea.Model
	search     SearchModel
	mode       Mode
	width      int
	height     int
	days       []notes.DayNote
	dayMarkers []dayLine
	daysLoaded int
	ready      bool
	syncStatus string
	err        error
	focusedDay int
}

// syncDoneMsg signals sync completion.
type syncDoneMsg struct{ err error }

// streamLoadedMsg signals stream data is ready.
type streamLoadedMsg struct {
	days []notes.DayNote
	err  error
}

// moreDaysMsg signals older days were loaded.
type moreDaysMsg struct {
	days []notes.DayNote
	err  error
}

// editorFinishedMsg signals the external editor closed.
type editorFinishedMsg struct{ err error }

// searchLoadAndNavigateMsg signals older days were loaded and we need to navigate to a target date.
type searchLoadAndNavigateMsg struct {
	days   []notes.DayNote
	target time.Time
}

// NewApp creates the root TUI model.
func NewApp(store *notes.Store, syncer *sync.Client, editor string) AppModel {
	ta := textarea.New()
	ta.Placeholder = "Write something..."
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.Focus()

	search := NewSearchModel(store)

	if editor == "" {
		editor = "nvim"
	}

	return AppModel{
		store:      store,
		syncer:     syncer,
		editor:     editor,
		input:      ta,
		search:     search,
		mode:       ModeInsert,
		daysLoaded: 7,
		syncStatus: "",
	}
}

// Init loads the initial stream data.
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.loadStream(),
	)
}

func (m AppModel) loadStream() tea.Cmd {
	return func() tea.Msg {
		days, err := m.store.LoadStream(m.daysLoaded)
		return streamLoadedMsg{days: days, err: err}
	}
}

func (m AppModel) loadMoreDays() tea.Cmd {
	offset := m.daysLoaded
	return func() tea.Msg {
		days, err := m.store.LoadMoreDays(offset, 7)
		return moreDaysMsg{days: days, err: err}
	}
}

func (m AppModel) syncNote() tea.Cmd {
	if m.syncer == nil {
		return nil
	}
	today := notes.Today()
	return func() tea.Msg {
		content, err := m.store.ReadDay(today)
		if err != nil {
			return syncDoneMsg{err: err}
		}
		err = m.syncer.PushNote(today, content)
		return syncDoneMsg{err: err}
	}
}

func (m AppModel) openEditor(date interface{ Format(string) string }) tea.Cmd {
	path := m.store.PathForDate(notes.Today())
	// Use the actual date from the focused day marker
	if d, ok := date.(interface{ Format(string) string }); ok {
		_ = d
	}
	c := exec.Command(m.editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
}

func (m AppModel) openEditorForDay() tea.Cmd {
	if len(m.dayMarkers) == 0 || m.focusedDay < 0 || m.focusedDay >= len(m.dayMarkers) {
		return nil
	}
	date := m.dayMarkers[m.focusedDay].date
	path := m.store.PathForDate(date)
	c := exec.Command(m.editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
}

// inputAreaHeight returns the fixed height used by input box + status bar.
func inputAreaHeight() int {
	return 7
}

// Update handles all messages.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		streamHeight := m.height - inputAreaHeight()
		if streamHeight < 3 {
			streamHeight = 3
		}

		if !m.ready {
			m.viewport = viewport.New(m.width, streamHeight)
			m.viewport.Style = lipgloss.NewStyle()
			m.ready = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = streamHeight
		}

		m.input.SetWidth(m.width - 4)
		m.search.SetSize(m.width, m.height)

		m.refreshViewport()
		return m, nil

	case streamLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.days = msg.days
		m.refreshViewport()
		m.viewport.GotoBottom()
		if len(m.dayMarkers) > 0 {
			m.focusedDay = len(m.dayMarkers) - 1
		}
		return m, nil

	case moreDaysMsg:
		if msg.err != nil || len(msg.days) == 0 {
			return m, nil
		}
		m.days = append(msg.days, m.days...)
		m.daysLoaded += 7
		m.refreshViewport()
		return m, nil

	case syncDoneMsg:
		if msg.err != nil {
			m.syncStatus = "sync failed"
		} else {
			m.syncStatus = "synced"
		}
		return m, nil

	case editorFinishedMsg:
		// Reload stream after editor closes
		if msg.err != nil {
			m.err = msg.err
		}
		m.syncStatus = "syncing..."
		return m, tea.Batch(m.loadStream(), m.syncNote())

	case searchNavigateMsg:
		return m.handleSearchNavigate(msg.date)

	case searchLoadAndNavigateMsg:
		if len(msg.days) > 0 {
			m.days = append(msg.days, m.days...)
			m.daysLoaded += len(msg.days)
			m.refreshViewport()
			// Now try to navigate to the target
			targetDate := msg.target.Format("2006-01-02")
			for i, marker := range m.dayMarkers {
				if marker.date.Format("2006-01-02") == targetDate {
					m.focusedDay = i
					m.refreshViewport()
					m.viewport.SetYOffset(m.dayMarkers[m.focusedDay].lineNum)
					return m, nil
				}
			}
		}
		return m, nil

	case tea.KeyMsg:
		if m.search.IsActive() {
			cmd := m.search.Update(msg)
			return m, cmd
		}

		switch m.mode {
		case ModeNormal:
			return m.handleNormalMode(msg)
		case ModeInsert:
			return m.handleInsertMode(msg)
		}
	}

	return m, nil
}

func (m AppModel) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "i", "a":
		m.mode = ModeInsert
		m.input.Focus()
		return m, textarea.Blink
	case "enter":
		return m.submitEntry()
	case "/":
		m.search.Open()
		return m, nil

	// Edit focused day in nvim
	case "e":
		return m, m.openEditorForDay()

	// Navigate between days
	case "[":
		if m.focusedDay > 0 {
			m.focusedDay--
			m.refreshViewport()
			m.viewport.SetYOffset(m.dayMarkers[m.focusedDay].lineNum)
		} else {
			return m, m.loadMoreDays()
		}
		return m, nil
	case "]":
		if m.focusedDay < len(m.dayMarkers)-1 {
			m.focusedDay++
			m.refreshViewport()
			m.viewport.SetYOffset(m.dayMarkers[m.focusedDay].lineNum)
		}
		return m, nil

	// Scroll
	case "k", "up":
		m.viewport.LineUp(1)
		m.updateFocusedDayFromScroll()
		m.refreshViewport()
		return m, nil
	case "j", "down":
		m.viewport.LineDown(1)
		m.updateFocusedDayFromScroll()
		m.refreshViewport()
		return m, nil
	case "ctrl+u":
		m.viewport.HalfViewUp()
		m.updateFocusedDayFromScroll()
		m.refreshViewport()
		return m, nil
	case "ctrl+d":
		m.viewport.HalfViewDown()
		m.updateFocusedDayFromScroll()
		m.refreshViewport()
		return m, nil
	case "g":
		m.viewport.GotoTop()
		m.updateFocusedDayFromScroll()
		m.refreshViewport()
		return m, m.loadMoreDays()
	case "G":
		m.viewport.GotoBottom()
		m.updateFocusedDayFromScroll()
		m.refreshViewport()
		return m, nil
	}
	return m, nil
}

func (m AppModel) handleInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.input.Blur()
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m AppModel) handleSearchNavigate(target time.Time) (tea.Model, tea.Cmd) {
	targetDate := target.Format("2006-01-02")

	// Check if the day is already loaded
	for i, marker := range m.dayMarkers {
		if marker.date.Format("2006-01-02") == targetDate {
			m.focusedDay = i
			m.refreshViewport()
			m.viewport.SetYOffset(m.dayMarkers[m.focusedDay].lineNum)
			return m, nil
		}
	}

	// Day not loaded yet — load more days until we have it
	// Estimate how many more days we need: diff between oldest loaded day and target
	if len(m.days) > 0 {
		oldest := m.days[0].Date
		diff := int(oldest.Sub(target).Hours()/24) + 7 // add buffer
		if diff < 7 {
			diff = 7
		}
		offset := m.daysLoaded
		return m, func() tea.Msg {
			days, err := m.store.LoadMoreDays(offset, diff)
			if err != nil || len(days) == 0 {
				return moreDaysMsg{days: days, err: err}
			}
			// Return a special message that includes the target date to navigate after loading
			return searchLoadAndNavigateMsg{days: days, target: target}
		}
	}

	return m, nil
}

func (m *AppModel) updateFocusedDayFromScroll() {
	if len(m.dayMarkers) == 0 {
		return
	}
	yOffset := m.viewport.YOffset
	m.focusedDay = 0
	for i, marker := range m.dayMarkers {
		if marker.lineNum <= yOffset+m.viewport.Height/2 {
			m.focusedDay = i
		} else {
			break
		}
	}
}

func (m AppModel) submitEntry() (tea.Model, tea.Cmd) {
	content := strings.TrimSpace(m.input.Value())
	if content == "" {
		return m, nil
	}

	isTask := false
	if strings.HasPrefix(content, "todo ") || strings.HasPrefix(content, "TODO ") {
		isTask = true
		content = strings.TrimPrefix(content, "todo ")
		content = strings.TrimPrefix(content, "TODO ")
	}

	today := notes.Today()
	if err := m.store.AppendEntry(today, content, isTask); err != nil {
		m.err = err
		return m, nil
	}

	m.input.Reset()
	m.mode = ModeInsert
	m.input.Focus()

	m.syncStatus = "syncing..."
	return m, tea.Batch(
		m.loadStream(),
		m.syncNote(),
		textarea.Blink,
	)
}

func (m *AppModel) refreshViewport() {
	if !m.ready {
		return
	}
	content, markers := buildStreamContent(m.days, m.width, m.store, m.focusedDay)
	m.dayMarkers = markers
	m.viewport.SetContent(content)
}

// --- View ---

func (m AppModel) View() string {
	if !m.ready {
		return "\n  Loading scrbl...\n"
	}

	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n", m.err)
	}

	if m.search.IsActive() {
		return m.search.View()
	}

	var sb strings.Builder

	sb.WriteString(m.viewport.View())
	sb.WriteString("\n")

	inputStyle := inputBoxStyle
	if m.mode == ModeInsert {
		inputStyle = inputBoxFocusedStyle
	}
	sb.WriteString(inputStyle.Render(m.input.View()))
	sb.WriteString("\n")

	sb.WriteString(m.statusBar())

	return sb.String()
}

func (m AppModel) statusBar() string {
	var modeStr string
	if m.mode == ModeInsert {
		modeStr = insertModeStyle.Render(" INSERT ")
	} else {
		modeStr = normalModeStyle.Render(" NORMAL ")
	}

	var syncStr string
	switch m.syncStatus {
	case "syncing...":
		syncStr = syncingStyle.Render(" syncing...")
	case "synced":
		syncStr = syncedStyle.Render(" synced")
	case "sync failed":
		syncStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render(" sync failed")
	}

	var dayStr string
	if m.mode == ModeNormal && len(m.dayMarkers) > 0 && m.focusedDay >= 0 && m.focusedDay < len(m.dayMarkers) {
		focused := m.dayMarkers[m.focusedDay].date
		if focused.Equal(notes.Today()) {
			dayStr = helpStyle.Render(" Today")
		} else {
			dayStr = helpStyle.Render(" " + focused.Format("Jan 2"))
		}
	}

	var help string
	if m.mode == ModeNormal {
		help = helpStyle.Render("[i]nsert [e]dit [/]search [j/k]scroll [[]prev []]next [q]uit")
	} else {
		help = helpStyle.Render("[Esc] normal  [Ctrl+C] quit  'todo ...' for tasks")
	}

	left := modeStr + syncStr + dayStr
	right := help

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + right
}

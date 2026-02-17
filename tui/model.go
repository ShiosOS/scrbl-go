package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/juliuswalton/scrbl/notes"
	"github.com/juliuswalton/scrbl/sync"
)

type mode int

const (
	modeStream mode = iota
	modeCompose
)

type composeKind int

const (
	composeNew composeKind = iota
	composeEdit
)

const (
	initialLoadDays      = 60
	loadMoreDaysStep     = 60
	composePanelMinRows  = 1
	composePanelMaxRows  = 12
	composePanelEditRows = 8
)

type streamLoadedMsg struct {
	days       []notes.DayNote
	hasMore    bool
	anchorDate string
	err        error
}

type syncResultMsg struct {
	err error
}

type composerPollMsg struct{}

type Model struct {
	store    *notes.Store
	syncer   *sync.Client
	composer *Composer

	mode        mode
	composeKind composeKind
	composeDay  time.Time

	viewport   viewport.Model
	ready      bool
	width      int
	height     int
	days       []notes.DayNote
	focusedDay int

	streamLines  []string
	lineDayIndex []int
	dayStartLine []int
	guideLine    int

	loadLimit   int
	hasMoreDays bool
	loadingMore bool

	snapshot ComposerSnapshot
	status   string
	err      error
}

func NewApp(store *notes.Store, syncer *sync.Client, editor string) Model {
	return Model{
		store:      store,
		syncer:     syncer,
		composer:   NewComposer(editor),
		mode:       modeStream,
		status:     "ready",
		focusedDay: -1,
		loadLimit:  initialLoadDays,
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadStreamCmd("")
}

func (m Model) loadStreamCmd(anchorDate string) tea.Cmd {
	limit := m.loadLimit

	return func() tea.Msg {
		days, hasMore, err := m.store.LoadRecent(limit)
		return streamLoadedMsg{days: days, hasMore: hasMore, anchorDate: anchorDate, err: err}
	}
}

func (m Model) pushDayCmd(day time.Time) tea.Cmd {
	if m.syncer == nil {
		return nil
	}

	return func() tea.Msg {
		content, err := m.store.ReadDay(day)
		if err != nil {
			return syncResultMsg{err: err}
		}

		err = m.syncer.PushNote(day, content)
		return syncResultMsg{err: err}
	}
}

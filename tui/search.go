package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/juliuswalton/scrbl/notes"
)

// searchNavigateMsg tells the app to navigate to a specific day.
type searchNavigateMsg struct {
	date time.Time
}

// SearchModel is the fuzzy search overlay.
type SearchModel struct {
	input   textinput.Model
	store   *notes.Store
	results []notes.SearchResult
	cursor  int
	width   int
	height  int
	active  bool
}

// NewSearchModel creates a new search overlay.
func NewSearchModel(store *notes.Store) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search notes..."
	ti.CharLimit = 256
	ti.Width = 50

	return SearchModel{
		input: ti,
		store: store,
	}
}

// SetSize updates dimensions.
func (m *SearchModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.input.Width = w - 10
	if m.input.Width > 56 {
		m.input.Width = 56
	}
}

// Open activates the search overlay.
func (m *SearchModel) Open() {
	m.active = true
	m.input.SetValue("")
	m.input.Focus()
	m.results = nil
	m.cursor = 0
}

// Close deactivates the search overlay.
func (m *SearchModel) Close() {
	m.active = false
	m.input.Blur()
	m.results = nil
}

// IsActive returns whether the search overlay is open.
func (m *SearchModel) IsActive() bool {
	return m.active
}

// Update handles input for the search overlay.
func (m *SearchModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Close()
			return nil
		case "enter":
			if len(m.results) > 0 && m.cursor < len(m.results) {
				date := m.results[m.cursor].Date
				m.Close()
				return func() tea.Msg {
					return searchNavigateMsg{date: date}
				}
			}
			m.Close()
			return nil
		case "up", "ctrl+k":
			if m.cursor > 0 {
				m.cursor--
			}
			return nil
		case "down", "ctrl+j":
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
			return nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	// Perform search on every keystroke
	query := m.input.Value()
	if query != "" {
		results, err := m.store.Search(query)
		if err == nil {
			m.results = results
			if m.cursor >= len(m.results) {
				m.cursor = 0
			}
		}
	} else {
		m.results = nil
		m.cursor = 0
	}

	return cmd
}

// View renders the search overlay.
func (m *SearchModel) View() string {
	var sb strings.Builder

	sb.WriteString(searchBoxStyle.Render(m.input.View()))
	sb.WriteString("\n\n")

	if len(m.results) == 0 && m.input.Value() != "" {
		sb.WriteString(helpStyle.Render("  No results found"))
	}

	maxResults := m.height - 6
	if maxResults < 3 {
		maxResults = 3
	}
	if maxResults > 15 {
		maxResults = 15
	}

	for i, r := range m.results {
		if i >= maxResults {
			remaining := len(m.results) - maxResults
			sb.WriteString(helpStyle.Render(fmt.Sprintf("\n  ... and %d more results", remaining)))
			break
		}

		dateStr := r.Date.Format("Jan 2")
		line := fmt.Sprintf("  %s  %s", searchResultDateStyle.Render(dateStr), r.Line)

		if i == m.cursor {
			sb.WriteString(searchResultSelectedStyle.Render(line))
		} else {
			sb.WriteString(searchResultStyle.Render(line))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("  [Enter] go to day  [Esc] close  [up/down] navigate"))

	return sb.String()
}

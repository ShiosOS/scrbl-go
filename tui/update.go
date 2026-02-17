package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/juliuswalton/scrbl/notes"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeViewport()
		m.refreshStream(false)
		return m, nil

	case streamLoadedMsg:
		m.loadingMore = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}

		m.days = msg.days
		m.hasMoreDays = msg.hasMore

		if len(m.days) > 0 && msg.anchorDate != "" {
			for i, day := range m.days {
				if day.Date.Format("2006-01-02") == msg.anchorDate {
					m.focusedDay = i
					break
				}
			}
		} else if len(m.days) > 0 && (m.focusedDay < 0 || m.focusedDay >= len(m.days)) {
			m.focusedDay = len(m.days) - 1
		}

		m.refreshStream(true)
		return m, nil

	case syncResultMsg:
		if msg.err != nil {
			m.status = "sync failed"
		} else {
			m.status = "synced"
		}
		return m, nil

	case composerPollMsg:
		if m.mode != modeCompose {
			return m, nil
		}

		saveRequested, quitRequested, quitAfterSave := m.composer.ConsumeRequests()
		if !saveRequested && !quitRequested {
			return m, pollComposerCmd()
		}

		if quitRequested && !saveRequested {
			m.mode = modeStream
			m.status = "stream"
			m.resizeViewport()
			m.refreshStream(false)
			return m, nil
		}

		next, cmd := m.saveCompose()
		if quitAfterSave {
			next.mode = modeStream
			next.status = "stream"
			next.resizeViewport()
			next.refreshStream(false)
			return next, cmd
		}

		if next.mode == modeCompose {
			return next, tea.Batch(cmd, pollComposerCmd())
		}
		return next, cmd

	case tea.KeyMsg:
		if m.mode == modeCompose {
			return m.handleComposeKey(msg)
		}
		return m.handleStreamKey(msg)
	}

	return m, nil
}

func (m Model) handleStreamKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.composer.Close()
		return m, tea.Quit
	case "i":
		return m.startComposeNew()
	case "e", "enter":
		return m.startComposeEditFocused()
	case "r":
		m.status = "reloading"
		return m, m.loadStreamCmd("")
	case "up", "k":
		atTop := m.guideLine == 0
		m.moveGuide(-1)
		if atTop {
			if cmd := m.loadMoreCmd(); cmd != nil {
				return m, cmd
			}
		}
		return m, nil
	case "down", "j":
		m.moveGuide(1)
		return m, nil
	case "pgup", "ctrl+u":
		m.moveGuide(-max(1, m.viewport.Height/2))
		return m, nil
	case "pgdown", "ctrl+d":
		m.moveGuide(max(1, m.viewport.Height/2))
		return m, nil
	case "g":
		m.setGuideLine(0)
		if cmd := m.loadMoreCmd(); cmd != nil {
			return m, cmd
		}
		return m, nil
	case "G":
		m.setGuideLine(len(m.streamLines) - 1)
		return m, nil
	case "[":
		if m.focusedDay > 0 {
			m.focusedDay--
			m.jumpToFocusedDay()
		}
		return m, nil
	case "]":
		if m.focusedDay >= 0 && m.focusedDay < len(m.days)-1 {
			m.focusedDay++
			m.jumpToFocusedDay()
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleComposeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.composer.Close()
		return m, tea.Quit
	case "ctrl+g":
		m.mode = modeStream
		m.status = "stream"
		m.resizeViewport()
		m.refreshStream(false)
		return m, nil
	case "ctrl+s":
		next, cmd := m.saveCompose()
		if next.mode == modeCompose {
			return next, tea.Batch(cmd, pollComposerCmd())
		}
		return next, cmd
	}

	if err := m.composer.Input(msg); err != nil {
		if isSessionClosedError(err) {
			m.composer.Close()
			m.mode = modeStream
			m.status = "stream"
			m.resizeViewport()
			m.refreshStream(false)
			return m, nil
		}
		m.err = err
		return m, nil
	}

	m.snapshot = m.composer.Snapshot()
	return m, nil
}

func (m Model) startComposeNew() (tea.Model, tea.Cmd) {
	if err := m.composer.Start(); err != nil {
		m.err = err
		return m, nil
	}
	if err := m.composer.Clear(); err != nil {
		m.err = err
		return m, nil
	}

	m.snapshot = m.composer.Snapshot()
	m.mode = modeCompose
	m.composeKind = composeNew
	m.composeDay = notes.Today()
	m.status = "compose new"
	m.resizeViewport()
	m.refreshStream(false)

	return m, pollComposerCmd()
}

func (m Model) startComposeEditFocused() (tea.Model, tea.Cmd) {
	if len(m.days) == 0 {
		m.status = "no notes"
		return m, nil
	}
	if m.focusedDay < 0 || m.focusedDay >= len(m.days) {
		m.focusedDay = len(m.days) - 1
	}

	day := m.days[m.focusedDay].Date
	raw, err := m.store.ReadDay(day)
	if err != nil {
		m.err = err
		return m, nil
	}
	if strings.TrimSpace(raw) == "" {
		raw = dayFileTemplate(day)
	}

	if err := m.composer.Start(); err != nil {
		m.err = err
		return m, nil
	}
	if err := m.composer.SetContent(raw, false); err != nil {
		m.err = err
		return m, nil
	}

	m.snapshot = m.composer.Snapshot()
	m.mode = modeCompose
	m.composeKind = composeEdit
	m.composeDay = day
	m.status = "edit " + day.Format("2006-01-02")
	m.resizeViewport()
	m.refreshStream(false)

	return m, pollComposerCmd()
}

func (m Model) saveCompose() (Model, tea.Cmd) {
	if m.composeKind == composeEdit {
		content := strings.ReplaceAll(m.snapshot.Content, "\r\n", "\n")
		if strings.TrimSpace(content) == "" {
			content = dayFileTemplate(m.composeDay)
		}
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}

		if err := m.store.WriteDay(m.composeDay, content); err != nil {
			m.err = err
			return m, nil
		}

		if m.syncer != nil {
			m.status = "syncing..."
		} else {
			m.status = "saved " + m.composeDay.Format("2006-01-02")
		}

		return m, tea.Batch(m.loadStreamCmd(""), m.pushDayCmd(m.composeDay))
	}

	content := strings.TrimSpace(m.snapshot.Content)
	if content == "" {
		m.status = "empty note"
		return m, nil
	}

	today := notes.Today()
	if err := m.store.AppendEntry(today, content); err != nil {
		m.err = err
		return m, nil
	}
	if err := m.composer.Clear(); err != nil {
		m.err = err
		return m, nil
	}

	m.snapshot = m.composer.Snapshot()
	if m.syncer != nil {
		m.status = "syncing..."
	} else {
		m.status = "saved today"
	}

	return m, tea.Batch(m.loadStreamCmd(""), m.pushDayCmd(today))
}

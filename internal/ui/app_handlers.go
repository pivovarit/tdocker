package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

func (m App) handleLogsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "l":
		m = m.closeLogs()
	case "f":
		if m.logs.cancel != nil {
			m.logs.cancel()
		}
		m.logs.allMode = !m.logs.allMode
		m.logs.lines = nil
		m.logs.scrollOffset = 0
		m.logs.autoScroll = true
		m.logs.gen++
		ctx, cancel := context.WithCancel(context.Background())
		m.logs.cancel = cancel
		tail := logsTailDefault
		if m.logs.allMode {
			tail = "all"
		}
		return m, m.client.StartLogs(ctx, m.logs.containerID, tail, m.logs.gen)
	case "up", "k":
		if m.logs.scrollOffset > 0 {
			m.logs.scrollOffset--
			m.logs.autoScroll = false
		}
	case "down", "j":
		maxOffset := max(0, len(m.logs.lines)-(logsPanelHeight-2))
		if m.logs.scrollOffset < maxOffset {
			m.logs.scrollOffset++
		}
		if m.logs.scrollOffset >= maxOffset {
			m.logs.autoScroll = true
		}
	case "g", "home":
		m.logs.scrollOffset = 0
		m.logs.autoScroll = false
	case "G", "end":
		m.logs.scrollOffset = max(0, len(m.logs.lines)-(logsPanelHeight-2))
		m.logs.autoScroll = true
	}
	return m, nil
}

func (m App) handleInspectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "i":
		m = m.closeInspect()
	case "up", "k":
		if m.inspect.offset > 0 {
			m.inspect.offset--
		}
	case "down", "j":
		maxOff := max(0, len(m.inspect.lines)-(inspectPanelHeight-2))
		if m.inspect.offset < maxOff {
			m.inspect.offset++
		}
	case "g", "home":
		m.inspect.offset = 0
	case "G", "end":
		m.inspect.offset = max(0, len(m.inspect.lines)-(inspectPanelHeight-2))
	}
	return m, nil
}

func (m App) handleEventsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "v":
		m = m.closeEvents()
	case "up", "k":
		if m.events.scrollOffset > 0 {
			m.events.scrollOffset--
			m.events.autoScroll = false
		}
	case "down", "j":
		maxOffset := max(0, len(m.events.events)-(eventsPanelHeight-2))
		if m.events.scrollOffset < maxOffset {
			m.events.scrollOffset++
		}
		if m.events.scrollOffset >= maxOffset {
			m.events.autoScroll = true
		}
	case "g", "home":
		m.events.scrollOffset = 0
		m.events.autoScroll = false
	case "G", "end":
		m.events.scrollOffset = max(0, len(m.events.events)-(eventsPanelHeight-2))
		m.events.autoScroll = true
	}
	return m, nil
}

func (m App) handleStatsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "t":
		m = m.closeStats()
	case "r":
		if m.stats.fetching {
			return m, nil
		}
		m.loading = true
		m.err = nil
		if m.stats.entry != nil {
			m.stats.prevEntry = m.stats.entry
		}
		m.stats.entry = nil
		m.stats.fetching = true
		return m, tea.Batch(m.client.FetchContainers(m.showAll), m.client.FetchStats(m.stats.containerID))
	}
	return m, nil
}

func (m App) handleContextKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.ctxPicker.visible = false
		m.ctxPicker.contexts = nil
		m.ctxPicker.cursor = 0
	case "up", "k":
		if m.ctxPicker.cursor > 0 {
			m.ctxPicker.cursor--
		}
	case "down", "j":
		if m.ctxPicker.cursor < len(m.ctxPicker.contexts)-1 {
			m.ctxPicker.cursor++
		}
	case "enter":
		if len(m.ctxPicker.contexts) > 0 {
			return m, m.client.SwitchContext(m.ctxPicker.contexts[m.ctxPicker.cursor].Name)
		}
	}
	return m, nil
}

func (m App) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.op = OpNone
		m.err = nil
		switch m.confirmAction {
		case "stop":
			m.op = OpStopping
			return m, m.client.StopContainer(m.confirmID)
		case "start":
			m.op = OpStarting
			return m, m.client.StartContainer(m.confirmID)
		case "restart":
			m.op = OpRestarting
			return m, m.client.RestartContainer(m.confirmID)
		case "delete":
			m.op = OpDeleting
			return m, m.client.DeleteContainer(m.confirmID)
		}
	case "n", "N", "esc":
		m.op = OpNone
	}
	return m, nil
}

func (m App) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyEnter:
		m.filtering = false
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.filterQuery) > 0 {
			selectedID := m.currentSelectedID()
			runes := []rune(m.filterQuery)
			m.filterQuery = string(runes[:len(runes)-1])
			m = m.rebuildTable(selectedID)
		}
	case tea.KeyRunes:
		selectedID := m.currentSelectedID()
		m.filterQuery += string(msg.Runes)
		m = m.rebuildTable(selectedID)
	}
	return m, nil
}

func (m App) handleMainKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		m.loading = true
		m.err = nil
		return m, m.client.FetchContainers(m.showAll)
	case "A":
		m.showAll = !m.showAll
		m.loading = true
		m.err = nil
		return m, m.client.FetchContainers(m.showAll)
	case "/":
		m.filtering = true
		return m, nil
	case "esc":
		if m.filterQuery != "" {
			selectedID := m.currentSelectedID()
			m.filterQuery = ""
			m = m.rebuildTable(selectedID)
		}
	case "l":
		cursor := m.table.Cursor()
		filtered := m.filtered()
		if cursor >= 0 && cursor < len(filtered) {
			m.logs.container = filtered[cursor].Names
			m.logs.containerID = filtered[cursor].ID
			m.logs.lines = nil
			m.logs.scrollOffset = 0
			m.logs.autoScroll = true
			m.logs.allMode = false
			m.logs.visible = true
			m.logs.gen++
			ctx, cancel := context.WithCancel(context.Background())
			m.logs.cancel = cancel
			firstLine := m.client.StartLogs(ctx, filtered[cursor].ID, logsTailDefault, m.logs.gen)
			m.table.SetHeight(m.tableHeight())
			return m, firstLine
		}
	case "S":
		cursor := m.table.Cursor()
		filtered := m.filtered()
		if cursor >= 0 && cursor < len(filtered) && filtered[cursor].State == "running" {
			m.op = OpConfirming
			m.confirmAction = "stop"
			m.confirmID = filtered[cursor].ID
			m.confirmName = filtered[cursor].Names
			return m, nil
		}
	case "s":
		cursor := m.table.Cursor()
		filtered := m.filtered()
		if cursor >= 0 && cursor < len(filtered) && filtered[cursor].State != "running" {
			m.op = OpConfirming
			m.confirmAction = "start"
			m.confirmID = filtered[cursor].ID
			m.confirmName = filtered[cursor].Names
			return m, nil
		}
	case "R":
		cursor := m.table.Cursor()
		filtered := m.filtered()
		if cursor >= 0 && cursor < len(filtered) && filtered[cursor].State == "running" {
			m.op = OpConfirming
			m.confirmAction = "restart"
			m.confirmID = filtered[cursor].ID
			m.confirmName = filtered[cursor].Names
			return m, nil
		}
	case "D":
		cursor := m.table.Cursor()
		filtered := m.filtered()
		if cursor >= 0 && cursor < len(filtered) && filtered[cursor].State != "running" {
			m.op = OpConfirming
			m.confirmAction = "delete"
			m.confirmID = filtered[cursor].ID
			m.confirmName = filtered[cursor].Names
			return m, nil
		}
	case "e":
		cursor := m.table.Cursor()
		filtered := m.filtered()
		if cursor >= 0 && cursor < len(filtered) && filtered[cursor].State == "running" {
			return m, m.client.ExecContainer(filtered[cursor].ID)
		}
	case "x":
		cursor := m.table.Cursor()
		filtered := m.filtered()
		if cursor >= 0 && cursor < len(filtered) {
			return m, m.client.CheckDebugAvailable(filtered[cursor].ID)
		}
	case "X":
		m.ctxPicker.requested = true
		return m, m.client.FetchContexts()
	case "i":
		cursor := m.table.Cursor()
		filtered := m.filtered()
		if cursor >= 0 && cursor < len(filtered) {
			m.inspect.visible = true
			m.inspect.lines = nil
			m.inspect.offset = 0
			m.inspect.container = filtered[cursor].Names
			m.table.SetHeight(m.tableHeight())
			return m, m.client.InspectContainer(filtered[cursor].ID)
		}
	case "c":
		cursor := m.table.Cursor()
		filtered := m.filtered()
		if cursor >= 0 && cursor < len(filtered) {
			c := filtered[cursor]
			return m, copyToClipboard(c.Names, c.ID)
		}
	case "t":
		cursor := m.table.Cursor()
		filtered := m.filtered()
		if cursor >= 0 && cursor < len(filtered) && filtered[cursor].State == "running" {
			m.stats.visible = true
			m.stats.entry = nil
			m.stats.container = filtered[cursor].Names
			m.stats.containerID = filtered[cursor].ID
			m.stats.fetching = true
			m.table.SetHeight(m.tableHeight())
			return m, m.client.FetchStats(filtered[cursor].ID)
		}
	case "v":
		if m.events.visible {
			m = m.closeEvents()
		} else {
			m.events.visible = true
			m.events.events = nil
			m.events.scrollOffset = 0
			m.events.autoScroll = true
			m.events.gen++
			ctx, cancel := context.WithCancel(context.Background())
			m.events.cancel = cancel
			m.table.SetHeight(m.tableHeight())
			return m, m.client.StartEvents(ctx, m.events.gen)
		}
	}

	switch msg.String() {
	case "j":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "k":
		msg = tea.KeyMsg{Type: tea.KeyUp}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	cursor := m.table.Cursor()
	height := m.tableHeight()
	if cursor < m.viewportStart {
		m.viewportStart = cursor
	} else if height > 0 && cursor >= m.viewportStart+height {
		m.viewportStart = cursor - height + 1
	}
	return m, cmd
}

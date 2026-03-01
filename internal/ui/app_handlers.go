package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

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
			m.logs.scroll = scrollState{autoScroll: true}
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
			m.inspect.scroll = scrollState{}
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
			m.events.scroll = scrollState{autoScroll: true}
			m.table.SetHeight(m.tableHeight())
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

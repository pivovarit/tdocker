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
		if m.logsCancel != nil {
			m.logsCancel()
		}
		m.logsAllMode = !m.logsAllMode
		m.logsLines = nil
		m.logsScrollOffset = 0
		m.logsAutoScroll = true
		m.logsGen++
		ctx, cancel := context.WithCancel(context.Background())
		m.logsCancel = cancel
		tail := "200"
		if m.logsAllMode {
			tail = "all"
		}
		return m, m.client.StartLogs(ctx, m.logsContainerID, tail, m.logsGen)
	case "up", "k":
		if m.logsScrollOffset > 0 {
			m.logsScrollOffset--
			m.logsAutoScroll = false
		}
	case "down", "j":
		maxOffset := max(0, len(m.logsLines)-(logsPanelHeight-2))
		if m.logsScrollOffset < maxOffset {
			m.logsScrollOffset++
		}
		if m.logsScrollOffset >= maxOffset {
			m.logsAutoScroll = true
		}
	case "g", "home":
		m.logsScrollOffset = 0
		m.logsAutoScroll = false
	case "G", "end":
		m.logsScrollOffset = max(0, len(m.logsLines)-(logsPanelHeight-2))
		m.logsAutoScroll = true
	}
	return m, nil
}

func (m App) handleInspectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "i":
		m = m.closeInspect()
	case "up", "k":
		if m.inspectOffset > 0 {
			m.inspectOffset--
		}
	case "down", "j":
		maxOff := max(0, len(m.inspectLines)-(inspectPanelHeight-2))
		if m.inspectOffset < maxOff {
			m.inspectOffset++
		}
	case "g", "home":
		m.inspectOffset = 0
	case "G", "end":
		m.inspectOffset = max(0, len(m.inspectLines)-(inspectPanelHeight-2))
	}
	return m, nil
}

func (m App) handleStatsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "t":
		m = m.closeStats()
	case "r":
		if m.statsFetching {
			return m, nil
		}
		m.loading = true
		m.err = nil
		m.statsEntry = nil
		m.statsFetching = true
		return m, tea.Batch(m.client.FetchContainers(m.showAll), m.client.FetchStats(m.statsContainerID))
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
	case "a":
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
			m.logsContainer = filtered[cursor].Names
			m.logsContainerID = filtered[cursor].ID
			m.logsLines = nil
			m.logsScrollOffset = 0
			m.logsAutoScroll = true
			m.logsAllMode = false
			m.logsVisible = true
			m.logsGen++
			ctx, cancel := context.WithCancel(context.Background())
			m.logsCancel = cancel
			firstLine := m.client.StartLogs(ctx, filtered[cursor].ID, "200", m.logsGen)
			m.table.SetHeight(m.tableHeight())
			return m, firstLine
		}
	case "s":
		cursor := m.table.Cursor()
		filtered := m.filtered()
		if cursor >= 0 && cursor < len(filtered) && filtered[cursor].State == "running" {
			m.op = OpConfirming
			m.confirmAction = "stop"
			m.confirmID = filtered[cursor].ID
			m.confirmName = filtered[cursor].Names
			return m, nil
		}
	case "S":
		cursor := m.table.Cursor()
		filtered := m.filtered()
		if cursor >= 0 && cursor < len(filtered) && filtered[cursor].State != "running" {
			m.op = OpConfirming
			m.confirmAction = "start"
			m.confirmID = filtered[cursor].ID
			m.confirmName = filtered[cursor].Names
			return m, nil
		}
	case "d":
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
	case "i":
		cursor := m.table.Cursor()
		filtered := m.filtered()
		if cursor >= 0 && cursor < len(filtered) {
			m.inspectVisible = true
			m.inspectLines = nil
			m.inspectOffset = 0
			m.inspectContainer = filtered[cursor].Names
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
			m.statsVisible = true
			m.statsEntry = nil
			m.statsContainer = filtered[cursor].Names
			m.statsContainerID = filtered[cursor].ID
			m.statsFetching = true
			m.table.SetHeight(m.tableHeight())
			return m, m.client.FetchStats(filtered[cursor].ID)
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

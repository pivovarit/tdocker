package ui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

type confirmEntry struct {
	op     Operation
	execFn func(docker.Client, string) tea.Cmd
}

var confirmActions = map[string]confirmEntry{
	"stop":    {OpStopping, docker.Client.StopContainer},
	"start":   {OpStarting, docker.Client.StartContainer},
	"restart": {OpRestarting, docker.Client.RestartContainer},
	"delete":  {OpDeleting, docker.Client.DeleteContainer},
}

func (m App) handleConfirmKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.Code {
	case 'y', 'Y':
		if entry, ok := confirmActions[m.op.action]; ok {
			m.err = nil
			m.op.gen++
			m.op.kind = entry.op
			return m, tea.Batch(entry.execFn(m.client, m.op.id), opDisplayCmd(m.op.gen), opSlowCmd(m.op.gen))
		}
	case 'n', 'N', tea.KeyEsc:
		m.op = operationState{}
	}
	return m, nil
}

func (m App) handleFilterKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.Code {
	case tea.KeyUp, tea.KeyDown:
		m.filtering = false
		return m.handleMainKey(msg)
	case tea.KeyEsc, tea.KeyEnter:
		m.filtering = false
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.filterQuery) > 0 {
			selectedID := m.currentSelectedID()
			runes := []rune(m.filterQuery)
			m.filterQuery = string(runes[:len(runes)-1])
			m = m.rebuildTable(selectedID)
		}
	default:
		if len(msg.Text) > 0 {
			selectedID := m.currentSelectedID()
			m.filterQuery += msg.Text
			m = m.rebuildTable(selectedID)
		}
	}
	return m, nil
}

func (m App) handleRenameKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.Code {
	case tea.KeyEsc:
		m.rename = renameState{}
	case tea.KeyEnter:
		newName := strings.TrimSpace(m.rename.input)
		if newName == "" {
			m.rename = renameState{}
			return m, nil
		}
		id := m.rename.id
		m.rename = renameState{}
		m.op.gen++
		m.op.kind = OpRenaming
		return m, tea.Batch(m.client.RenameContainer(id, newName), opDisplayCmd(m.op.gen), opSlowCmd(m.op.gen))
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.rename.input) > 0 {
			runes := []rune(m.rename.input)
			m.rename.input = string(runes[:len(runes)-1])
		}
	default:
		if len(msg.Text) > 0 {
			m.rename.input += msg.Text
		}
	}
	return m, nil
}

func (m App) handleMainKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	cursor := m.table.Cursor()
	filtered := m.filtered()

	switch msg.Text {
	case keyRefresh:
		m.fetch.loading = true
		m.err = nil
		return m, m.client.FetchContainers(m.showAll)
	case keyToggleAll:
		m.showAll = !m.showAll
		m.fetch.loading = true
		m.err = nil
		return m, m.client.FetchContainers(m.showAll)
	case keyFilter:
		m.filtering = true
		return m, nil
	case keyLogs:
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
			firstLine := m.client.StartLogs(ctx, filtered[cursor].ID, logsTailDefault, false, "", m.logs.gen)
			m.table.SetHeight(m.tableHeight())
			return m, firstLine
		}
	case keyStop:
		if cursor >= 0 && cursor < len(filtered) {
			c := filtered[cursor]
			action := "start"
			if c.State == "running" {
				action = "stop"
			}
			m.op = operationState{kind: OpConfirming, id: c.ID, name: c.Names, action: action}
			return m, nil
		}
	case keyRestart:
		if cursor >= 0 && cursor < len(filtered) {
			c := filtered[cursor]
			action := "start"
			if c.State == "running" {
				action = "restart"
			}
			m.op = operationState{kind: OpConfirming, id: c.ID, name: c.Names, action: action}
			return m, nil
		}
	case keyDelete:
		if cursor >= 0 && cursor < len(filtered) {
			if filtered[cursor].State == "running" {
				m.warnMsg = "stop the container before deleting"
				return m, nil
			}
			c := filtered[cursor]
			m.op = operationState{kind: OpConfirming, id: c.ID, name: c.Names, action: "delete"}
			return m, nil
		}
	case keyPause:
		if cursor >= 0 && cursor < len(filtered) {
			c := filtered[cursor]
			m.op.gen++
			gen := m.op.gen
			if c.State == "running" {
				m.op.kind = OpPausing
				return m, tea.Batch(m.client.PauseContainer(c.ID), opDisplayCmd(gen), opSlowCmd(gen))
			} else if c.State == "paused" {
				m.op.kind = OpUnpausing
				return m, tea.Batch(m.client.UnpauseContainer(c.ID), opDisplayCmd(gen), opSlowCmd(gen))
			}
		}
	case keyRename:
		if cursor >= 0 && cursor < len(filtered) {
			c := filtered[cursor]
			m.rename = renameState{active: true, id: c.ID, input: strings.TrimPrefix(c.Names, "/")}
			return m, nil
		}
	case keyExec:
		if cursor >= 0 && cursor < len(filtered) && filtered[cursor].State == "running" {
			return m, m.client.CheckShellAvailable(filtered[cursor].ID)
		}
	case keyDebug:
		if cursor >= 0 && cursor < len(filtered) {
			return m, m.client.CheckDebugAvailable(filtered[cursor].ID)
		}
	case keyContext:
		m.ctxPicker.requested = true
		return m, m.client.FetchContexts()
	case keyInspect:
		if cursor >= 0 && cursor < len(filtered) {
			m.inspect.visible = true
			m.inspect.lines = nil
			m.inspect.scroll = scrollState{}
			m.inspect.container = filtered[cursor].Names
			m.table.SetHeight(m.tableHeight())
			return m, m.client.InspectContainer(filtered[cursor].ID)
		}
	case keyCopy:
		if cursor >= 0 && cursor < len(filtered) {
			c := filtered[cursor]
			return m, copyToClipboard(c.Names, c.ID)
		}
	case keyStats:
		if cursor >= 0 && cursor < len(filtered) && filtered[cursor].State == "running" {
			m.stats.visible = true
			m.stats.entry = nil
			m.stats.container = filtered[cursor].Names
			m.stats.containerID = filtered[cursor].ID
			m.stats.fetching = true
			m.table.SetHeight(m.tableHeight())
			return m, m.client.FetchStats(filtered[cursor].ID)
		}
	case keyEvents:
		if m.events.visible {
			m = m.closeEvents()
		} else {
			m.events.visible = true
			m.events.events = nil
			m.events.scroll = scrollState{autoScroll: true}
			m.table.SetHeight(m.tableHeight())
		}
	case keyHelp:
		m.helpVisible = true
		return m, nil
	default:
		if msg.Code == tea.KeyEsc {
			if m.filterQuery != "" {
				selectedID := ""
				if cursor >= 0 && cursor < len(filtered) {
					selectedID = filtered[cursor].ID
				}
				m.filterQuery = ""
				m = m.rebuildTable(selectedID)
			}
		}
	}

	var tableMsg tea.Msg = msg
	switch msg.Text {
	case keyVimDown:
		tableMsg = tea.KeyPressMsg{Code: tea.KeyDown}
	case keyVimUp:
		tableMsg = tea.KeyPressMsg{Code: tea.KeyUp}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(tableMsg)
	cursor = m.table.Cursor()
	height := m.tableHeight()
	if cursor < m.viewportStart {
		m.viewportStart = cursor
	} else if height > 0 && cursor >= m.viewportStart+height {
		m.viewportStart = cursor - height + 1
	}
	return m, cmd
}

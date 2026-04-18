package ui

import (
	"context"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

type confirmEntry struct {
	op     Operation
	execFn func(docker.Client, string) tea.Cmd
}

var confirmActions = map[string]confirmEntry{
	"stop":            {OpStopping, docker.Client.StopContainer},
	"start":           {OpStarting, docker.Client.StartContainer},
	"restart":         {OpRestarting, docker.Client.RestartContainer},
	"delete":          {OpDeleting, docker.Client.DeleteContainer},
	"compose-stop":    {OpStopping, docker.Client.StopCompose},
	"compose-start":   {OpStarting, docker.Client.StartCompose},
	"compose-restart": {OpRestarting, docker.Client.RestartCompose},
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
			m.filterQuery = trimLastRune(m.filterQuery)
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
			m.rename.input = trimLastRune(m.rename.input)
		}
	default:
		if len(msg.Text) > 0 {
			m.rename.input += msg.Text
		}
	}
	return m, nil
}

func (m App) handleMainKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.Text {
	case keyRefresh:
		m.err = nil
		return m.startFetch()
	case keyToggleAll:
		m.showAll = !m.showAll
		m.err = nil
		return m.startFetch()
	case keyFilter:
		m.filtering = true
		return m, nil
	case keyLogs:
		if c, ok := m.selectedContainer(); ok {
			if c.State == docker.StateCollapsed {
				proj := c.ComposeProject()
				if m.logs.cancel != nil {
					m.logs.cancel()
				}
				m.logs.container = proj
				m.logs.containerID = ""
				m.logs.isCompose = true
				m.logs.composeProject = proj
				m.logs.lines = nil
				m.logs.scroll = scrollState{autoScroll: true}
				m.logs.allMode = false
				m.logs.visible = true
				m.logs.gen++
				ctx, cancel := context.WithCancel(context.Background())
				m.logs.cancel = cancel
				firstLine := m.client.StartComposeLogs(ctx, proj, logsTailDefault, false, m.logs.gen)
				m.table.SetHeight(m.tableHeight())
				return m, firstLine
			} else if c.ID != "" {
				if m.logs.cancel != nil {
					m.logs.cancel()
				}
				m.logs.container = c.Names
				m.logs.containerID = c.ID
				m.logs.isCompose = false
				m.logs.composeProject = ""
				m.logs.lines = nil
				m.logs.scroll = scrollState{autoScroll: true}
				m.logs.allMode = false
				m.logs.visible = true
				m.logs.gen++
				ctx, cancel := context.WithCancel(context.Background())
				m.logs.cancel = cancel
				firstLine := m.client.StartLogs(ctx, c.ID, logsTailDefault, false, "", m.logs.gen)
				m.table.SetHeight(m.tableHeight())
				return m, firstLine
			}
		}
	case keyStop:
		if c, ok := m.selectedContainer(); ok {
			if c.State == docker.StateCollapsed {
				proj := c.ComposeProject()
				action := "compose-start"
				if m.projectHasRunning(proj) {
					action = "compose-stop"
				}
				m.op = operationState{kind: OpConfirming, id: proj, name: c.Names, action: action}
				return m, nil
			}
			if c.ID != "" {
				action := "start"
				if c.State == docker.StateRunning {
					action = "stop"
				}
				m.op = operationState{kind: OpConfirming, id: c.ID, name: c.Names, action: action}
				return m, nil
			}
		}
	case keyRestart:
		if c, ok := m.selectedContainer(); ok {
			if c.State == docker.StateCollapsed {
				proj := c.ComposeProject()
				action := "compose-start"
				if m.projectHasRunning(proj) {
					action = "compose-restart"
				}
				m.op = operationState{kind: OpConfirming, id: proj, name: c.Names, action: action}
				return m, nil
			}
			if c.ID != "" {
				action := "start"
				if c.State == docker.StateRunning {
					action = "restart"
				}
				m.op = operationState{kind: OpConfirming, id: c.ID, name: c.Names, action: action}
				return m, nil
			}
		}
	case keyDelete:
		if c, ok := m.selectedContainer(); ok && c.ID != "" {
			if c.State == docker.StateRunning {
				m.warnMsg = "stop the container before deleting"
				return m, nil
			}
			m.op = operationState{kind: OpConfirming, id: c.ID, name: c.Names, action: "delete"}
			return m, nil
		}
	case keyPause:
		if c, ok := m.selectedContainer(); ok && c.ID != "" {
			m.op.gen++
			gen := m.op.gen
			if c.State == docker.StateRunning {
				m.op.kind = OpPausing
				return m, tea.Batch(m.client.PauseContainer(c.ID), opDisplayCmd(gen), opSlowCmd(gen))
			} else if c.State == docker.StatePaused {
				m.op.kind = OpUnpausing
				return m, tea.Batch(m.client.UnpauseContainer(c.ID), opDisplayCmd(gen), opSlowCmd(gen))
			}
		}
	case keyRename:
		if c, ok := m.selectedContainer(); ok && c.ID != "" {
			m.rename = renameState{active: true, id: c.ID, input: strings.TrimPrefix(c.Names, "/")}
			return m, nil
		}
	case keyExec:
		if c, ok := m.selectedContainer(); ok && c.ID != "" && c.State == docker.StateRunning {
			return m, m.client.CheckShellAvailable(c.ID)
		}
	case keyDebug:
		if c, ok := m.selectedContainer(); ok && c.ID != "" {
			return m, m.client.CheckDebugAvailable(c.ID)
		}
	case keyContext:
		m.ctxPicker.requested = true
		return m, m.client.FetchContexts()
	case keyInspect:
		if c, ok := m.selectedContainer(); ok && c.ID != "" {
			m.inspect.visible = true
			m.inspect.lines = nil
			m.inspect.scroll = scrollState{}
			m.inspect.container = c.Names
			m.table.SetHeight(m.tableHeight())
			return m, m.client.InspectContainer(c.ID)
		}
	case keyDiagnose:
		if c, ok := m.selectedContainer(); ok && c.ID != "" && c.State != docker.StateCollapsed && c.State != docker.StateDetail {
			m.diagnostic.visible = true
			m.diagnostic.loading = true
			m.diagnostic.container = c.Names
			m.diagnostic.containerID = c.ID
			m.diagnostic.scroll = scrollState{}
			m.diagnostic.logsGen++
			ctx, cancel := context.WithCancel(context.Background())
			m.diagnostic.logsCancel = cancel
			m.table.SetHeight(m.tableHeight())
			return m, tea.Batch(
				m.client.InspectContainer(c.ID),
				m.client.FetchContainerEvents(c.ID, time.Hour),
				m.client.StartLogs(ctx, c.ID, "50", false, "", m.diagnostic.logsGen),
			)
		}
	case keyCopy:
		if c, ok := m.selectedContainer(); ok {
			if c.State == docker.StateDetail {
				content := detailRowContent(c.Names)
				return m, copyToClipboard(content, content)
			}
			if c.ID != "" {
				return m, copyToClipboard(c.Names, c.ID)
			}
		}
	case keyStats:
		if c, ok := m.selectedContainer(); ok && c.ID != "" && c.State == docker.StateRunning {
			m.stats.visible = true
			m.stats.entry = nil
			m.stats.prevEntry = nil
			m.stats.container = c.Names
			m.stats.containerID = c.ID
			if e, ok := m.inlineStats[c.ID]; ok {
				m.stats.entry = &e
			}
			m.table.SetHeight(m.tableHeight())
			return m, nil
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
	case keyInlineStats:
		m.showInlineStats = !m.showInlineStats
		m = m.rebuildTable(m.currentSelectedID())
		return m, nil
	case keyHelp:
		m.helpVisible = true
		return m, nil
	default:
		if msg.Code == tea.KeyEsc {
			if m.filterQuery != "" {
				m.filterQuery = ""
				m = m.rebuildTable(m.currentSelectedID())
			}
		}
	}

	switch msg.Code {
	case tea.KeyLeft:
		if c, ok := m.selectedContainer(); ok {
			proj := c.ComposeProject()
			if c.ID != "" {
				if _, expanded := m.expandedContainers[c.ID]; expanded {
					delete(m.expandedContainers, c.ID)
					m = m.rebuildTable(c.ID)
					return m, nil
				}
			}
			if proj != "" && c.State != docker.StateCollapsed {
				m.collapsedProjects[proj] = true
				m = m.rebuildTable("")
				for i, fc := range m.filtered() {
					if fc.State == docker.StateCollapsed && fc.ComposeProject() == proj {
						m.table.SetCursor(i)
						m = m.ensureCursorVisible()
						break
					}
				}
				return m, nil
			}
		}
		return m, nil
	case tea.KeyRight:
		if c, ok := m.selectedContainer(); ok {
			proj := c.ComposeProject()
			if c.State == docker.StateCollapsed && proj != "" {
				delete(m.collapsedProjects, proj)
				m = m.rebuildTable("")
				filtered := m.filtered()
				for i, fc := range filtered {
					if fc.ComposeProject() == proj {
						m.table.SetCursor(i)
						break
					}
				}
				m = m.rebuildTable(m.currentSelectedID())
				return m, nil
			}
			if c.ID != "" {
				if _, alreadyExpanded := m.expandedContainers[c.ID]; alreadyExpanded {
					return m, nil
				}
				m.expandedContainers[c.ID] = nil
				m = m.rebuildTable(c.ID)
				if m.expandedContainers[c.ID] != nil {
					return m, nil
				}
				return m, m.client.InspectContainerExpand(c.ID)
			}
		}
		return m, nil
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
	m = m.ensureCursorVisible()
	return m, cmd
}

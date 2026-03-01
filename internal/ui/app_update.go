package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pivovarit/tdocker/internal/docker"
)

type statsTickMsg struct{}

func isContainerLifecycleEvent(ev docker.Event) bool {
	if ev.Type != "container" {
		return false
	}
	switch ev.Action {
	case "start", "stop", "die", "kill", "create", "destroy", "pause", "unpause", "rename", "oom":
		return true
	}
	return false
}

func statsTickCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(2 * time.Second)
		return statsTickMsg{}
	}
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.rebuildTable(m.currentSelectedID())
		return m, nil

	case tea.KeyMsg:
		m.copiedName = ""
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			if m.logs.visible {
				m = m.closeLogs()
			}
			return m, tea.Quit
		}
		if m.logs.visible {
			return m.handleLogsKey(msg)
		}
		if m.inspect.visible {
			return m.handleInspectKey(msg)
		}
		if m.stats.visible {
			return m.handleStatsKey(msg)
		}
		if m.events.visible {
			return m.handleEventsKey(msg)
		}
		if m.op == OpConfirming {
			return m.handleConfirmKey(msg)
		}
		if m.filtering {
			return m.handleFilterKey(msg)
		}
		if m.ctxPicker.visible {
			return m.handleContextKey(msg)
		}
		return m.handleMainKey(msg)

	case docker.ContainersMsg:
		selectedID := m.currentSelectedID()
		m.containers = msg
		m.sorted = docker.Sort(m.containers)
		m.containersByID = indexContainers(m.containers)
		m.loading = false
		m.err = nil
		m = m.rebuildTable(selectedID)
		return m, nil

	case docker.ErrMsg:
		m.err = msg.Err
		m.loading = false
		return m, nil

	case docker.StopMsg:
		m.op = OpNone
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		return m, m.client.FetchContainers(m.showAll)

	case docker.StartMsg:
		m.op = OpNone
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		return m, m.client.FetchContainers(m.showAll)

	case docker.RestartMsg:
		m.op = OpNone
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		return m, m.client.FetchContainers(m.showAll)

	case docker.DeleteMsg:
		m.op = OpNone
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		selectedID := m.currentSelectedID()
		kept := m.containers[:0]
		for _, c := range m.containers {
			if c.ID != msg.ID {
				kept = append(kept, c)
			}
		}
		m.containers = kept
		m.sorted = docker.Sort(m.containers)
		m.containersByID = indexContainers(m.containers)
		m = m.rebuildTable(selectedID)
		return m, nil

	case clipboardMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.copiedName = msg.name
		}
		return m, nil

	case docker.LogsLineMsg:
		if !m.logs.visible || msg.Gen != m.logs.gen {
			return m, nil
		}
		m.logs.lines = append(m.logs.lines, msg.Line)
		if m.logs.autoScroll {
			m.logs.scrollOffset = max(0, len(m.logs.lines)-(logsPanelHeight-2))
		}
		return m, msg.Next

	case docker.LogsEndMsg:
		if !m.logs.visible || msg.Gen != m.logs.gen {
			return m, nil
		}
		if msg.Err != nil {
			m.err = msg.Err
			m = m.closeLogs()
		}
		return m, nil

	case docker.InspectMsg:
		if !m.inspect.visible {
			return m, nil
		}
		if msg.Err != nil {
			m.err = msg.Err
			m = m.closeInspect()
			return m, nil
		}
		m.inspect.lines = buildInspectLines(msg.Data, m.width)
		return m, nil

	case docker.DebugAvailableMsg:
		if !msg.Available {
			m.err = fmt.Errorf("docker debug is not available (requires Docker Desktop or the debug plugin)")
			return m, nil
		}
		return m, m.client.DebugContainer(msg.ID)

	case docker.ExecDoneMsg:
		m.loading = true
		return m, m.client.FetchContainers(m.showAll)

	case docker.StatsMsg:
		m.stats.fetching = false
		if !m.stats.visible {
			return m, nil
		}
		if msg.Err != nil {
			m.err = msg.Err
			m = m.closeStats()
			return m, nil
		}
		if m.stats.entry != nil {
			m.stats.prevEntry = m.stats.entry
		}
		m.stats.entry = &msg.Entry
		return m, statsTickCmd()

	case statsTickMsg:
		if !m.stats.visible || m.stats.fetching {
			return m, nil
		}
		m.stats.fetching = true
		return m, m.client.FetchStats(m.stats.containerID)

	case docker.ContextsMsg:
		m.ctxPicker.contexts = []docker.DockerContext(msg)
		for i, c := range m.ctxPicker.contexts {
			if c.Current {
				m.ctxPicker.current = c.Name
				if m.ctxPicker.requested {
					m.ctxPicker.cursor = i
				}
				break
			}
		}
		if m.ctxPicker.requested {
			m.ctxPicker.visible = true
			m.ctxPicker.requested = false
		}
		return m, nil

	case docker.EventLineMsg:
		if !m.events.visible || msg.Gen != m.events.gen {
			return m, msg.Next
		}
		const maxEvents = 500
		if len(m.events.events) >= maxEvents {
			m.events.events = m.events.events[1:]
		}
		m.events.events = append(m.events.events, msg.Event)
		if m.events.autoScroll {
			m.events.scrollOffset = max(0, len(m.events.events)-(eventsPanelHeight-2))
		}
		var refreshCmd tea.Cmd
		if !m.loading && isContainerLifecycleEvent(msg.Event) {
			m.loading = true
			refreshCmd = m.client.FetchContainers(m.showAll)
		}
		return m, tea.Batch(msg.Next, refreshCmd)

	case docker.EventEndMsg:
		if !m.events.visible || msg.Gen != m.events.gen {
			return m, nil
		}
		if msg.Err != nil {
			m.err = msg.Err
			m = m.closeEvents()
		}
		return m, nil

	case docker.ContextSwitchMsg:
		m.ctxPicker.visible = false
		m.ctxPicker.contexts = nil
		m.ctxPicker.cursor = 0
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		m.err = nil
		return m, tea.Batch(m.client.FetchContainers(m.showAll), m.client.FetchContexts())
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

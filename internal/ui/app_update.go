package ui

import (
	"context"
	"fmt"
	"slices"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

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

func fetchTimerCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(100 * time.Millisecond)
		return fetchTimerTickMsg{}
	}
}

func fetchSlowCmd(gen int) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(30 * time.Second)
		return fetchSlowMsg{gen: gen}
	}
}

func opSlowCmd(gen int) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(10 * time.Second)
		return opSlowMsg{gen: gen}
	}
}

func opDisplayCmd(gen int) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(150 * time.Millisecond)
		return opDisplayMsg{gen: gen}
	}
}

func (m App) startFetch() (App, tea.Cmd) {
	m.fetch.loading = true
	m.fetch.start = time.Now()
	m.fetch.gen++
	m.fetch.slow = false
	return m, tea.Batch(m.client.FetchContainers(m.showAll), fetchTimerCmd(), fetchSlowCmd(m.fetch.gen))
}

func (m App) handleLifecycleMsg(err error) (tea.Model, tea.Cmd) {
	m.op = operationState{}
	m.warnMsg = ""
	if err != nil {
		m.err = err
		return m, nil
	}
	return m.startFetch()
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.rebuildTable(m.currentSelectedID())
		return m, nil

	case tea.KeyPressMsg:
		m.copiedName = ""
		m.warnMsg = ""
		if msg.String() == keyForceQuit {
			if m.logs.visible {
				m = m.closeLogs()
			}
			return m, tea.Quit
		}
		if msg.String() == keyQuit {
			switch {
			case m.rename.active:
				return m.handleRenameKey(msg)
			case m.helpVisible:
				m.helpVisible = false
				return m, nil
			case m.logs.searching:
				return m.handleLogsKey(msg)
			case m.logs.visible && m.logs.searchQuery != "":
				wasGrep := m.logs.grepMode
				m.logs.searchQuery = ""
				m.logs.grepMode = false
				m.logs.scroll = scrollState{autoScroll: true}
				if wasGrep {
					return m.restartLogs()
				}
				m.logs.scroll.offset = max(0, len(m.logs.lines)-(m.logsPanelHeight()-2))
				return m, nil
			case m.logs.visible:
				m = m.closeLogs()
				return m, nil
			case m.inspect.visible:
				m = m.closeInspect()
				return m, nil
			case m.stats.visible:
				m = m.closeStats()
				return m, nil
			case m.events.visible:
				m = m.closeEvents()
				return m, nil
			case m.ctxPicker.visible:
				m.ctxPicker = ctxPickerState{}
				return m, nil
			case m.filterQuery != "":
				m.filterQuery = ""
				m = m.rebuildTable("")
				return m, nil
			default:
				return m, tea.Quit
			}
		}
		if m.helpVisible {
			if msg.String() == keyHelp || msg.Code == tea.KeyEsc {
				m.helpVisible = false
			}
			return m, nil
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
		if m.op.kind == OpConfirming {
			return m.handleConfirmKey(msg)
		}
		if m.rename.active {
			return m.handleRenameKey(msg)
		}
		if m.filtering {
			return m.handleFilterKey(msg)
		}
		if m.ctxPicker.visible {
			return m.handleContextKey(msg)
		}
		return m.handleMainKey(msg)

	case fetchTimerTickMsg:
		if m.fetch.loading {
			m.fetch.visible = true
			return m, fetchTimerCmd()
		}
		return m, nil

	case opDisplayMsg:
		if msg.gen == m.op.gen && m.op.kind != OpNone && m.op.kind != OpConfirming {
			m.op.visible = true
		}
		return m, nil

	case updateAvailableMsg:
		m.updateAvailable = msg.version
		return m, nil

	case fetchSlowMsg:
		if m.fetch.loading && msg.gen == m.fetch.gen {
			m.fetch.slow = true
		}
		return m, nil

	case opSlowMsg:
		if msg.gen == m.op.gen && m.op.kind != OpNone && m.op.kind != OpConfirming {
			m.warnMsg = "Docker is taking a long time to respond…"
		}
		return m, nil

	case docker.ContainersMsg:
		selectedID := m.currentSelectedID()
		m.containers = msg
		m.sorted = docker.Sort(m.containers)
		m.containersByID = indexContainers(m.containers)
		m.fetch.loading = false
		m.fetch.visible = false
		m.fetch.slow = false
		m.err = nil
		m = m.rebuildTable(selectedID)
		return m, nil

	case docker.ErrMsg:
		m.err = msg.Err
		m.fetch.loading = false
		return m, nil

	case docker.StopMsg:
		return m.handleLifecycleMsg(msg.Err)

	case docker.StartMsg:
		return m.handleLifecycleMsg(msg.Err)

	case docker.RestartMsg:
		return m.handleLifecycleMsg(msg.Err)

	case docker.PauseMsg:
		return m.handleLifecycleMsg(msg.Err)

	case docker.UnpauseMsg:
		return m.handleLifecycleMsg(msg.Err)

	case docker.RenameMsg:
		return m.handleLifecycleMsg(msg.Err)

	case docker.DeleteMsg:
		m.op = operationState{}
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
		if m.logs.scroll.autoScroll {
			filtered := m.logsFiltered()
			m.logs.scroll.offset = max(0, len(filtered)-(m.logsPanelHeight()-2))
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

	case docker.ShellAvailableMsg:
		if !msg.Available {
			m.err = fmt.Errorf("shell not found in container (distroless/scratch image?) - press 'x' to use docker debug")
			return m, nil
		}
		return m, m.client.ExecContainer(msg.ID, msg.Shell)

	case docker.DebugAvailableMsg:
		if !msg.Available {
			m.err = fmt.Errorf("docker debug is not available (requires Docker Desktop or the debug plugin)")
			return m, nil
		}
		return m, m.client.DebugContainer(msg.ID)

	case docker.ExecDoneMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		return m.startFetch()

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
		if i := slices.IndexFunc(m.ctxPicker.contexts, func(c docker.DockerContext) bool { return c.Current }); i >= 0 {
			m.ctxPicker.current = m.ctxPicker.contexts[i].Name
			if m.ctxPicker.requested {
				m.ctxPicker.cursor = i
				if i >= ctxPanelMaxRows {
					m.ctxPicker.viewportStart = i - ctxPanelMaxRows + 1
				}
			}
		}
		if m.ctxPicker.requested {
			m.ctxPicker.visible = true
			m.ctxPicker.requested = false
		}
		return m, nil

	case docker.EventLineMsg:
		if msg.Gen != m.bgEventsGen {
			return m, msg.Next
		}
		var debounceCmd tea.Cmd
		opIdle := m.op.kind == OpNone || m.op.kind == OpConfirming
		if !m.fetch.loading && !m.pendingRefresh && opIdle && isContainerLifecycleEvent(msg.Event) {
			m.pendingRefresh = true
			debounceCmd = tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg {
				return autoRefreshMsg{}
			})
		}
		if m.events.visible {
			const maxEvents = 500
			if len(m.events.events) >= maxEvents {
				m.events.events = m.events.events[1:]
			}
			m.events.events = append(m.events.events, msg.Event)
			if m.events.scroll.autoScroll {
				m.events.scroll.offset = max(0, len(m.events.events)-(m.eventsPanelHeight()-2))
			}
		}
		return m, tea.Batch(msg.Next, debounceCmd)

	case autoRefreshMsg:
		m.pendingRefresh = false
		if !m.fetch.loading {
			return m.startFetch()
		}
		return m, nil

	case docker.EventEndMsg:
		if msg.Gen != m.bgEventsGen {
			return m, nil
		}
		m.bgEventsGen++
		newGen := m.bgEventsGen
		return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
			return bgEventsRestartMsg{gen: newGen}
		})

	case bgEventsRestartMsg:
		if msg.gen != m.bgEventsGen {
			return m, nil
		}
		return m, m.client.StartEvents(context.Background(), m.bgEventsGen)

	case docker.GrepSupportMsg:
		m.grepSupported = msg.Available
		return m, nil

	case docker.ContextSwitchMsg:
		m.ctxPicker.visible = false
		m.ctxPicker.contexts = nil
		m.ctxPicker.cursor = 0
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.err = nil
		m, fetchCmd := m.startFetch()
		return m, tea.Batch(fetchCmd, m.client.FetchContexts())
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

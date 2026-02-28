package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pivovarit/tdocker/internal/docker"
)

const (
	logsPanelHeight    = 15
	inspectPanelHeight = 20
	statsPanelHeight   = 9
)

type App struct {
	table            table.Model
	containers       []docker.Container
	sorted           []docker.Container
	viewportStart    int
	showAll          bool
	loading          bool
	stopping         bool
	starting         bool
	deleting         bool
	confirming       bool
	confirmAction    string
	confirmID        string
	confirmName      string
	filtering        bool
	filterQuery      string
	err              error
	width            int
	height           int
	logsVisible      bool
	logsLines        []string
	logsContainer    string
	logsScrollOffset int
	logsAutoScroll   bool
	logsStop         func()

	inspectVisible   bool
	inspectLines     []string
	inspectContainer string
	inspectOffset    int

	copiedName string

	statsVisible     bool
	statsContainer   string
	statsContainerID string
	statsEntry       *docker.StatsEntry
}

func New() App {
	return App{
		loading: true,
		showAll: true,
		table:   buildTable(nil, 120),
	}
}

func (m App) Init() tea.Cmd {
	return docker.FetchContainers(m.showAll)
}

func (m App) filtered() []docker.Container {
	if m.filterQuery == "" {
		return m.sorted
	}
	q := strings.ToLower(m.filterQuery)
	var out []docker.Container
	for _, c := range m.sorted {
		if strings.Contains(strings.ToLower(c.Names), q) ||
			strings.Contains(strings.ToLower(c.Image), q) ||
			strings.Contains(strings.ToLower(c.ID), q) ||
			strings.Contains(strings.ToLower(c.ComposeProject()), q) ||
			strings.Contains(strings.ToLower(c.ComposeService()), q) {
			out = append(out, c)
		}
	}
	return out
}

func (m App) rebuildTable() App {
	m.table = buildTable(m.filtered(), m.width)
	m.table.SetHeight(m.tableHeight())
	m.viewportStart = 0
	return m
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.rebuildTable()
		return m, nil

	case tea.KeyMsg:
		m.copiedName = ""
		if m.logsVisible {
			switch msg.String() {
			case "esc", "l":
				m = m.closeLogs()
			case "q", "ctrl+c":
				m = m.closeLogs()
				return m, tea.Quit
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
		if m.inspectVisible {
			switch msg.String() {
			case "esc", "i":
				m = m.closeInspect()
			case "q", "ctrl+c":
				m = m.closeInspect()
				return m, tea.Quit
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
		if m.statsVisible {
			switch msg.String() {
			case "esc", "t":
				m = m.closeStats()
			case "q", "ctrl+c":
				m = m.closeStats()
				return m, tea.Quit
			case "r":
				m.loading = true
				m.err = nil
				m.statsEntry = nil
				return m, tea.Batch(docker.FetchContainers(m.showAll), docker.FetchStats(m.statsContainerID))
			}
			return m, nil
		}
		if m.confirming {
			switch msg.String() {
			case "y", "Y":
				m.confirming = false
				m.err = nil
				switch m.confirmAction {
				case "stop":
					m.stopping = true
					return m, docker.StopContainer(m.confirmID)
				case "start":
					m.starting = true
					return m, docker.StartContainer(m.confirmID)
				case "delete":
					m.deleting = true
					return m, docker.DeleteContainer(m.confirmID)
				}
			case "n", "N", "esc", "q":
				m.confirming = false
			}
			return m, nil
		}
		if m.filtering {
			switch msg.Type {
			case tea.KeyEsc, tea.KeyEnter:
				m.filtering = false
			case tea.KeyBackspace, tea.KeyDelete:
				if len(m.filterQuery) > 0 {
					runes := []rune(m.filterQuery)
					m.filterQuery = string(runes[:len(runes)-1])
					m = m.rebuildTable()
				}
			case tea.KeyRunes:
				m.filterQuery += string(msg.Runes)
				m = m.rebuildTable()
			}
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			m.err = nil
			return m, docker.FetchContainers(m.showAll)
		case "a":
			m.showAll = !m.showAll
			m.loading = true
			m.err = nil
			return m, docker.FetchContainers(m.showAll)
		case "/":
			m.filtering = true
			return m, nil
		case "esc":
			if m.filterQuery != "" {
				m.filterQuery = ""
				m = m.rebuildTable()
			}
		case "l":
			cursor := m.table.Cursor()
			filtered := m.filtered()
			if cursor >= 0 && cursor < len(filtered) {
				m.logsContainer = filtered[cursor].Names
				m.logsLines = nil
				m.logsScrollOffset = 0
				m.logsAutoScroll = true
				m.logsVisible = true
				firstLine, stop := docker.StartLogs(filtered[cursor].ID)
				m.logsStop = stop
				m.table.SetHeight(m.tableHeight())
				return m, firstLine
			}
		case "s":
			cursor := m.table.Cursor()
			filtered := m.filtered()
			if cursor >= 0 && cursor < len(filtered) && filtered[cursor].State == "running" {
				m.confirming = true
				m.confirmAction = "stop"
				m.confirmID = filtered[cursor].ID
				m.confirmName = filtered[cursor].Names
				return m, nil
			}
		case "S":
			cursor := m.table.Cursor()
			filtered := m.filtered()
			if cursor >= 0 && cursor < len(filtered) && filtered[cursor].State != "running" {
				m.confirming = true
				m.confirmAction = "start"
				m.confirmID = filtered[cursor].ID
				m.confirmName = filtered[cursor].Names
				return m, nil
			}
		case "d":
			cursor := m.table.Cursor()
			filtered := m.filtered()
			if cursor >= 0 && cursor < len(filtered) && filtered[cursor].State != "running" {
				m.confirming = true
				m.confirmAction = "delete"
				m.confirmID = filtered[cursor].ID
				m.confirmName = filtered[cursor].Names
				return m, nil
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
				return m, docker.InspectContainer(filtered[cursor].ID)
			}
		case "c":
			cursor := m.table.Cursor()
			filtered := m.filtered()
			if cursor >= 0 && cursor < len(filtered) {
				c := filtered[cursor]
				m.copiedName = c.Names
				return m, copyToClipboard(c.ID)
			}
		case "t":
			cursor := m.table.Cursor()
			filtered := m.filtered()
			if cursor >= 0 && cursor < len(filtered) && filtered[cursor].State == "running" {
				m.statsVisible = true
				m.statsEntry = nil
				m.statsContainer = filtered[cursor].Names
				m.statsContainerID = filtered[cursor].ID
				m.table.SetHeight(m.tableHeight())
				return m, docker.FetchStats(filtered[cursor].ID)
			}
		}

	case docker.ContainersMsg:
		m.containers = msg
		m.sorted = docker.Sort(m.containers)
		m.loading = false
		m.err = nil
		m = m.rebuildTable()
		return m, nil

	case docker.ErrMsg:
		m.err = msg.Err
		m.loading = false
		return m, nil

	case docker.StopMsg:
		m.stopping = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		return m, docker.FetchContainers(m.showAll)

	case docker.StartMsg:
		m.starting = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		return m, docker.FetchContainers(m.showAll)

	case docker.DeleteMsg:
		m.deleting = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		kept := m.containers[:0]
		for _, c := range m.containers {
			if c.ID != msg.ID {
				kept = append(kept, c)
			}
		}
		m.containers = kept
		m.sorted = docker.Sort(m.containers)
		m = m.rebuildTable()
		return m, nil

	case docker.LogsLineMsg:
		if !m.logsVisible {
			return m, nil
		}
		m.logsLines = append(m.logsLines, msg.Line)
		if m.logsAutoScroll {
			m.logsScrollOffset = max(0, len(m.logsLines)-(logsPanelHeight-2))
		}
		return m, msg.Next

	case docker.LogsEndMsg:
		return m, nil

	case docker.InspectMsg:
		if !m.inspectVisible {
			return m, nil
		}
		if msg.Err != nil {
			m.err = msg.Err
			m = m.closeInspect()
			return m, nil
		}
		m.inspectLines = buildInspectLines(msg.Data, m.width)
		return m, nil

	case docker.StatsMsg:
		if !m.statsVisible {
			return m, nil
		}
		if msg.Err != nil {
			m.err = msg.Err
			m = m.closeStats()
			return m, nil
		}
		m.statsEntry = &msg.Entry
		return m, nil
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

func (m App) closeLogs() App {
	if m.logsStop != nil {
		m.logsStop()
		m.logsStop = nil
	}
	m.logsVisible = false
	m.logsLines = nil
	m.logsContainer = ""
	m.logsScrollOffset = 0
	m.logsAutoScroll = true
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) closeInspect() App {
	m.inspectVisible = false
	m.inspectLines = nil
	m.inspectContainer = ""
	m.inspectOffset = 0
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) closeStats() App {
	m.statsVisible = false
	m.statsEntry = nil
	m.statsContainer = ""
	m.statsContainerID = ""
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) tableHeight() int {
	reserved := 8
	if m.logsVisible {
		reserved += logsPanelHeight
	}
	if m.inspectVisible {
		reserved += inspectPanelHeight
	}
	if m.statsVisible {
		reserved += statsPanelHeight
	}
	h := m.height - reserved
	if h < 3 {
		return 3
	}
	return h
}

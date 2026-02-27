package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pivovarit/tdocker/internal/docker"
)

const logsPanelHeight = 15 // 1 divider + 1 title + 13 log lines

type Model struct {
	table            table.Model
	containers       []docker.Container
	sorted           []docker.Container
	viewportStart    int
	showAll          bool
	loading          bool
	stopping         bool
	confirming       bool
	confirmIdx       int
	err              error
	width            int
	height           int
	logsVisible      bool
	logsLines        []string
	logsContainer    string
	logsScrollOffset int
	logsAutoScroll   bool
	logsStop         func()
}

func InitialModel() Model {
	return Model{
		loading: true,
		showAll: true,
		table:   buildTable(nil, 120),
	}
}

func (m Model) Init() tea.Cmd {
	return docker.FetchContainers(m.showAll)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table = buildTable(m.sorted, m.width)
		m.table.SetHeight(m.tableHeight())
		return m, nil

	case tea.KeyMsg:
		if m.logsVisible {
			switch msg.String() {
			case "esc", "l":
				m.closeLogs()
			case "q", "ctrl+c":
				m.closeLogs()
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
		if m.confirming {
			switch msg.String() {
			case "y", "Y":
				m.confirming = false
				m.stopping = true
				m.err = nil
				return m, docker.StopContainer(m.sorted[m.confirmIdx].ID)
			case "n", "N", "esc", "q":
				m.confirming = false
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
		case "l":
			cursor := m.table.Cursor()
			if cursor >= 0 && cursor < len(m.sorted) {
				m.logsContainer = m.sorted[cursor].Names
				m.logsLines = nil
				m.logsScrollOffset = 0
				m.logsAutoScroll = true
				m.logsVisible = true
				firstLine, stop := docker.StartLogs(m.sorted[cursor].ID)
				m.logsStop = stop
				m.table.SetHeight(m.tableHeight())
				return m, firstLine
			}
		case "s":
			cursor := m.table.Cursor()
			if cursor >= 0 && cursor < len(m.sorted) && m.sorted[cursor].State == "running" {
				m.confirming = true
				m.confirmIdx = cursor
				return m, nil
			}
		}

	case docker.ContainersMsg:
		m.containers = msg
		m.sorted = docker.Sort(m.containers)
		m.viewportStart = 0
		m.loading = false
		m.err = nil
		m.table = buildTable(m.sorted, m.width)
		m.table.SetHeight(m.tableHeight())
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

	case docker.LogsLineMsg:
		if !m.logsVisible {
			return m, nil // panel closed; discard stale messages
		}
		m.logsLines = append(m.logsLines, msg.Line)
		if m.logsAutoScroll {
			m.logsScrollOffset = max(0, len(m.logsLines)-(logsPanelHeight-2))
		}
		return m, msg.Next

	case docker.LogsEndMsg:
		// streaming ended (container exited or panel closed)
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

func (m *Model) closeLogs() {
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
}

func (m Model) tableHeight() int {
	reserved := 8
	if m.logsVisible {
		reserved += logsPanelHeight
	}
	h := m.height - reserved
	if h < 3 {
		return 3
	}
	return h
}

func (m Model) renderLogsPanel() string {
	var b strings.Builder
	w := m.width

	b.WriteString(logsDividerStyle.Render(strings.Repeat("─", w)))
	b.WriteString("\n")
	b.WriteString(logsTitleStyle.Render(" Logs: " + m.logsContainer))
	b.WriteString("\n")

	maxLines := logsPanelHeight - 2
	start := m.logsScrollOffset
	end := start + maxLines
	if end > len(m.logsLines) {
		end = len(m.logsLines)
	}

	shown := 0
	for i := start; i < end; i++ {
		b.WriteString(logsLineStyle.Render("  " + m.logsLines[i]))
		b.WriteString("\n")
		shown++
	}
	for ; shown < maxLines; shown++ {
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) View() string {
	var b strings.Builder

	mode := "running"
	if m.showAll {
		mode = "all"
	}

	b.WriteString(titleStyle.Render(
		fmt.Sprintf(" tdocker  ·  %d container(s)  ·  showing %s", len(m.containers), mode),
	))
	b.WriteString("\n")

	switch {
	case m.stopping:
		b.WriteString(emptyStyle.Render("Stopping container…"))
	case m.loading:
		b.WriteString(emptyStyle.Render("Fetching containers…"))

	case m.err != nil:
		b.WriteString(errorStyle.Render("  Error: " + m.err.Error()))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  Press " + keyStyle.Render("r") + " to retry, " + keyStyle.Render("q") + " to quit."))
		return b.String()

	case len(m.containers) == 0:
		msg := "No running containers."
		if m.showAll {
			msg = "No containers found."
		}
		b.WriteString(emptyStyle.Render(msg))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  Press " +
			keyStyle.Render("a") + " to toggle all containers, " +
			keyStyle.Render("r") + " to refresh, " +
			keyStyle.Render("q") + " to quit."))
		return b.String()

	default:
		const headerLines = 2

		lines := strings.Split(m.table.View(), "\n")
		cursor := m.table.Cursor()
		for i, line := range lines {
			dataIdx := i - headerLines
			if dataIdx < 0 {
				continue
			}
			containerIdx := m.viewportStart + dataIdx
			if containerIdx >= len(m.sorted) {
				break
			}
			if m.sorted[containerIdx].State != "running" && containerIdx != cursor {
				lines[i] = stoppedRowStyle.Render(line)
			}
		}
		b.WriteString(tableStyle.Render(strings.Join(lines, "\n")))
	}

	if m.logsVisible {
		b.WriteString("\n")
		b.WriteString(m.renderLogsPanel())
	}

	b.WriteString("\n")
	if m.logsVisible {
		b.WriteString(helpStyle.Render(
			"  ↑/↓ scroll  ·  " +
				keyStyle.Render("g") + " top  ·  " +
				keyStyle.Render("G") + " bottom  ·  " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("l") + " close  ·  " +
				keyStyle.Render("q") + " quit",
		))
	} else if m.confirming && m.confirmIdx < len(m.sorted) {
		name := m.sorted[m.confirmIdx].Names
		b.WriteString(
			confirmStyle.Render("  Stop ") +
				keyStyle.Render(name) +
				confirmStyle.Render("? press ") +
				keyStyle.Render("y") +
				confirmStyle.Render(" to confirm, ") +
				keyStyle.Render("n") +
				confirmStyle.Render(" to cancel"),
		)
	} else {
		b.WriteString(helpStyle.Render(
			"  ↑/↓ navigate  ·  " +
				keyStyle.Render("l") + " logs  ·  " +
				keyStyle.Render("s") + " stop  ·  " +
				keyStyle.Render("a") + " toggle all  ·  " +
				keyStyle.Render("r") + " refresh  ·  " +
				keyStyle.Render("q") + " quit",
		))
	}

	return b.String()
}

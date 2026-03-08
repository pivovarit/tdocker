package ui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type logsState struct {
	visible     bool
	lines       []string
	container   string
	containerID string
	scroll      scrollState
	allMode     bool
	gen         int
	cancel      context.CancelFunc
	searching   bool
	searchQuery string
	timestamps  bool
}

func (m App) closeLogs() App {
	if m.logs.cancel != nil {
		m.logs.cancel()
	}
	m.logs = logsState{scroll: scrollState{autoScroll: true}}
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) restartLogs() (tea.Model, tea.Cmd) {
	if m.logs.cancel != nil {
		m.logs.cancel()
	}
	m.logs.lines = nil
	m.logs.searchQuery = ""
	m.logs.scroll = scrollState{autoScroll: true}
	m.logs.gen++
	ctx, cancel := context.WithCancel(context.Background())
	m.logs.cancel = cancel
	tail := logsTailDefault
	if m.logs.allMode {
		tail = "all"
	}
	return m, m.client.StartLogs(ctx, m.logs.containerID, tail, m.logs.timestamps, m.logs.gen)
}

func (m App) logsFiltered() []string {
	if m.logs.searchQuery == "" {
		return m.logs.lines
	}
	q := strings.ToLower(m.logs.searchQuery)
	var out []string
	for _, line := range m.logs.lines {
		if strings.Contains(strings.ToLower(line), q) {
			out = append(out, line)
		}
	}
	return out
}

func (m App) handleLogsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.logs.searching {
		return m.handleLogsSearchKey(msg)
	}
	lines := m.logsFiltered()
	switch msg.Code {
	case tea.KeyEsc:
		if m.logs.searchQuery != "" {
			m.logs.searchQuery = ""
			m.logs.scroll = scrollState{autoScroll: true}
			if m.logs.scroll.autoScroll {
				m.logs.scroll.offset = max(0, len(m.logs.lines)-(m.logsPanelHeight()-2))
			}
			return m, nil
		}
		m = m.closeLogs()
	case 'l':
		m = m.closeLogs()
	case '/':
		m.logs.searching = true
	case 'f':
		m.logs.allMode = !m.logs.allMode
		return m.restartLogs()
	case 't':
		if msg.Text == "T" {
			m.logs.timestamps = !m.logs.timestamps
			return m.restartLogs()
		}
	case tea.KeyUp, 'k':
		m.logs.scroll = m.logs.scroll.up()
	case tea.KeyDown, 'j':
		m.logs.scroll = m.logs.scroll.down(len(lines), m.logsPanelHeight()-2)
	case 'g', tea.KeyHome:
		if msg.Text == keyScrollBottom {
			m.logs.scroll = m.logs.scroll.bottom(len(lines), m.logsPanelHeight()-2)
		} else {
			m.logs.scroll = m.logs.scroll.top()
		}
	case tea.KeyEnd:
		m.logs.scroll = m.logs.scroll.bottom(len(lines), m.logsPanelHeight()-2)
	}
	return m, nil
}

func (m App) handleLogsSearchKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.Code {
	case tea.KeyEsc:
		m.logs.searching = false
		m.logs.searchQuery = ""
		m.logs.scroll = scrollState{autoScroll: true}
		if m.logs.scroll.autoScroll {
			m.logs.scroll.offset = max(0, len(m.logs.lines)-(m.logsPanelHeight()-2))
		}
	case tea.KeyEnter:
		m.logs.searching = false
		if m.logs.searchQuery != "" {
			m.logs.scroll = scrollState{}
		}
	case tea.KeyBackspace:
		if len(m.logs.searchQuery) > 0 {
			m.logs.searchQuery = m.logs.searchQuery[:len(m.logs.searchQuery)-1]
		}
		m.logs.scroll = scrollState{}
	default:
		if msg.Text != "" {
			m.logs.searchQuery += msg.Text
			m.logs.scroll = scrollState{}
		}
	}
	return m, nil
}

func (m App) renderLogsPanel() string {
	logsModeLabel := " (last 200)"
	if m.logs.allMode {
		logsModeLabel = " (all)"
	}
	if m.logs.timestamps {
		logsModeLabel += " (timestamps)"
	}
	searchLabel := ""
	if m.logs.searchQuery != "" || m.logs.searching {
		searchLabel = " [/" + m.logs.searchQuery
		if m.logs.searching {
			searchLabel += "▌"
		}
		searchLabel += "]"
	}
	lines := m.logsFiltered()
	return m.renderPanel(" Logs: "+m.logs.container+logsModeLabel+searchLabel, func(b *strings.Builder) {
		maxLines := m.logsPanelHeight() - 2
		start := m.logs.scroll.offset
		end := start + maxLines
		if end > len(lines) {
			end = len(lines)
		}
		for _, line := range lines[start:end] {
			if m.logs.timestamps {
				if ts, rest, ok := strings.Cut(line, " "); ok {
					b.WriteString(logsTimestampStyle.Render("  "+ts) + " " + logsLineStyle.Render(rest))
				} else {
					b.WriteString(logsLineStyle.Render("  " + line))
				}
			} else {
				b.WriteString(logsLineStyle.Render("  " + line))
			}
			b.WriteString("\n")
		}
		panelPad(b, end-start, maxLines)
	})
}

package ui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type logsState struct {
	visible        bool
	lines          []string
	container      string
	containerID    string
	scroll         scrollState
	allMode        bool
	gen            int
	cancel         context.CancelFunc
	searching      bool
	searchQuery    string
	timestamps     bool
	grepMode       bool
	isCompose      bool
	composeProject string
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
	m.logs.scroll = scrollState{autoScroll: true}
	m.logs.gen++
	ctx, cancel := context.WithCancel(context.Background())
	m.logs.cancel = cancel
	tail := logsTailDefault
	if m.logs.allMode {
		tail = "all"
	}
	if m.logs.isCompose {
		return m, m.client.StartComposeLogs(ctx, m.logs.composeProject, tail, m.logs.timestamps, m.logs.gen)
	}
	grep := ""
	if m.logs.grepMode {
		grep = m.logs.searchQuery
	}
	return m, m.client.StartLogs(ctx, m.logs.containerID, tail, m.logs.timestamps, grep, m.logs.gen)
}

func (m App) logsFiltered() []string {
	if m.logs.searchQuery == "" || m.logs.grepMode {
		return m.logs.lines
	}
	q := strings.ToLower(m.logs.searchQuery)
	out := make([]string, 0, len(m.logs.lines))
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
	if msg.String() == "ctrl+g" && m.grepSupported && m.logs.searchQuery != "" && !m.logs.isCompose {
		m.logs.grepMode = !m.logs.grepMode
		return m.restartLogs()
	}
	lines := m.logsFiltered()
	switch msg.Code {
	case tea.KeyEsc:
		if m.logs.searchQuery != "" {
			wasGrep := m.logs.grepMode
			m.logs.searchQuery = ""
			m.logs.grepMode = false
			m.logs.scroll = scrollState{autoScroll: true}
			if wasGrep {
				return m.restartLogs()
			}
			m.logs.scroll.offset = max(0, len(m.logs.lines)-(m.logsPanelHeight()-2))
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
	default:
		m.logs.scroll, _ = m.logs.scroll.handleKey(msg, len(lines), m.logsPanelHeight()-2)
	}
	return m, nil
}

func (m App) confirmLogsSearch() App {
	m.logs.searching = false
	if m.logs.searchQuery != "" {
		m.logs.scroll = scrollState{}
	}
	return m
}

func (m App) handleLogsSearchKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.Code {
	case tea.KeyEsc:
		m.logs.searching = false
		m.logs.searchQuery = ""
		m.logs.scroll = scrollState{autoScroll: true}
		m.logs.scroll.offset = max(0, len(m.logs.lines)-(m.logsPanelHeight()-2))
	case tea.KeyEnter:
		m = m.confirmLogsSearch()
	case tea.KeyBackspace:
		if len(m.logs.searchQuery) > 0 {
			m.logs.searchQuery = trimLastRune(m.logs.searchQuery)
		}
		m.logs.scroll = scrollState{}
	case tea.KeyUp, tea.KeyDown, tea.KeyHome, tea.KeyEnd:
		m = m.confirmLogsSearch()
		return m.handleLogsKey(msg)
	default:
		if msg.Text != "" {
			m.logs.searchQuery += msg.Text
			m.logs.scroll = scrollState{}
		}
	}
	return m, nil
}

func highlightMatches(s, query string) string {
	if query == "" {
		return logsLineStyle.Render(s)
	}
	lower := strings.ToLower(s)
	q := strings.ToLower(query)
	var b strings.Builder
	pos := 0
	for {
		idx := strings.Index(lower[pos:], q)
		if idx < 0 {
			b.WriteString(logsLineStyle.Render(s[pos:]))
			return b.String()
		}
		abs := pos + idx
		if abs > pos {
			b.WriteString(logsLineStyle.Render(s[pos:abs]))
		}
		b.WriteString(logsHighlightStyle.Render(s[abs : abs+len(q)]))
		pos = abs + len(q)
	}
}

func (m App) renderLogsPanel() string {
	logsModeLabel := " (last 200)"
	if m.logs.allMode {
		logsModeLabel = " (all)"
	}
	if m.logs.timestamps {
		logsModeLabel += " (timestamps)"
	}
	composeLabel := ""
	if m.logs.isCompose {
		composeLabel = " (compose)"
	}
	searchLabel := ""
	if m.logs.searchQuery != "" || m.logs.searching {
		prefix := " [/"
		if m.logs.grepMode {
			prefix = " [grep: "
		}
		searchLabel = prefix + m.logs.searchQuery
		if m.logs.searching {
			searchLabel += "▌"
		}
		searchLabel += "]"
	}
	lines := m.logsFiltered()
	query := m.logs.searchQuery
	return m.renderPanel(" Logs: "+m.logs.container+composeLabel+logsModeLabel+searchLabel, func(b *strings.Builder) {
		maxLines := m.logsPanelHeight() - 2
		start := m.logs.scroll.offset
		end := start + maxLines
		if end > len(lines) {
			end = len(lines)
		}
		for _, line := range lines[start:end] {
			if m.logs.timestamps {
				if ts, rest, ok := strings.Cut(line, " "); ok {
					b.WriteString(logsTimestampStyle.Render("  "+ts) + " " + highlightMatches(rest, query))
				} else {
					b.WriteString(highlightMatches("  "+line, query))
				}
			} else {
				b.WriteString(highlightMatches("  "+line, query))
			}
			b.WriteString("\n")
		}
		panelPad(b, end-start, maxLines)
	})
}

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
}

func (m App) closeLogs() App {
	if m.logs.cancel != nil {
		m.logs.cancel()
	}
	m.logs = logsState{scroll: scrollState{autoScroll: true}}
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) handleLogsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.Code {
	case tea.KeyEsc, 'l':
		m = m.closeLogs()
	case 'f':
		if m.logs.cancel != nil {
			m.logs.cancel()
		}
		m.logs.allMode = !m.logs.allMode
		m.logs.lines = nil
		m.logs.scroll = scrollState{autoScroll: true}
		m.logs.gen++
		ctx, cancel := context.WithCancel(context.Background())
		m.logs.cancel = cancel
		tail := logsTailDefault
		if m.logs.allMode {
			tail = "all"
		}
		return m, m.client.StartLogs(ctx, m.logs.containerID, tail, m.logs.gen)
	case tea.KeyUp, 'k':
		m.logs.scroll = m.logs.scroll.up()
	case tea.KeyDown, 'j':
		m.logs.scroll = m.logs.scroll.down(len(m.logs.lines), m.logsPanelHeight()-2)
	case 'g', tea.KeyHome:
		if msg.Text == keyScrollBottom {
			m.logs.scroll = m.logs.scroll.bottom(len(m.logs.lines), m.logsPanelHeight()-2)
		} else {
			m.logs.scroll = m.logs.scroll.top()
		}
	case tea.KeyEnd:
		m.logs.scroll = m.logs.scroll.bottom(len(m.logs.lines), m.logsPanelHeight()-2)
	}
	return m, nil
}

func (m App) renderLogsPanel() string {
	logsModeLabel := " (last 200)"
	if m.logs.allMode {
		logsModeLabel = " (all)"
	}
	return m.renderPanel(" Logs: "+m.logs.container+logsModeLabel, func(b *strings.Builder) {
		maxLines := m.logsPanelHeight() - 2
		start := m.logs.scroll.offset
		end := start + maxLines
		if end > len(m.logs.lines) {
			end = len(m.logs.lines)
		}
		for _, line := range m.logs.lines[start:end] {
			b.WriteString(logsLineStyle.Render("  " + line))
			b.WriteString("\n")
		}
		panelPad(b, end-start, maxLines)
	})
}

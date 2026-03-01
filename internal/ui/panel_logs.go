package ui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

func (m App) handleLogsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "l":
		m = m.closeLogs()
	case "f":
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
	case "up", "k":
		m.logs.scroll = m.logs.scroll.up()
	case "down", "j":
		m.logs.scroll = m.logs.scroll.down(len(m.logs.lines), logsPanelHeight-2)
	case "g", "home":
		m.logs.scroll = m.logs.scroll.top()
	case "G", "end":
		m.logs.scroll = m.logs.scroll.bottom(len(m.logs.lines), logsPanelHeight-2)
	}
	return m, nil
}

func (m App) renderLogsPanel() string {
	logsModeLabel := " (last 200)"
	if m.logs.allMode {
		logsModeLabel = " (all)"
	}
	return m.renderPanel(" Logs: "+m.logs.container+logsModeLabel, func(b *strings.Builder) {
		maxLines := logsPanelHeight - 2
		start := m.logs.scroll.offset
		end := start + maxLines
		if end > len(m.logs.lines) {
			end = len(m.logs.lines)
		}
		shown := 0
		for i := start; i < end; i++ {
			b.WriteString(logsLineStyle.Render("  " + m.logs.lines[i]))
			b.WriteString("\n")
			shown++
		}
		panelPad(b, shown, maxLines)
	})
}

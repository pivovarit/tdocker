package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pivovarit/tdocker/internal/docker"
)

type inspectState struct {
	visible   bool
	lines     []string
	container string
	scroll    scrollState
}

func (m App) closeInspect() App {
	m.inspect = inspectState{}
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) handleInspectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "i":
		m = m.closeInspect()
	case "up", "k":
		m.inspect.scroll = m.inspect.scroll.up()
	case "down", "j":
		m.inspect.scroll = m.inspect.scroll.down(len(m.inspect.lines), inspectPanelHeight-2)
	case "g", "home":
		m.inspect.scroll = m.inspect.scroll.top()
	case "G", "end":
		m.inspect.scroll = m.inspect.scroll.bottom(len(m.inspect.lines), inspectPanelHeight-2)
	}
	return m, nil
}

func (m App) renderInspectPanel() string {
	return m.renderPanel(" Inspect: "+m.inspect.container, func(b *strings.Builder) {
		maxLines := inspectPanelHeight - 2
		if len(m.inspect.lines) == 0 {
			b.WriteString(emptyStyle.Render("Loading…"))
			b.WriteString("\n")
			panelPad(b, 1, maxLines)
			return
		}
		start := m.inspect.scroll.offset
		end := start + maxLines
		if end > len(m.inspect.lines) {
			end = len(m.inspect.lines)
		}
		for _, line := range m.inspect.lines[start:end] {
			b.WriteString(line)
			b.WriteString("\n")
		}
		panelPad(b, end-start, maxLines)
	})
}

func buildInspectLines(d *docker.InspectData, width int) []string {
	var lines []string
	for _, l := range d.Lines(width) {
		switch l.Kind {
		case docker.InspectLineSection:
			lines = append(lines, inspectSectionStyle.Render(l.Key))
		case docker.InspectLineKeyValue:
			lines = append(lines, "  "+keyStyle.Render(l.Key)+"  "+inspectValueStyle.Render(l.Value))
		case docker.InspectLineValue:
			lines = append(lines, "  "+inspectValueStyle.Render(l.Value))
		case docker.InspectLineBlank:
			lines = append(lines, "")
		}
	}
	return lines
}

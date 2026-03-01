package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

type ctxPickerState struct {
	visible   bool
	requested bool
	contexts  []docker.DockerContext
	cursor    int
	current   string
}

func (m App) handleContextKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.Code {
	case tea.KeyEsc:
		m.ctxPicker.visible = false
		m.ctxPicker.contexts = nil
		m.ctxPicker.cursor = 0
	case tea.KeyUp, 'k':
		if m.ctxPicker.cursor > 0 {
			m.ctxPicker.cursor--
		}
	case tea.KeyDown, 'j':
		if m.ctxPicker.cursor < len(m.ctxPicker.contexts)-1 {
			m.ctxPicker.cursor++
		}
	case tea.KeyEnter:
		if len(m.ctxPicker.contexts) > 0 {
			return m, m.client.SwitchContext(m.ctxPicker.contexts[m.ctxPicker.cursor].Name)
		}
	}
	return m, nil
}

func (m App) renderContextPicker() string {
	return m.renderPanel(" Docker Contexts", func(b *strings.Builder) {
		if len(m.ctxPicker.contexts) == 0 {
			b.WriteString(emptyStyle.Render("No contexts found."))
			b.WriteString("\n")
			return
		}
		for i, c := range m.ctxPicker.contexts {
			name := c.Name
			label := "  " + name
			if c.Description != "" {
				label += "  " + c.Description
			}
			if c.Current {
				label = "* " + name
				if c.Description != "" {
					label += "  " + c.Description
				}
			}
			switch {
			case i == m.ctxPicker.cursor && c.Current:
				b.WriteString(contextCursorStyle.Render("* " + name + "  ✓"))
			case i == m.ctxPicker.cursor:
				b.WriteString(contextCursorStyle.Render(label))
			case c.Current:
				b.WriteString(contextActiveStyle.Render(label))
			default:
				b.WriteString(logsLineStyle.Render(label))
			}
			b.WriteString("\n")
		}
	})
}

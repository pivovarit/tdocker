package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

type ctxPickerState struct {
	visible       bool
	requested     bool
	contexts      []docker.Context
	cursor        int
	current       string
	viewportStart int
}

func (m App) handleContextKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	maxRows := min(len(m.ctxPicker.contexts), ctxPanelMaxRows)
	switch msg.Code {
	case tea.KeyEsc:
		m.ctxPicker.visible = false
		m.ctxPicker.contexts = nil
		m.ctxPicker.cursor = 0
		m.ctxPicker.viewportStart = 0
	case tea.KeyUp, 'k':
		if m.ctxPicker.cursor > 0 {
			m.ctxPicker.cursor--
			if m.ctxPicker.cursor < m.ctxPicker.viewportStart {
				m.ctxPicker.viewportStart--
			}
		}
	case tea.KeyDown, 'j':
		if m.ctxPicker.cursor < len(m.ctxPicker.contexts)-1 {
			m.ctxPicker.cursor++
			if m.ctxPicker.cursor >= m.ctxPicker.viewportStart+maxRows {
				m.ctxPicker.viewportStart++
			}
		}
	case tea.KeyEnter:
		if len(m.ctxPicker.contexts) > 0 {
			return m, m.client.SwitchContext(m.ctxPicker.contexts[m.ctxPicker.cursor].Name)
		}
	}
	return m, nil
}

func (m App) renderContextPicker() string {
	var b strings.Builder

	if len(m.ctxPicker.contexts) == 0 {
		b.WriteString(emptyStyle.MarginLeft(0).Render("No contexts found."))
		b.WriteString("\n")
	} else {
		maxRows := min(len(m.ctxPicker.contexts), ctxPanelMaxRows)
		start := m.ctxPicker.viewportStart
		end := min(start+maxRows, len(m.ctxPicker.contexts))

		labels := make([]string, end-start)
		maxWidth := 0
		for i := start; i < end; i++ {
			c := m.ctxPicker.contexts[i]
			label := "  " + c.Name
			if c.Description != "" {
				label += "  " + c.Description
			}
			if c.Current {
				label = "* " + c.Name
				if c.Description != "" {
					label += "  " + c.Description
				}
				if i == m.ctxPicker.cursor {
					label += "  ✓"
				}
			}
			labels[i-start] = label
			if w := len([]rune(label)); w > maxWidth {
				maxWidth = w
			}
		}

		for i := start; i < end; i++ {
			label := labels[i-start]
			c := m.ctxPicker.contexts[i]
			padded := label + strings.Repeat(" ", maxWidth-len([]rune(label)))

			var line string
			switch {
			case i == m.ctxPicker.cursor:
				line = contextCursorStyle.Render(padded)
			case c.Current:
				line = contextActiveStyle.Render(label)
			default:
				line = logsLineStyle.Render(label)
			}
			b.WriteString(line)
			if i < end-1 {
				b.WriteString("\n")
			}
		}
	}

	popup := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#0369A1")).
		Padding(1, 2).
		Render(b.String())

	title := " Docker Contexts "
	styledTitle := logsTitleStyle.Render(title)
	borderColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#0369A1"))

	popupLines := strings.Split(popup, "\n")
	if len(popupLines) > 0 {
		plain := ansiRe.ReplaceAllString(popupLines[0], "")
		runes := []rune(plain)
		titleStart := 2
		titleRunes := []rune(title)
		if titleStart+len(titleRunes) < len(runes) {
			popupLines[0] = borderColor.Render(string(runes[:titleStart])) +
				styledTitle +
				borderColor.Render(string(runes[titleStart+len(titleRunes):]))
		}
		popup = strings.Join(popupLines, "\n")
	}

	return popup
}

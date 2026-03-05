package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

func (m App) renderHelpOverlay() string {
	type entry struct{ key, desc string }
	type section struct {
		title   string
		entries []entry
	}

	sections := []section{
		{"Navigation", []entry{
			{"↑ / k", "move up"},
			{"↓ / j", "move down"},
			{"g", "jump to top"},
			{"G", "jump to bottom"},
		}},
		{"Container", []entry{
			{"s", "start"},
			{"S", "stop"},
			{"R", "restart"},
			{"D", "delete (stopped only)"},
			{"e", "exec shell"},
			{"x", "docker debug"},
			{"c", "copy ID"},
		}},
		{"Panels", []entry{
			{"l", "logs"},
			{"i", "inspect"},
			{"t", "stats"},
			{"v", "events"},
		}},
		{"App", []entry{
			{"r", "refresh"},
			{"A", "toggle all / running"},
			{"/", "filter"},
			{"X", "switch context"},
			{"?", "close this help"},
			{"q / ctrl+c", "quit"},
		}},
	}

	var columns []string
	for _, s := range sections {
		var b strings.Builder
		b.WriteString(inspectSectionStyle.Render(s.title) + "\n")
		for _, e := range s.entries {
			b.WriteString("  " + keyStyle.Render(e.key) + helpStyle.MarginTop(0).Render("  "+e.desc) + "\n")
		}
		columns = append(columns, b.String())
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, columns...)

	var b strings.Builder
	b.WriteString(logsDividerStyle.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")
	b.WriteString("  " + row)
	return b.String()
}

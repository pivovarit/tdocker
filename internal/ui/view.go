package ui

import (
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

func errorHintFor(err error) string {
	if errors.Is(err, docker.ErrDaemonUnavailable) {
		return "Is Docker running? Check that the daemon is started and your socket path is correct."
	}
	return ""
}

func (m App) View() tea.View {
	var b strings.Builder
	mode := "running"
	if m.showAll {
		mode = "all"
	}

	filtered := m.filtered()

	leftPlain := " tdocker  ·  " + mode + " [A]  ·  / filter  ·  r refresh"
	if m.filterQuery != "" {
		leftPlain += ": " + fmt.Sprintf("%q", m.filterQuery)
	}
	rightPlain := ""
	if m.ctxPicker.current != "" {
		rightPlain = "ctx [X]: " + m.ctxPicker.current + " "
	}

	pad := 2
	if rightPlain != "" && m.width > 0 {
		if p := m.width - len([]rune(leftPlain)) - len([]rune(rightPlain)); p > pad {
			pad = p
		}
	}

	sep := titleHintStyle.Render("  ·  ")
	styledLeft := titleStyle.Render(" tdocker") + sep +
		titleStyle.Render(mode) + titleHintStyle.Render(" [A]") + sep +
		titleHintStyle.Render("/ filter") + sep +
		titleHintStyle.Render("r refresh")
	if m.filterQuery != "" {
		styledLeft += titleHintStyle.Render(": ") + titleStyle.Render(fmt.Sprintf("%q", m.filterQuery))
	}
	styledRight := ""
	if m.ctxPicker.current != "" {
		styledRight = titleHintStyle.Render("ctx [X]: ") + titleStyle.Render(m.ctxPicker.current) + " "
	}

	b.WriteString(styledLeft + strings.Repeat(" ", pad) + styledRight)
	b.WriteString("\n\n")

	switch {
	case m.op == OpStopping:
		b.WriteString(emptyStyle.Render("Stopping container…"))
	case m.op == OpStarting:
		b.WriteString(emptyStyle.Render("Starting container…"))
	case m.op == OpRestarting:
		b.WriteString(emptyStyle.Render("Restarting container…"))
	case m.op == OpDeleting:
		b.WriteString(emptyStyle.Render("Deleting container…"))
	case m.loading:
		b.WriteString(emptyStyle.Render("Fetching containers…"))

	case m.err != nil:
		b.WriteString(errorStyle.Render("  Error: " + m.err.Error()))
		b.WriteString("\n")
		if hint := errorHintFor(m.err); hint != "" {
			b.WriteString(helpStyle.Render("  " + hint))
			b.WriteString("\n")
		}
		b.WriteString(helpStyle.Render("  Press " + keyStyle.Render("r") + " to retry, " + keyStyle.Render("q") + " to quit."))
		v := tea.NewView(b.String())
		v.AltScreen = true
		return v

	case len(m.containers) == 0:
		msg := "No running containers."
		if m.showAll {
			msg = "No containers found."
		}
		b.WriteString(emptyStyle.Render(msg))
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  Press " +
			keyStyle.Render("A") + " to toggle all containers, " +
			keyStyle.Render("r") + " to refresh, " +
			keyStyle.Render("q") + " to quit."))
		v := tea.NewView(b.String())
		v.AltScreen = true
		return v

	case len(filtered) == 0:
		b.WriteString(emptyStyle.Render(fmt.Sprintf("No containers match %q.", m.filterQuery)))

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
			if containerIdx >= len(filtered) {
				break
			}
			if containerIdx != cursor {
				switch filtered[containerIdx].State {
				case "paused":
					lines[i] = pausedRowStyle.Render(line)
				case "running":
				default:
					lines[i] = stoppedRowStyle.Render(line)
				}
			}
		}
		b.WriteString(tableStyle.Render(strings.Join(lines, "\n")))
	}

	if m.logs.visible {
		b.WriteString("\n")
		b.WriteString(m.renderLogsPanel())
	}

	if m.inspect.visible {
		b.WriteString("\n")
		b.WriteString(m.renderInspectPanel())
	}

	if m.stats.visible {
		b.WriteString("\n")
		b.WriteString(m.renderStatsPanel())
	}

	if m.ctxPicker.visible {
		b.WriteString("\n")
		b.WriteString(m.renderContextPicker())
	}

	if m.events.visible {
		b.WriteString("\n")
		b.WriteString(m.renderEventsPanel())
	}

	b.WriteString("\n")
	b.WriteString(m.helpBar())

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func (m App) helpBar() string {
	switch {
	case m.events.visible:
		return helpStyle.Render(
			"  ↑/↓ scroll · " +
				keyStyle.Render("g") + " top · " +
				keyStyle.Render("G") + " bottom · " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("v") + " close · " +
				keyStyle.Render("q") + " quit",
		)
	case m.logs.visible:
		return helpStyle.Render(
			"  ↑/↓ scroll · " +
				keyStyle.Render("g") + " top · " +
				keyStyle.Render("G") + " bottom · " +
				keyStyle.Render("f") + " toggle all · " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("l") + " close · " +
				keyStyle.Render("q") + " quit",
		)
	case m.inspect.visible:
		return helpStyle.Render(
			"  ↑/↓ scroll · " +
				keyStyle.Render("g") + " top · " +
				keyStyle.Render("G") + " bottom · " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("i") + " close · " +
				keyStyle.Render("q") + " quit",
		)
	case m.stats.visible:
		return helpStyle.Render(
			"  " + keyStyle.Render("r") + " refresh · " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("t") + " close · " +
				keyStyle.Render("q") + " quit",
		)
	case m.op == OpConfirming:
		verb := "Stop"
		switch m.confirmAction {
		case "start":
			verb = "Start"
		case "restart":
			verb = "Restart"
		case "delete":
			verb = "Delete"
		}
		return confirmStyle.Render("  "+verb+" ") +
			confirmNameStyle.Render(m.confirmName) +
			confirmStyle.Render("? press ") +
			keyStyle.Render("y") +
			confirmStyle.Render(" to confirm, ") +
			keyStyle.Render("n") +
			confirmStyle.Render(" to cancel")
	case m.ctxPicker.visible:
		return helpStyle.Render(
			"  ↑/↓/j/k navigate · " +
				keyStyle.Render("enter") + " switch · " +
				keyStyle.Render("esc") + " cancel",
		)
	case m.filtering:
		return helpStyle.Render(
			"  / " + keyStyle.Render(m.filterQuery+"▌") + " · esc/enter exit",
		)
	default:
		if m.warnMsg != "" {
			return helpStyle.Render("  ") + eventWarnStyle.Render("⚠ "+m.warnMsg)
		}
		if m.copiedName != "" {
			return helpStyle.Render(
				"  " + confirmStyle.Render("✓ copied ID of ") + keyStyle.Render(m.copiedName),
			)
		}
		prefix := ""
		if m.filterQuery != "" {
			prefix = keyStyle.Render("["+m.filterQuery+"]") + " · " + keyStyle.Render("esc") + " clear · "
		}
		return helpStyle.Render(
			"  " + prefix +
				keyStyle.Render("l") + " logs · " +
				keyStyle.Render("i") + " inspect · " +
				keyStyle.Render("e") + " exec · " +
				keyStyle.Render("S") + " stop · " +
				keyStyle.Render("s") + " start · " +
				keyStyle.Render("R") + " restart · " +
				keyStyle.Render("D") + " delete · " +
				keyStyle.Render("t") + " stats · " +
				keyStyle.Render("v") + " events · " +
				keyStyle.Render("c") + " copy id · " +
				keyStyle.Render("x") + " debug",
		)
	}
}

func (m App) renderPanel(title string, body func(*strings.Builder)) string {
	var b strings.Builder
	b.WriteString(logsDividerStyle.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")
	b.WriteString(logsTitleStyle.Render(title))
	b.WriteString("\n")
	body(&b)
	return b.String()
}

func panelPad(b *strings.Builder, shown, maxLines int) {
	for ; shown < maxLines; shown++ {
		b.WriteString("\n")
	}
}

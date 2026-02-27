package ui

import (
	"fmt"
	"strings"
)

func (m Model) View() string {
	var b strings.Builder

	mode := "running"
	if m.showAll {
		mode = "all"
	}

	filtered := m.filtered()
	countStr := fmt.Sprintf("%d", len(m.containers))
	if m.filterQuery != "" {
		countStr = fmt.Sprintf("%d/%d", len(filtered), len(m.containers))
	}

	b.WriteString(titleStyle.Render(
		fmt.Sprintf(" tdocker  ·  %s container(s)  ·  showing %s", countStr, mode),
	))
	b.WriteString("\n")

	switch {
	case m.stopping:
		b.WriteString(emptyStyle.Render("Stopping container…"))
	case m.starting:
		b.WriteString(emptyStyle.Render("Starting container…"))
	case m.deleting:
		b.WriteString(emptyStyle.Render("Deleting container…"))
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
			if filtered[containerIdx].State != "running" && containerIdx != cursor {
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
	switch {
	case m.logsVisible:
		b.WriteString(helpStyle.Render(
			"  ↑/↓ scroll  ·  " +
				keyStyle.Render("g") + " top  ·  " +
				keyStyle.Render("G") + " bottom  ·  " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("l") + " close  ·  " +
				keyStyle.Render("q") + " quit",
		))
	case m.confirming:
		verb := "Stop"
		switch m.confirmAction {
		case "start":
			verb = "Start"
		case "delete":
			verb = "Delete"
		}
		b.WriteString(
			confirmStyle.Render("  "+verb+" ") +
				keyStyle.Render(m.confirmName) +
				confirmStyle.Render("? press ") +
				keyStyle.Render("y") +
				confirmStyle.Render(" to confirm, ") +
				keyStyle.Render("n") +
				confirmStyle.Render(" to cancel"),
		)
	case m.filtering:
		b.WriteString(helpStyle.Render(
			"  / " + keyStyle.Render(m.filterQuery+"▌") + "   ·  esc/enter exit",
		))
	default:
		prefix := ""
		if m.filterQuery != "" {
			prefix = keyStyle.Render("["+m.filterQuery+"]") + "  ·  " + keyStyle.Render("esc") + " clear  ·  "
		}
		b.WriteString(helpStyle.Render(
			"  " + prefix + "↑/↓ navigate  ·  " +
				keyStyle.Render("/") + " filter  ·  " +
				keyStyle.Render("l") + " logs  ·  " +
				keyStyle.Render("s") + " stop  ·  " +
				keyStyle.Render("S") + " start  ·  " +
				keyStyle.Render("d") + " delete  ·  " +
				keyStyle.Render("a") + " toggle all  ·  " +
				keyStyle.Render("r") + " refresh  ·  " +
				keyStyle.Render("q") + " quit",
		))
	}

	return b.String()
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

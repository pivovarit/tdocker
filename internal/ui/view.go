package ui

import (
	"fmt"
	"strings"

	"github.com/pivovarit/tdocker/internal/docker"
)

func errorHintFor(err error) string {
	msg := err.Error()
	if strings.Contains(msg, "Cannot connect to the Docker daemon") ||
		strings.Contains(msg, "Is the docker daemon running") ||
		strings.Contains(msg, "connection refused") ||
		(strings.Contains(msg, "no such file or directory") && strings.Contains(msg, "docker.sock")) {
		return "Is Docker running? Check that the daemon is started and your socket path is correct."
	}
	return ""
}

func (m App) View() string {
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

	title := fmt.Sprintf(" tdocker  ·  %s container(s)  ·  showing %s", countStr, mode)
	if m.filterQuery != "" {
		title += fmt.Sprintf("  ·  filter: %q", m.filterQuery)
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	switch {
	case m.op == OpStopping:
		b.WriteString(emptyStyle.Render("Stopping container…"))
	case m.op == OpStarting:
		b.WriteString(emptyStyle.Render("Starting container…"))
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

	if m.logsVisible {
		b.WriteString("\n")
		b.WriteString(m.renderLogsPanel())
	}

	if m.inspectVisible {
		b.WriteString("\n")
		b.WriteString(m.renderInspectPanel())
	}

	if m.statsVisible {
		b.WriteString("\n")
		b.WriteString(m.renderStatsPanel())
	}

	b.WriteString("\n")
	switch {
	case m.logsVisible:
		b.WriteString(helpStyle.Render(
			"  ↑/↓ scroll · " +
				keyStyle.Render("g") + " top · " +
				keyStyle.Render("G") + " bottom · " +
				keyStyle.Render("f") + " toggle all · " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("l") + " close · " +
				keyStyle.Render("q") + " quit",
		))
	case m.inspectVisible:
		b.WriteString(helpStyle.Render(
			"  ↑/↓ scroll · " +
				keyStyle.Render("g") + " top · " +
				keyStyle.Render("G") + " bottom · " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("i") + " close · " +
				keyStyle.Render("q") + " quit",
		))
	case m.statsVisible:
		b.WriteString(helpStyle.Render(
			"  " + keyStyle.Render("r") + " refresh · " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("t") + " close · " +
				keyStyle.Render("q") + " quit",
		))
	case m.op == OpConfirming:
		verb := "Stop"
		switch m.confirmAction {
		case "start":
			verb = "Start"
		case "delete":
			verb = "Delete"
		}
		b.WriteString(
			confirmStyle.Render("  "+verb+" ") +
				confirmNameStyle.Render(m.confirmName) +
				confirmStyle.Render("? press ") +
				keyStyle.Render("y") +
				confirmStyle.Render(" to confirm, ") +
				keyStyle.Render("n") +
				confirmStyle.Render(" to cancel"),
		)
	case m.filtering:
		b.WriteString(helpStyle.Render(
			"  / " + keyStyle.Render(m.filterQuery+"▌") + " · esc/enter exit",
		))
	default:
		if m.copiedName != "" {
			b.WriteString(helpStyle.Render(
				"  " + confirmStyle.Render("✓ copied ID of ") + keyStyle.Render(m.copiedName),
			))
		} else {
			prefix := ""
			if m.filterQuery != "" {
				prefix = keyStyle.Render("["+m.filterQuery+"]") + " · " + keyStyle.Render("esc") + " clear · "
			}
			b.WriteString(helpStyle.Render(
				"  " + prefix + "↑/↓ navigate · " +
					keyStyle.Render("/") + " filter · " +
					keyStyle.Render("c") + " copy · " +
					keyStyle.Render("e") + " exec · " +
					keyStyle.Render("x") + " debug · " +
					keyStyle.Render("i") + " inspect · " +
					keyStyle.Render("t") + " stats · " +
					keyStyle.Render("l") + " logs · " +
					keyStyle.Render("s") + " stop · " +
					keyStyle.Render("S") + " start · " +
					keyStyle.Render("d") + " delete · " +
					keyStyle.Render("a") + " toggle all · " +
					keyStyle.Render("r") + " refresh · " +
					keyStyle.Render("q") + " quit",
			))
		}
	}

	return b.String()
}

func (m App) renderLogsPanel() string {
	var b strings.Builder
	w := m.width

	b.WriteString(logsDividerStyle.Render(strings.Repeat("─", w)))
	b.WriteString("\n")
	logsModeLabel := " (last 200)"
	if m.logsAllMode {
		logsModeLabel = " (all)"
	}
	b.WriteString(logsTitleStyle.Render(" Logs: " + m.logsContainer + logsModeLabel))
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

func (m App) renderInspectPanel() string {
	var b strings.Builder
	w := m.width

	b.WriteString(logsDividerStyle.Render(strings.Repeat("─", w)))
	b.WriteString("\n")
	b.WriteString(logsTitleStyle.Render(" Inspect: " + m.inspectContainer))
	b.WriteString("\n")

	maxLines := inspectPanelHeight - 2

	if len(m.inspectLines) == 0 {
		b.WriteString(emptyStyle.Render("Loading…"))
		b.WriteString("\n")
		for i := 1; i < maxLines; i++ {
			b.WriteString("\n")
		}
		return b.String()
	}

	start := m.inspectOffset
	end := start + maxLines
	if end > len(m.inspectLines) {
		end = len(m.inspectLines)
	}

	shown := 0
	for i := start; i < end; i++ {
		b.WriteString(m.inspectLines[i])
		b.WriteString("\n")
		shown++
	}
	for ; shown < maxLines; shown++ {
		b.WriteString("\n")
	}

	return b.String()
}

func (m App) renderStatsPanel() string {
	var b strings.Builder
	w := m.width

	b.WriteString(logsDividerStyle.Render(strings.Repeat("─", w)))
	b.WriteString("\n")
	b.WriteString(logsTitleStyle.Render(" Stats: " + m.statsContainer))
	b.WriteString("\n")

	if m.statsEntry == nil {
		b.WriteString(emptyStyle.Render("Loading…"))
		b.WriteString("\n")
		for i := 1; i < statsPanelHeight-2; i++ {
			b.WriteString("\n")
		}
		return b.String()
	}

	e := m.statsEntry
	row := func(label, value string) string {
		return "  " + inspectSectionStyle.Render(fmt.Sprintf("%-10s", label)) + "  " + inspectValueStyle.Render(value) + "\n"
	}

	b.WriteString("\n")
	b.WriteString(row("CPU", e.CPUPerc))
	b.WriteString(row("Memory", e.MemUsage+"  ("+e.MemPerc+")"))
	b.WriteString(row("Net I/O", e.NetIO))
	b.WriteString(row("Block I/O", e.BlockIO))
	b.WriteString(row("PIDs", e.PIDs))

	return b.String()
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

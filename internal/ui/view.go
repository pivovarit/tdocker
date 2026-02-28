package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pivovarit/tdocker/internal/docker"
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
			"  ↑/↓ scroll  ·  " +
				keyStyle.Render("g") + " top  ·  " +
				keyStyle.Render("G") + " bottom  ·  " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("l") + " close  ·  " +
				keyStyle.Render("q") + " quit",
		))
	case m.inspectVisible:
		b.WriteString(helpStyle.Render(
			"  ↑/↓ scroll  ·  " +
				keyStyle.Render("g") + " top  ·  " +
				keyStyle.Render("G") + " bottom  ·  " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("i") + " close  ·  " +
				keyStyle.Render("q") + " quit",
		))
	case m.statsVisible:
		b.WriteString(helpStyle.Render(
			"  " + keyStyle.Render("r") + " refresh  ·  " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("t") + " close  ·  " +
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
				keyStyle.Render("i") + " inspect  ·  " +
				keyStyle.Render("t") + " stats  ·  " +
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

func (m Model) renderInspectPanel() string {
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

func (m Model) renderStatsPanel() string {
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
	add := func(s string) { lines = append(lines, s) }

	add(inspectSectionStyle.Render("Image"))
	digest := d.ImageDigest
	if width > 4 && len(digest) > width-4 {
		digest = digest[:width-5] + "…"
	}
	add("  " + inspectValueStyle.Render(digest))
	add("")

	add(inspectSectionStyle.Render("Ports"))
	if len(d.Ports) == 0 {
		add("  " + inspectValueStyle.Render("(none)"))
	} else {
		portKeys := make([]string, 0, len(d.Ports))
		for k := range d.Ports {
			portKeys = append(portKeys, k)
		}
		sort.Strings(portKeys)
		for _, containerPort := range portKeys {
			bindings := d.Ports[containerPort]
			if len(bindings) == 0 {
				add("  " + keyStyle.Render(containerPort) + "  " + inspectValueStyle.Render("→  (not published)"))
			} else {
				for _, b := range bindings {
					host := b.HostIP + ":" + b.HostPort
					add("  " + keyStyle.Render(containerPort) + "  " + inspectValueStyle.Render("→  "+host))
				}
			}
		}
	}
	add("")

	add(inspectSectionStyle.Render("Environment"))
	if len(d.Env) == 0 {
		add("  " + inspectValueStyle.Render("(none)"))
	} else {
		for _, e := range d.Env {
			if idx := strings.Index(e, "="); idx > 0 {
				add("  " + keyStyle.Render(e[:idx]+"=") + inspectValueStyle.Render(e[idx+1:]))
			} else {
				add("  " + inspectValueStyle.Render(e))
			}
		}
	}
	add("")

	add(inspectSectionStyle.Render("Mounts"))
	if len(d.Mounts) == 0 {
		add("  " + inspectValueStyle.Render("(none)"))
	} else {
		for _, mount := range d.Mounts {
			rw := "ro"
			if mount.RW {
				rw = "rw"
			}
			src := mount.Source
			if src == "" {
				src = "(" + mount.Type + ")"
			}
			add("  " + keyStyle.Render(src) + "  " + inspectValueStyle.Render("→  "+mount.Destination+"  ("+rw+")"))
		}
	}

	return lines
}

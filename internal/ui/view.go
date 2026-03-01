package ui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/pivovarit/tdocker/internal/docker"
)

func errorHintFor(err error) string {
	if errors.Is(err, docker.ErrDaemonUnavailable) {
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
		return b.String()

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
	switch {
	case m.events.visible:
		b.WriteString(helpStyle.Render(
			"  ↑/↓ scroll · " +
				keyStyle.Render("g") + " top · " +
				keyStyle.Render("G") + " bottom · " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("v") + " close · " +
				keyStyle.Render("q") + " quit",
		))
	case m.logs.visible:
		b.WriteString(helpStyle.Render(
			"  ↑/↓ scroll · " +
				keyStyle.Render("g") + " top · " +
				keyStyle.Render("G") + " bottom · " +
				keyStyle.Render("f") + " toggle all · " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("l") + " close · " +
				keyStyle.Render("q") + " quit",
		))
	case m.inspect.visible:
		b.WriteString(helpStyle.Render(
			"  ↑/↓ scroll · " +
				keyStyle.Render("g") + " top · " +
				keyStyle.Render("G") + " bottom · " +
				keyStyle.Render("esc") + "/" + keyStyle.Render("i") + " close · " +
				keyStyle.Render("q") + " quit",
		))
	case m.stats.visible:
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
		case "restart":
			verb = "Restart"
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
	case m.ctxPicker.visible:
		b.WriteString(helpStyle.Render(
			"  ↑/↓/j/k navigate · " +
				keyStyle.Render("enter") + " switch · " +
				keyStyle.Render("esc") + " cancel",
		))
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
			))
		}
	}

	return b.String()
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
		shown := 0
		for i := start; i < end; i++ {
			b.WriteString(m.inspect.lines[i])
			b.WriteString("\n")
			shown++
		}
		panelPad(b, shown, maxLines)
	})
}

func (m App) renderStatsPanel() string {
	return m.renderPanel(" Stats: "+m.stats.container, func(b *strings.Builder) {
		maxLines := statsPanelHeight - 2
		if m.stats.entry == nil {
			b.WriteString(emptyStyle.Render("Loading…"))
			b.WriteString("\n")
			panelPad(b, 1, maxLines)
			return
		}
		e := m.stats.entry
		p := m.stats.prevEntry

		cpuTrend, memTrend, netTrend, blkTrend, pidTrend := "", "", "", "", ""
		if p != nil {
			cpuTrend = statsTrend(p.CPUPerc, e.CPUPerc, parsePercent)
			memTrend = statsTrend(p.MemPerc, e.MemPerc, parsePercent)
			netTrend = statsTrend(p.NetIO, e.NetIO, parseSizeFirst)
			blkTrend = statsTrend(p.BlockIO, e.BlockIO, parseSizeFirst)
			pidTrend = statsTrend(p.PIDs, e.PIDs, parseNumber)
		}

		row := func(label, value, trend string) {
			b.WriteString("  " + inspectSectionStyle.Render(fmt.Sprintf("%-10s", label)) + "  " + inspectValueStyle.Render(value) + trend + "\n")
		}

		b.WriteString("\n")
		row("CPU", e.CPUPerc, cpuTrend)
		row("Memory", e.MemUsage+"  ("+e.MemPerc+")", memTrend)
		row("Net I/O", e.NetIO, netTrend)
		row("Block I/O", e.BlockIO, blkTrend)
		row("PIDs", e.PIDs, pidTrend)
	})
}

func parsePercent(s string) (float64, bool) {
	s = strings.TrimSuffix(strings.TrimSpace(s), "%")
	v, err := strconv.ParseFloat(s, 64)
	return v, err == nil
}

func parseByteSize(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	i := 0
	for i < len(s) && (s[i] == '.' || (s[i] >= '0' && s[i] <= '9')) {
		i++
	}
	if i == 0 {
		return 0, false
	}
	num, err := strconv.ParseFloat(s[:i], 64)
	if err != nil {
		return 0, false
	}
	switch strings.TrimSpace(s[i:]) {
	case "B":
		return num, true
	case "kB":
		return num * 1e3, true
	case "MB":
		return num * 1e6, true
	case "GB":
		return num * 1e9, true
	case "TB":
		return num * 1e12, true
	case "KiB":
		return num * 1024, true
	case "MiB":
		return num * 1024 * 1024, true
	case "GiB":
		return num * 1024 * 1024 * 1024, true
	case "TiB":
		return num * 1024 * 1024 * 1024 * 1024, true
	default:
		return num, true
	}
}

func parseSizeFirst(s string) (float64, bool) {
	if idx := strings.Index(s, " / "); idx != -1 {
		s = s[:idx]
	}
	return parseByteSize(strings.TrimSpace(s))
}

func parseNumber(s string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v, err == nil
}

const (
	// trendRelThreshold is the minimum relative change (1%) required to
	// show an up/down trend arrow instead of a steady dot.
	trendRelThreshold = 0.01
	// trendAbsMinimum prevents noise on near-zero values where the relative
	// threshold would be smaller than measurement precision.
	trendAbsMinimum = 0.001
)

func statsTrend(prev, curr string, parse func(string) (float64, bool)) string {
	p, ok1 := parse(prev)
	c, ok2 := parse(curr)
	if !ok1 || !ok2 {
		return ""
	}
	th := p * trendRelThreshold
	if th < trendAbsMinimum {
		th = trendAbsMinimum
	}
	d := c - p
	if d > th {
		return " " + trendUpStyle.Render("↑")
	}
	if d < -th {
		return " " + trendDownStyle.Render("↓")
	}
	return " " + trendSteadyStyle.Render("·")
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

func (m App) renderEventsPanel() string {
	return m.renderPanel(" Events  (live)", func(b *strings.Builder) {
		maxLines := eventsPanelHeight - 2
		if len(m.events.events) == 0 {
			b.WriteString(emptyStyle.Render("Waiting for events…"))
			b.WriteString("\n")
			panelPad(b, 1, maxLines)
			return
		}
		start := m.events.scroll.offset
		end := start + maxLines
		if end > len(m.events.events) {
			end = len(m.events.events)
		}
		shown := 0
		for i := start; i < end; i++ {
			ev := m.events.events[i]
			actionStyle := eventDimStyle
			switch ev.Action {
			case "start", "unpause", "create":
				actionStyle = eventStartStyle
			case "die", "stop", "kill", "destroy", "oom":
				actionStyle = eventStopStyle
			case "pause":
				actionStyle = eventWarnStyle
			}
			line := eventTimeStyle.Render(ev.Timestamp()) + "  " +
				eventTypeStyle.Render(fmt.Sprintf("%-11s", ev.Type)) +
				actionStyle.Render(fmt.Sprintf("%-12s", ev.Action)) +
				eventNameStyle.Render(ev.Name())
			b.WriteString("  " + line + "\n")
			shown++
		}
		panelPad(b, shown, maxLines)
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

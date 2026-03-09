package ui

import (
	"errors"
	"fmt"
	"strings"
	"time"

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
	const ctxPrefix = "ctx [X]: "
	const ctxSuffix = " "
	const minPad = 2

	updatePlain := ""
	if m.updateAvailable != "" {
		updatePlain = "new version available"
	}

	ctxName := m.ctxPicker.current
	if ctxName != "" && m.width > 0 {
		updateW := 0
		if updatePlain != "" {
			updateW = len([]rune(updatePlain)) + 5
		}
		maxNameW := m.width - len([]rune(leftPlain)) - len([]rune(ctxPrefix)) - len([]rune(ctxSuffix)) - updateW - minPad
		if maxNameW < 1 {
			maxNameW = 1
		}
		ctxName = trunc(ctxName, maxNameW)
	}

	rightPlain := ""
	switch {
	case updatePlain != "" && ctxName != "":
		rightPlain = updatePlain + "  ·  " + ctxPrefix + ctxName + ctxSuffix
	case updatePlain != "":
		rightPlain = updatePlain + " "
	case ctxName != "":
		rightPlain = ctxPrefix + ctxName + ctxSuffix
	}

	pad := minPad
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
	switch {
	case updatePlain != "" && ctxName != "":
		styledRight = titleHintStyle.Render(updatePlain) + sep + titleHintStyle.Render(ctxPrefix) + titleStyle.Render(ctxName) + ctxSuffix
	case updatePlain != "":
		styledRight = titleHintStyle.Render(updatePlain) + " "
	case ctxName != "":
		styledRight = titleHintStyle.Render(ctxPrefix) + titleStyle.Render(ctxName) + ctxSuffix
	}

	b.WriteString(styledLeft + strings.Repeat(" ", pad) + styledRight)
	b.WriteString("\n\n")

	switch {
	case m.helpVisible:
		b.WriteString(m.renderHelpOverlay())

	case m.fetch.visible && len(m.containers) == 0:
		elapsed := time.Since(m.fetch.start)
		loadingMsg := "Fetching containers…"
		if elapsed >= time.Second {
			loadingMsg += fmt.Sprintf(" (%ds)", int(elapsed.Seconds()))
		}
		b.WriteString(emptyStyle.Render(loadingMsg))
		if m.fetch.slow {
			b.WriteString("\n")
			b.WriteString(helpStyle.Render("  Docker is taking a long time to respond. Press " +
				keyStyle.Render("q") + " to quit or keep waiting."))
		}

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
		if m.logs.visible {
			b.WriteString(m.renderLogsPanel())
		} else if m.inspect.visible {
			b.WriteString(m.renderInspectPanel())
		} else {
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
					case "collapsed":
						lines[i] = collapsedRowStyle.Render(line)
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

var opLabels = map[Operation]string{
	OpStopping:   "Stopping container…",
	OpStarting:   "Starting container…",
	OpRestarting: "Restarting container…",
	OpDeleting:   "Deleting container…",
	OpRenaming:   "Renaming container…",
}

var confirmVerbs = map[string]string{
	"stop": "Stop", "start": "Start", "restart": "Restart", "delete": "Delete",
	"compose-stop": "Stop", "compose-start": "Start", "compose-restart": "Restart",
}

func (m App) helpBar() string {
	switch {
	case m.helpVisible:
		return helpBarHelp()
	case m.op.visible:
		if label, ok := opLabels[m.op.kind]; ok {
			return confirmStyle.Render("  " + label)
		}
	case m.events.visible:
		return helpBarEvents()
	case m.logs.searching:
		return helpBarLogsSearch(m.logs.searchQuery)
	case m.logs.visible:
		return helpBarLogs(m.logs.searchQuery, m.logs.grepMode, m.grepSupported)
	case m.inspect.visible:
		return helpBarInspect()
	case m.stats.visible:
		return helpBarStats()
	case m.op.kind == OpConfirming:
		return helpBarConfirm(m.op.action, m.op.name)
	case m.ctxPicker.visible:
		return helpBarCtxPicker()
	case m.rename.active:
		return helpBarRename(m.rename.input)
	case m.filtering:
		return helpBarFilter(m.filterQuery)
	case m.isCollapsedSelected():
		c, _ := m.selectedContainer()
		return helpBarCollapsed(m.projectHasRunning(c.ComposeProject()))
	default:
		canCollapse := false
		if c, ok := m.selectedContainer(); ok {
			canCollapse = c.ComposeProject() != ""
		}
		return helpBarDefault(m.warnMsg, m.copiedName, m.filterQuery, canCollapse)
	}
	return ""
}

func helpBarHelp() string {
	return helpStyle.Render("  " + keyStyle.Render("?") + "/" + keyStyle.Render("esc") + "/" + keyStyle.Render("q") + " close")
}

func helpBarEvents() string {
	return helpStyle.Render(
		"  ↑/↓ scroll · " +
			keyStyle.Render("g") + " top · " +
			keyStyle.Render("G") + " bottom · " +
			keyStyle.Render("esc") + "/" + keyStyle.Render("v") + " close · " +
			keyStyle.Render("q") + " close",
	)
}

func helpBarLogsSearch(query string) string {
	return helpStyle.Render(
		"  / " + keyStyle.Render(query+"▌") + " · esc cancel · enter confirm",
	)
}

func helpBarLogs(searchQuery string, grepMode, grepSupported bool) string {
	searchHint := keyStyle.Render("/") + " search · "
	if searchQuery != "" {
		grepHint := ""
		if grepSupported {
			if grepMode {
				grepHint = keyStyle.Render("ctrl+g") + " client filter · "
			} else {
				grepHint = keyStyle.Render("ctrl+g") + " server grep · "
			}
		}
		searchHint = keyStyle.Render("["+searchQuery+"]") + " · " + grepHint + keyStyle.Render("esc") + " clear · "
	}
	return helpStyle.Render(
		"  ↑/↓ scroll · " +
			keyStyle.Render("g") + " top · " +
			keyStyle.Render("G") + " bottom · " +
			searchHint +
			keyStyle.Render("f") + " toggle all · " +
			keyStyle.Render("T") + " timestamps · " +
			keyStyle.Render("l") + " close · " +
			keyStyle.Render("q") + " close",
	)
}

func helpBarInspect() string {
	return helpStyle.Render(
		"  ↑/↓ scroll · " +
			keyStyle.Render("g") + " top · " +
			keyStyle.Render("G") + " bottom · " +
			keyStyle.Render("esc") + "/" + keyStyle.Render("i") + " close · " +
			keyStyle.Render("q") + " close",
	)
}

func helpBarStats() string {
	return helpStyle.Render(
		"  " + keyStyle.Render("r") + " refresh · " +
			keyStyle.Render("esc") + "/" + keyStyle.Render("t") + " close · " +
			keyStyle.Render("q") + " close",
	)
}

func helpBarConfirm(action, name string) string {
	verb := confirmVerbs[action]
	if verb == "" {
		verb = "Stop"
	}
	return confirmStyle.Render("  "+verb+" ") +
		confirmNameStyle.Render(name) +
		confirmStyle.Render("? press ") +
		keyStyle.Render("y") +
		confirmStyle.Render(" to confirm, ") +
		keyStyle.Render("n") +
		confirmStyle.Render(" to cancel")
}

func helpBarCtxPicker() string {
	return helpStyle.Render(
		"  ↑/↓/j/k navigate · " +
			keyStyle.Render("enter") + " switch · " +
			keyStyle.Render("esc") + " cancel",
	)
}

func helpBarRename(input string) string {
	return helpStyle.Render(
		"  rename: " + keyStyle.Render(input+"▌") + " · " +
			keyStyle.Render("enter") + " confirm · " +
			keyStyle.Render("esc") + " cancel",
	)
}

func helpBarFilter(query string) string {
	return helpStyle.Render(
		"  / " + keyStyle.Render(query+"▌") + " · esc/enter exit",
	)
}

func (m App) isCollapsedSelected() bool {
	c, ok := m.selectedContainer()
	return ok && c.State == "collapsed"
}

func helpBarCollapsed(hasRunning bool) string {
	stopLabel := " start/stop · "
	restartLabel := " restart · "
	if !hasRunning {
		stopLabel = " start · "
		restartLabel = " start · "
	}
	return helpStyle.Render(
		"  " + keyStyle.Render("→") + " expand · " +
			keyStyle.Render("S") + stopLabel +
			keyStyle.Render("R") + restartLabel +
			keyStyle.Render("?") + " help · " +
			keyStyle.Render("q") + " quit",
	)
}

func helpBarDefault(warnMsg, copiedName, filterQuery string, canCollapse bool) string {
	if warnMsg != "" {
		return helpStyle.Render("  ") + eventWarnStyle.Render("⚠ "+warnMsg)
	}
	if copiedName != "" {
		return helpStyle.Render(
			"  " + confirmStyle.Render("✓ copied ID of ") + keyStyle.Render(copiedName),
		)
	}
	prefix := ""
	if filterQuery != "" {
		prefix = keyStyle.Render("["+filterQuery+"]") + " · " + keyStyle.Render("esc") + " clear · "
	}
	collapseHint := ""
	if canCollapse {
		collapseHint = keyStyle.Render("←") + " collapse · "
	}
	return helpStyle.Render(
		"  " + prefix +
			collapseHint +
			keyStyle.Render("l") + " logs · " +
			keyStyle.Render("i") + " inspect · " +
			keyStyle.Render("e") + " exec · " +
			keyStyle.Render("S") + " start/stop · " +
			keyStyle.Render("R") + " restart · " +
			keyStyle.Render("P") + " pause · " +
			keyStyle.Render("N") + " rename · " +
			keyStyle.Render("D") + " delete · " +
			keyStyle.Render("t") + " stats · " +
			keyStyle.Render("v") + " events · " +
			keyStyle.Render("c") + " copy id · " +
			keyStyle.Render("x") + " debug",
	)
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

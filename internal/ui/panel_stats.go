package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

type statsState struct {
	visible     bool
	container   string
	containerID string
	entry       *docker.StatsEntry
	prevEntry   *docker.StatsEntry

	cpuHistory [sparklineMaxSamples]float64
	memHistory [sparklineMaxSamples]float64
	historyPos int
	historyLen int
}

func (m App) closeStats() App {
	m.stats = statsState{}
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) handleStatsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.Code {
	case tea.KeyEsc, 't':
		m = m.closeStats()
	}
	return m, nil
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

		cpuSpark := ""
		memSpark := ""
		if m.stats.historyLen > 0 {
			cpuSpark = "  " + sparklineStyle.Render(sparkline(m.stats.cpuHistory, m.stats.historyPos, m.stats.historyLen))
			memSpark = "  " + sparklineStyle.Render(sparkline(m.stats.memHistory, m.stats.historyPos, m.stats.historyLen))
		}

		row := func(label, value, trend string) {
			b.WriteString("  " + inspectSectionStyle.Render(fmt.Sprintf("%-10s", label)) + "  " + inspectValueStyle.Render(value) + trend + "\n")
		}

		b.WriteString("\n")
		row("CPU", e.CPUPerc, cpuTrend+cpuSpark)
		row("Memory", e.MemUsage+"  ("+e.MemPerc+")", memTrend+memSpark)
		row("Net I/O", e.NetIO, netTrend)
		row("Block I/O", e.BlockIO, blkTrend)
		row("PIDs", e.PIDs, pidTrend)
		panelPad(b, 1+statsRows, maxLines)
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
		return 0, false
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
	trendRelThreshold   = 0.01
	trendAbsMinimum     = 0.001
	sparklineMaxSamples = 20
)

var sparklineBlocks = []rune("▁▂▃▄▅▆▇█")

func sparkline(history [sparklineMaxSamples]float64, pos, n int) string {
	if n == 0 {
		return ""
	}
	start := (pos - n + sparklineMaxSamples) % sparklineMaxSamples
	mn, mx := history[start], history[start]
	for i := 1; i < n; i++ {
		v := history[(start+i)%sparklineMaxSamples]
		if v < mn {
			mn = v
		}
		if v > mx {
			mx = v
		}
	}
	var b strings.Builder
	for i := 0; i < n; i++ {
		v := history[(start+i)%sparklineMaxSamples]
		idx := 0
		if mx > mn {
			idx = int((v - mn) / (mx - mn) * 7)
		}
		b.WriteRune(sparklineBlocks[idx])
	}
	return b.String()
}

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

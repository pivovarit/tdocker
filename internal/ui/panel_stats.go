package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pivovarit/tdocker/internal/docker"
)

type statsState struct {
	visible     bool
	container   string
	containerID string
	entry       *docker.StatsEntry
	prevEntry   *docker.StatsEntry
	fetching    bool
}

func (m App) closeStats() App {
	m.stats = statsState{}
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) handleStatsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "t":
		m = m.closeStats()
	case "r":
		if m.stats.fetching {
			return m, nil
		}
		m.loading = true
		m.err = nil
		if m.stats.entry != nil {
			m.stats.prevEntry = m.stats.entry
		}
		m.stats.entry = nil
		m.stats.fetching = true
		return m, tea.Batch(m.client.FetchContainers(m.showAll), m.client.FetchStats(m.stats.containerID))
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

		row := func(label, value, trend string) {
			b.WriteString("  " + inspectSectionStyle.Render(fmt.Sprintf("%-10s", label)) + "  " + inspectValueStyle.Render(value) + trend + "\n")
		}

		b.WriteString("\n")
		row("CPU", e.CPUPerc, cpuTrend)
		row("Memory", e.MemUsage+"  ("+e.MemPerc+")", memTrend)
		row("Net I/O", e.NetIO, netTrend)
		row("Block I/O", e.BlockIO, blkTrend)
		row("PIDs", e.PIDs, pidTrend)
		panelPad(b, 1+statsRows, maxLines)
	})
}

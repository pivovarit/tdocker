package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

func TestUpdate_TKeyOnRunningOpensStatsPanel(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	got := update(m, runeKey("t"))
	if !got.stats.visible {
		t.Fatal("want stats.visible=true")
	}
	if got.stats.container != runningContainer.Names {
		t.Errorf("want stats.container=%q, got %q", runningContainer.Names, got.stats.container)
	}
	if got.stats.containerID != runningContainer.ID {
		t.Errorf("want stats.containerID=%q, got %q", runningContainer.ID, got.stats.containerID)
	}
}

func TestUpdate_TKeyOnRunningWithInlineStatsPopulatesEntry(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.inlineStats[runningContainer.ID] = docker.StatsEntry{CPUPerc: "2.50%"}
	got := update(m, runeKey("t"))
	if got.stats.entry == nil {
		t.Fatal("want stats.entry populated from inlineStats")
	}
	if got.stats.entry.CPUPerc != "2.50%" {
		t.Errorf("want CPUPerc=%q, got %q", "2.50%", got.stats.entry.CPUPerc)
	}
}

func TestUpdate_TKeyOnRunningWithoutInlineStatsNilEntry(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	got := update(m, runeKey("t"))
	if got.stats.entry != nil {
		t.Error("want stats.entry=nil when no inline stats available")
	}
}

func TestUpdate_TKeyOnStoppedDoesNothing(t *testing.T) {
	m := modelWithSorted([]docker.Container{stoppedContainer})
	got := update(m, runeKey("t"))
	if got.stats.visible {
		t.Error("want stats.visible=false for non-running container")
	}
}

func TestUpdate_TKeyOnEmptyListDoesNothing(t *testing.T) {
	m := modelWithSorted(nil)
	got := update(m, runeKey("t"))
	if got.stats.visible {
		t.Error("want stats.visible=false for empty list")
	}
}

func TestUpdate_StatsEscClosesPanel(t *testing.T) {
	m := statsPanel()
	got := update(m, tea.KeyPressMsg{Code: tea.KeyEsc})
	if got.stats.visible {
		t.Error("want stats.visible=false after esc")
	}
}

func TestUpdate_StatsTClosesPanel(t *testing.T) {
	m := statsPanel()
	got := update(m, runeKey("t"))
	if got.stats.visible {
		t.Error("want stats.visible=false after t (toggle)")
	}
}

func TestUpdate_StatsCloseResetsState(t *testing.T) {
	m := statsPanel()
	entry := docker.StatsEntry{CPUPerc: "1.00%"}
	m.stats.entry = &entry
	got := update(m, tea.KeyPressMsg{Code: tea.KeyEsc})
	if got.stats.entry != nil {
		t.Error("want stats.entry=nil after close")
	}
	if got.stats.container != "" {
		t.Error("want stats.container empty after close")
	}
	if got.stats.containerID != "" {
		t.Error("want stats.containerID empty after close")
	}
}

func TestUpdate_StatsOtherKeysIgnored(t *testing.T) {
	for _, key := range []tea.Msg{runeKey("a"), runeKey("s")} {
		m := statsPanel()
		got := update(m, key)
		if !got.stats.visible {
			t.Errorf("key %v: want stats.visible=true (panel should stay open)", key)
		}
	}
}

func TestUpdate_FlushUpdatesStatsPanelEntry(t *testing.T) {
	m := statsPanel()
	m.inlineStats[runningContainer.ID] = docker.StatsEntry{
		CPUPerc:  "0.42%",
		MemUsage: "3.4MiB / 1.9GiB",
		MemPerc:  "1.2%",
		NetIO:    "1.2kB / 456B",
		BlockIO:  "0B / 0B",
		PIDs:     "4",
	}
	m.statsDirty = true
	got := update(m, inlineStatsFlushMsg{})
	if got.stats.entry == nil {
		t.Fatal("want stats.entry set after flush")
	}
	if got.stats.entry.CPUPerc != "0.42%" {
		t.Errorf("want CPUPerc=%q, got %q", "0.42%", got.stats.entry.CPUPerc)
	}
	if got.stats.entry.MemUsage != "3.4MiB / 1.9GiB" {
		t.Errorf("want MemUsage=%q, got %q", "3.4MiB / 1.9GiB", got.stats.entry.MemUsage)
	}
	if got.stats.entry.PIDs != "4" {
		t.Errorf("want PIDs=%q, got %q", "4", got.stats.entry.PIDs)
	}
}

func TestUpdate_FlushSetsPrevEntry(t *testing.T) {
	m := statsPanel()
	oldEntry := docker.StatsEntry{CPUPerc: "1.00%"}
	m.stats.entry = &oldEntry
	m.inlineStats[runningContainer.ID] = docker.StatsEntry{CPUPerc: "2.00%"}
	m.statsDirty = true
	got := update(m, inlineStatsFlushMsg{})
	if got.stats.prevEntry == nil {
		t.Fatal("want prevEntry set")
	}
	if got.stats.prevEntry.CPUPerc != "1.00%" {
		t.Errorf("want prevEntry.CPUPerc=%q, got %q", "1.00%", got.stats.prevEntry.CPUPerc)
	}
}

func TestUpdate_FlushNoopWhenPanelClosed(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.inlineStats[runningContainer.ID] = docker.StatsEntry{CPUPerc: "1.00%"}
	m.statsDirty = true
	got := update(m, inlineStatsFlushMsg{})
	if got.stats.entry != nil {
		t.Error("want stats.entry=nil when panel not open")
	}
}

func statsPanel() App {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.stats.visible = true
	m.stats.container = runningContainer.Names
	m.stats.containerID = runningContainer.ID
	return m
}

func TestFlushPopulatesHistoryBuffers(t *testing.T) {
	m := statsPanel()
	m.inlineStats[runningContainer.ID] = docker.StatsEntry{
		CPUPerc: "10.00%",
		MemPerc: "20.00%",
	}
	m.statsDirty = true
	got := update(m, inlineStatsFlushMsg{})
	if got.stats.historyLen != 1 {
		t.Errorf("want historyLen=1, got %d", got.stats.historyLen)
	}
	if got.stats.cpuHistory[0] != 10.0 {
		t.Errorf("want cpuHistory[0]=10.0, got %f", got.stats.cpuHistory[0])
	}
	if got.stats.memHistory[0] != 20.0 {
		t.Errorf("want memHistory[0]=20.0, got %f", got.stats.memHistory[0])
	}
	if got.stats.historyPos != 1 {
		t.Errorf("want historyPos=1, got %d", got.stats.historyPos)
	}
}

func TestFlushWrapsHistoryAt20(t *testing.T) {
	m := statsPanel()
	m.stats.historyLen = 20
	m.stats.historyPos = 0
	m.stats.cpuHistory[0] = 1.0
	m.inlineStats[runningContainer.ID] = docker.StatsEntry{
		CPUPerc: "50.00%",
		MemPerc: "30.00%",
	}
	m.statsDirty = true
	got := update(m, inlineStatsFlushMsg{})
	if got.stats.historyLen != 20 {
		t.Errorf("want historyLen=20 (capped), got %d", got.stats.historyLen)
	}
	if got.stats.cpuHistory[0] != 50.0 {
		t.Errorf("want cpuHistory[0]=50.0 (overwritten), got %f", got.stats.cpuHistory[0])
	}
	if got.stats.historyPos != 1 {
		t.Errorf("want historyPos=1, got %d", got.stats.historyPos)
	}
}

func TestSparklineEmpty(t *testing.T) {
	var h [20]float64
	if got := sparkline(h, 0, 0); got != "" {
		t.Errorf("want empty string for n=0, got %q", got)
	}
}

func TestSparklineAllSameValues(t *testing.T) {
	var h [20]float64
	h[0] = 50.0
	h[1] = 50.0
	h[2] = 50.0
	got := sparkline(h, 3, 3)
	want := "▁▁▁"
	if got != want {
		t.Errorf("want %q for equal values, got %q", want, got)
	}
}

func TestSparklineMinMaxBars(t *testing.T) {
	var h [20]float64
	h[0] = 0.0
	h[1] = 100.0
	got := sparkline(h, 2, 2)
	want := "▁█"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestSparklineSingleSample(t *testing.T) {
	var h [20]float64
	h[5] = 99.0
	got := sparkline(h, 6, 1)
	want := "▁"
	if got != want {
		t.Errorf("want %q for single sample, got %q", want, got)
	}
}

func TestSparklineRingWrap(t *testing.T) {
	var h [20]float64
	h[18] = 0.0
	h[19] = 50.0
	h[0] = 100.0
	got := sparkline(h, 1, 3)
	want := "▁▄█"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

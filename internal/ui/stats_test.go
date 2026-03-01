package ui

import (
	"errors"
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
	if got.stats.entry != nil {
		t.Error("want stats.entry=nil (loading) on open")
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
	for _, key := range []tea.Msg{runeKey("r"), runeKey("a"), runeKey("s")} {
		m := statsPanel()
		got := update(m, key)
		if !got.stats.visible {
			t.Errorf("key %v: want stats.visible=true (panel should stay open)", key)
		}
	}
}

func TestUpdate_RKeyWithStatsPanelOpenResetsEntry(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.stats.visible = true
	m.stats.containerID = runningContainer.ID
	entry := docker.StatsEntry{CPUPerc: "5.00%"}
	m.stats.entry = &entry

	got, cmd := m.Update(runeKey("r"))
	if cmd == nil {
		t.Fatal("want non-nil batch cmd")
	}
	if !got.(App).loading {
		t.Error("want loading=true after r")
	}
	if got.(App).stats.entry != nil {
		t.Error("want stats.entry=nil (cleared for reload)")
	}
}

func TestUpdate_RKeyWithoutStatsPanelFetchesContainersOnly(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	_, cmd := m.Update(runeKey("r"))
	if cmd == nil {
		t.Fatal("want non-nil cmd")
	}
}

func TestUpdate_StatsMsgPopulatesEntry(t *testing.T) {
	m := statsPanel()
	entry := docker.StatsEntry{
		CPUPerc:  "0.42%",
		MemUsage: "3.4MiB / 1.9GiB",
		MemPerc:  "1.2%",
		NetIO:    "1.2kB / 456B",
		BlockIO:  "0B / 0B",
		PIDs:     "4",
	}
	got := update(m, docker.StatsMsg{Entry: entry})
	if got.stats.entry == nil {
		t.Fatal("want stats.entry set")
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

func TestUpdate_StatsMsgWhenPanelClosedIsNoop(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	entry := docker.StatsEntry{CPUPerc: "1.00%"}
	got := update(m, docker.StatsMsg{Entry: entry})
	if got.stats.entry != nil {
		t.Error("want stats.entry=nil when panel not open")
	}
}

func TestUpdate_StatsMsgErrorClosesPanelAndSetsErr(t *testing.T) {
	m := statsPanel()
	got := update(m, docker.StatsMsg{Err: errors.New("container not running")})
	if got.stats.visible {
		t.Error("want stats.visible=false on error")
	}
	if got.err == nil {
		t.Error("want err set")
	}
}

func TestUpdate_StatsMsgErrorDoesNotSetEntry(t *testing.T) {
	m := statsPanel()
	got := update(m, docker.StatsMsg{Err: errors.New("failed")})
	if got.stats.entry != nil {
		t.Error("want stats.entry=nil on error")
	}
}

func TestUpdate_StatsMsgSchedulesNextTick(t *testing.T) {
	m := statsPanel()
	entry := docker.StatsEntry{CPUPerc: "1.00%"}
	_, cmd := m.Update(docker.StatsMsg{Entry: entry})
	if cmd == nil {
		t.Error("want non-nil cmd (tick) after successful StatsMsg")
	}
}

func TestUpdate_StatsTickFetchesStats(t *testing.T) {
	mc := newStubClient()
	var gotID string
	mc.fetchStats = func(id string) tea.Cmd {
		gotID = id
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, []docker.Container{runningContainer})
	m.stats.visible = true
	m.stats.containerID = runningContainer.ID
	update(m, statsTickMsg{})
	if gotID != runningContainer.ID {
		t.Errorf("want FetchStats(%q) on tick, got %q", runningContainer.ID, gotID)
	}
}

func TestUpdate_StatsTickNoopWhenPanelClosed(t *testing.T) {
	mc := newStubClient()
	fetched := false
	mc.fetchStats = func(_ string) tea.Cmd {
		fetched = true
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, []docker.Container{runningContainer})
	m.stats.visible = false
	update(m, statsTickMsg{})
	if fetched {
		t.Error("want no FetchStats call when panel is closed")
	}
}

func statsPanel() App {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.stats.visible = true
	m.stats.container = runningContainer.Names
	m.stats.containerID = runningContainer.ID
	return m
}

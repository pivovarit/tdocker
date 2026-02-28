package ui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pivovarit/tdocker/internal/docker"
)

func TestUpdate_TKeyOnRunningOpensStatsPanel(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	got := update(m, runeKey("t"))
	if !got.statsVisible {
		t.Fatal("want statsVisible=true")
	}
	if got.statsContainer != runningContainer.Names {
		t.Errorf("want statsContainer=%q, got %q", runningContainer.Names, got.statsContainer)
	}
	if got.statsContainerID != runningContainer.ID {
		t.Errorf("want statsContainerID=%q, got %q", runningContainer.ID, got.statsContainerID)
	}
	if got.statsEntry != nil {
		t.Error("want statsEntry=nil (loading) on open")
	}
}

func TestUpdate_TKeyOnStoppedDoesNothing(t *testing.T) {
	m := modelWithSorted([]docker.Container{stoppedContainer})
	got := update(m, runeKey("t"))
	if got.statsVisible {
		t.Error("want statsVisible=false for non-running container")
	}
}

func TestUpdate_TKeyOnEmptyListDoesNothing(t *testing.T) {
	m := modelWithSorted(nil)
	got := update(m, runeKey("t"))
	if got.statsVisible {
		t.Error("want statsVisible=false for empty list")
	}
}

func TestUpdate_StatsEscClosesPanel(t *testing.T) {
	m := statsPanel()
	got := update(m, tea.KeyMsg{Type: tea.KeyEsc})
	if got.statsVisible {
		t.Error("want statsVisible=false after esc")
	}
}

func TestUpdate_StatsTClosesPanel(t *testing.T) {
	m := statsPanel()
	got := update(m, runeKey("t"))
	if got.statsVisible {
		t.Error("want statsVisible=false after t (toggle)")
	}
}

func TestUpdate_StatsCloseResetsState(t *testing.T) {
	m := statsPanel()
	entry := docker.StatsEntry{CPUPerc: "1.00%"}
	m.statsEntry = &entry
	got := update(m, tea.KeyMsg{Type: tea.KeyEsc})
	if got.statsEntry != nil {
		t.Error("want statsEntry=nil after close")
	}
	if got.statsContainer != "" {
		t.Error("want statsContainer empty after close")
	}
	if got.statsContainerID != "" {
		t.Error("want statsContainerID empty after close")
	}
}

func TestUpdate_StatsOtherKeysIgnored(t *testing.T) {
	for _, key := range []tea.Msg{runeKey("r"), runeKey("a"), runeKey("s")} {
		m := statsPanel()
		got := update(m, key)
		if !got.statsVisible {
			t.Errorf("key %v: want statsVisible=true (panel should stay open)", key)
		}
	}
}

func TestUpdate_RKeyWithStatsPanelOpenResetsEntry(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.statsVisible = true
	m.statsContainerID = runningContainer.ID
	entry := docker.StatsEntry{CPUPerc: "5.00%"}
	m.statsEntry = &entry

	got, cmd := m.Update(runeKey("r"))
	if cmd == nil {
		t.Fatal("want non-nil batch cmd")
	}
	if !got.(App).loading {
		t.Error("want loading=true after r")
	}
	if got.(App).statsEntry != nil {
		t.Error("want statsEntry=nil (cleared for reload)")
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
	if got.statsEntry == nil {
		t.Fatal("want statsEntry set")
	}
	if got.statsEntry.CPUPerc != "0.42%" {
		t.Errorf("want CPUPerc=%q, got %q", "0.42%", got.statsEntry.CPUPerc)
	}
	if got.statsEntry.MemUsage != "3.4MiB / 1.9GiB" {
		t.Errorf("want MemUsage=%q, got %q", "3.4MiB / 1.9GiB", got.statsEntry.MemUsage)
	}
	if got.statsEntry.PIDs != "4" {
		t.Errorf("want PIDs=%q, got %q", "4", got.statsEntry.PIDs)
	}
}

func TestUpdate_StatsMsgWhenPanelClosedIsNoop(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	entry := docker.StatsEntry{CPUPerc: "1.00%"}
	got := update(m, docker.StatsMsg{Entry: entry})
	if got.statsEntry != nil {
		t.Error("want statsEntry=nil when panel not open")
	}
}

func TestUpdate_StatsMsgErrorClosesPanelAndSetsErr(t *testing.T) {
	m := statsPanel()
	got := update(m, docker.StatsMsg{Err: errors.New("container not running")})
	if got.statsVisible {
		t.Error("want statsVisible=false on error")
	}
	if got.err == nil {
		t.Error("want err set")
	}
}

func TestUpdate_StatsMsgErrorDoesNotSetEntry(t *testing.T) {
	m := statsPanel()
	got := update(m, docker.StatsMsg{Err: errors.New("failed")})
	if got.statsEntry != nil {
		t.Error("want statsEntry=nil on error")
	}
}

func statsPanel() App {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.statsVisible = true
	m.statsContainer = runningContainer.Names
	m.statsContainerID = runningContainer.ID
	return m
}

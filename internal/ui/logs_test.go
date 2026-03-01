package ui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pivovarit/tdocker/internal/docker"
)

func logsOpen(mc *stubClient, c docker.Container) App {
	m := modelWithMock(mc, []docker.Container{c})
	m.logsVisible = true
	m.logsContainerID = c.ID
	m.logsContainer = c.Names
	return m
}

func logsOpenWithLines(lines []string) App {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.logsVisible = true
	m.logsContainerID = runningContainer.ID
	m.logsContainer = runningContainer.Names
	m.logsLines = lines
	m.logsScrollOffset = 0
	m.logsAutoScroll = false
	return m
}

func TestLogs_FToggle_SwitchesToAllTail(t *testing.T) {
	mc := newStubClient()
	var gotTail string
	mc.startLogs = func(_ context.Context, _ string, tail string, _ int) tea.Cmd {
		gotTail = tail
		return func() tea.Msg { return nil }
	}
	m := logsOpen(mc, runningContainer)
	update(m, runeKey("f"))
	if gotTail != "all" {
		t.Errorf("want tail=%q after first f-toggle, got %q", "all", gotTail)
	}
}

func TestLogs_FToggle_SwitchesBackToTail200(t *testing.T) {
	mc := newStubClient()
	var gotTail string
	mc.startLogs = func(_ context.Context, _ string, tail string, _ int) tea.Cmd {
		gotTail = tail
		return func() tea.Msg { return nil }
	}
	m := logsOpen(mc, runningContainer)
	m.logsAllMode = true
	update(m, runeKey("f"))
	if gotTail != "200" {
		t.Errorf("want tail=%q after toggling back, got %q", "200", gotTail)
	}
}

func TestLogs_FToggle_ClearsLines(t *testing.T) {
	mc := newStubClient()
	m := logsOpen(mc, runningContainer)
	m.logsLines = []string{"line1", "line2", "line3"}
	got := update(m, runeKey("f"))
	if len(got.logsLines) != 0 {
		t.Errorf("want logsLines cleared after f-toggle, got %d lines", len(got.logsLines))
	}
}

func TestLogs_FToggle_IncrementsGen(t *testing.T) {
	mc := newStubClient()
	var gotGen int
	mc.startLogs = func(_ context.Context, _ string, _ string, gen int) tea.Cmd {
		gotGen = gen
		return func() tea.Msg { return nil }
	}
	m := logsOpen(mc, runningContainer)
	m.logsGen = 3
	update(m, runeKey("f"))
	if gotGen != 4 {
		t.Errorf("want gen=4 after f-toggle, got %d", gotGen)
	}
}

func TestLogs_FToggle_PassesContainerID(t *testing.T) {
	mc := newStubClient()
	var gotID string
	mc.startLogs = func(_ context.Context, id string, _ string, _ int) tea.Cmd {
		gotID = id
		return func() tea.Msg { return nil }
	}
	m := logsOpen(mc, runningContainer)
	update(m, runeKey("f"))
	if gotID != runningContainer.ID {
		t.Errorf("want StartLogs called with id=%q, got %q", runningContainer.ID, gotID)
	}
}

func TestLogs_GKey_ScrollsToTopAndDisablesAutoScroll(t *testing.T) {
	lines := make([]string, 20)
	m := logsOpenWithLines(lines)
	m.logsScrollOffset = 10
	m.logsAutoScroll = true
	got := update(m, runeKey("g"))
	if got.logsScrollOffset != 0 {
		t.Errorf("want logsScrollOffset=0, got %d", got.logsScrollOffset)
	}
	if got.logsAutoScroll {
		t.Error("want logsAutoScroll=false after g")
	}
}

func TestLogs_HomeKey_ScrollsToTop(t *testing.T) {
	lines := make([]string, 20)
	m := logsOpenWithLines(lines)
	m.logsScrollOffset = 7
	got := update(m, tea.KeyMsg{Type: tea.KeyHome})
	if got.logsScrollOffset != 0 {
		t.Errorf("want logsScrollOffset=0 after Home, got %d", got.logsScrollOffset)
	}
}

func TestLogs_ShiftGKey_ScrollsToBottomAndEnablesAutoScroll(t *testing.T) {
	lines := make([]string, 20)
	m := logsOpenWithLines(lines)
	got := update(m, runeKey("G"))
	want := max(0, len(lines)-(logsPanelHeight-2))
	if got.logsScrollOffset != want {
		t.Errorf("want logsScrollOffset=%d after G, got %d", want, got.logsScrollOffset)
	}
	if !got.logsAutoScroll {
		t.Error("want logsAutoScroll=true after G")
	}
}

func TestLogs_EndKey_ScrollsToBottom(t *testing.T) {
	lines := make([]string, 20)
	m := logsOpenWithLines(lines)
	got := update(m, tea.KeyMsg{Type: tea.KeyEnd})
	want := max(0, len(lines)-(logsPanelHeight-2))
	if got.logsScrollOffset != want {
		t.Errorf("want logsScrollOffset=%d after End, got %d", want, got.logsScrollOffset)
	}
}

func TestLogs_GKey_OnFewLines_StaysAtZero(t *testing.T) {
	m := logsOpenWithLines([]string{"only one line"})
	m.logsScrollOffset = 0
	got := update(m, runeKey("G"))
	if got.logsScrollOffset != 0 {
		t.Errorf("want logsScrollOffset=0 when few lines, got %d", got.logsScrollOffset)
	}
}

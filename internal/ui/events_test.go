package ui

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

func lifecycleEvent(action string, gen int) docker.EventLineMsg {
	return docker.EventLineMsg{
		Event: docker.Event{
			Type:   "container",
			Action: action,
			Actor:  docker.EventActor{Attributes: map[string]string{"name": "test-container"}},
		},
		Next: func() tea.Msg { return nil },
		Gen:  gen,
	}
}

func nonLifecycleEvent(gen int) docker.EventLineMsg {
	return docker.EventLineMsg{
		Event: docker.Event{Type: "network", Action: "connect"},
		Next:  func() tea.Msg { return nil },
		Gen:   gen,
	}
}

func TestInit_StartsBackgroundEventStream(t *testing.T) {
	mc := newStubClient()
	startEventsCalled := false
	mc.startEvents = func(_ context.Context, gen int) tea.Cmd {
		if gen != 1 {
			t.Errorf("want StartEvents called with gen=1, got %d", gen)
		}
		startEventsCalled = true
		return func() tea.Msg { return nil }
	}
	m := newWithClient(mc, "")
	m.Init()
	if !startEventsCalled {
		t.Error("want StartEvents called from Init")
	}
	if m.bgEventsGen != 1 {
		t.Errorf("want bgEventsGen=1, got %d", m.bgEventsGen)
	}
}

func TestEventLineMsg_LifecycleSchedulesDebounce(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.bgEventsGen = 1

	got, cmd := m.Update(lifecycleEvent("stop", 1))
	app := got.(App)

	if !app.pendingRefresh {
		t.Error("want pendingRefresh=true after lifecycle event")
	}
	if cmd == nil {
		t.Error("want non-nil debounce cmd after lifecycle event")
	}
}

func TestEventLineMsg_NonLifecycleDoesNotScheduleDebounce(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.bgEventsGen = 1

	got, _ := m.Update(nonLifecycleEvent(1))
	app := got.(App)

	if app.pendingRefresh {
		t.Error("want pendingRefresh=false for non-lifecycle event")
	}
}

func TestEventLineMsg_StaleGenIsIgnored(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.bgEventsGen = 5

	got, _ := m.Update(lifecycleEvent("stop", 3))
	app := got.(App)

	if app.pendingRefresh {
		t.Error("want pendingRefresh=false for stale gen")
	}
}

func TestEventLineMsg_NoDuplicateDebounce(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.bgEventsGen = 1
	m.pendingRefresh = true

	_, cmd := m.Update(lifecycleEvent("die", 1))
	mc := newStubClient()
	fetched := false
	mc.fetchContainers = func(_ bool) tea.Cmd {
		fetched = true
		return func() tea.Msg { return nil }
	}
	m.client = mc
	_, _ = m.Update(lifecycleEvent("die", 1))
	_ = cmd
	if fetched {
		t.Error("want FetchContainers NOT called while pendingRefresh is set")
	}
}

func TestAutoRefreshMsg_TriggersFetch(t *testing.T) {
	mc := newStubClient()
	fetched := false
	mc.fetchContainers = func(_ bool) tea.Cmd {
		fetched = true
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, nil)
	m.pendingRefresh = true

	got, cmd := m.Update(autoRefreshMsg{})
	app := got.(App)

	if app.pendingRefresh {
		t.Error("want pendingRefresh cleared after autoRefreshMsg")
	}
	if !app.loading {
		t.Error("want loading=true after autoRefreshMsg")
	}
	if cmd == nil {
		t.Error("want non-nil fetch cmd after autoRefreshMsg")
	}
	if !fetched {
		t.Error("want FetchContainers called after autoRefreshMsg")
	}
}

func TestAutoRefreshMsg_SkipsWhenAlreadyLoading(t *testing.T) {
	mc := newStubClient()
	fetched := false
	mc.fetchContainers = func(_ bool) tea.Cmd {
		fetched = true
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, nil)
	m.loading = true
	m.pendingRefresh = true

	_, cmd := m.Update(autoRefreshMsg{})
	if fetched {
		t.Error("want FetchContainers NOT called when already loading")
	}
	if cmd != nil {
		t.Error("want nil cmd when already loading")
	}
}

func TestEventEndMsg_IncrementsGenAndSchedulesRestart(t *testing.T) {
	m := modelWithSorted(nil)
	m.bgEventsGen = 1

	got, cmd := m.Update(docker.EventEndMsg{Gen: 1})
	app := got.(App)

	if app.bgEventsGen != 2 {
		t.Errorf("want bgEventsGen=2 after stream death, got %d", app.bgEventsGen)
	}
	if cmd == nil {
		t.Error("want restart cmd after stream death")
	}
}

func TestEventEndMsg_StaleGenIsIgnored(t *testing.T) {
	m := modelWithSorted(nil)
	m.bgEventsGen = 3

	got, cmd := m.Update(docker.EventEndMsg{Gen: 1})
	app := got.(App)

	if app.bgEventsGen != 3 {
		t.Errorf("want bgEventsGen unchanged at 3, got %d", app.bgEventsGen)
	}
	if cmd != nil {
		t.Error("want nil cmd for stale EventEndMsg")
	}
}

func TestBgEventsRestartMsg_StartsNewStream(t *testing.T) {
	mc := newStubClient()
	var startedGen int
	mc.startEvents = func(_ context.Context, gen int) tea.Cmd {
		startedGen = gen
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, nil)
	m.bgEventsGen = 2

	_, cmd := m.Update(bgEventsRestartMsg{gen: 2})
	if cmd == nil {
		t.Error("want non-nil StartEvents cmd after bgEventsRestartMsg")
	}
	if startedGen != 2 {
		t.Errorf("want StartEvents called with gen=2, got %d", startedGen)
	}
}

func TestBgEventsRestartMsg_StaleIsNoOp(t *testing.T) {
	mc := newStubClient()
	called := false
	mc.startEvents = func(_ context.Context, _ int) tea.Cmd {
		called = true
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, nil)
	m.bgEventsGen = 3

	_, cmd := m.Update(bgEventsRestartMsg{gen: 2})
	if called {
		t.Error("want StartEvents NOT called for stale bgEventsRestartMsg")
	}
	if cmd != nil {
		t.Error("want nil cmd for stale bgEventsRestartMsg")
	}
}

func TestVKey_TogglesEventsPanelWithoutStartingStream(t *testing.T) {
	mc := newStubClient()
	startCalled := false
	mc.startEvents = func(_ context.Context, _ int) tea.Cmd {
		startCalled = true
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, nil)

	got := update(m, runeKey("v"))
	if !got.events.visible {
		t.Error("want events panel visible after v")
	}
	if startCalled {
		t.Error("want StartEvents NOT called on v (background stream already running)")
	}
}

func TestVKey_ClosesEventsPanelOnSecondPress(t *testing.T) {
	m := modelWithSorted(nil)
	m.events.visible = true

	got := update(m, runeKey("v"))
	if got.events.visible {
		t.Error("want events panel hidden after second v")
	}
}

func TestEventLineMsg_PopulatesPanelWhenVisible(t *testing.T) {
	m := modelWithSorted(nil)
	m.bgEventsGen = 1
	m.events.visible = true

	got := update(m, lifecycleEvent("start", 1))
	if len(got.events.events) != 1 {
		t.Errorf("want 1 event in panel, got %d", len(got.events.events))
	}
}

func TestEventLineMsg_DoesNotPopulatePanelWhenHidden(t *testing.T) {
	m := modelWithSorted(nil)
	m.bgEventsGen = 1
	m.events.visible = false

	got := update(m, lifecycleEvent("start", 1))
	if len(got.events.events) != 0 {
		t.Errorf("want 0 events buffered when panel is hidden, got %d", len(got.events.events))
	}
}

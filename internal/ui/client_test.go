package ui

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

func modelWithMock(mc *stubClient, containers []docker.Container) App {
	m := newWithClient(mc, "")
	m.sorted = containers
	m.containers = containers
	m.loading = false
	m.width = 120
	return m.rebuildTable("")
}

func TestClient_RKey_FetchesWithCurrentShowAll(t *testing.T) {
	for _, showAll := range []bool{true, false} {
		mc := newStubClient()
		var gotAll bool
		mc.fetchContainers = func(all bool) tea.Cmd {
			gotAll = all
			return func() tea.Msg { return nil }
		}
		m := modelWithMock(mc, nil)
		m.showAll = showAll
		update(m, runeKey("r"))
		if gotAll != showAll {
			t.Errorf("showAll=%v: FetchContainers called with all=%v", showAll, gotAll)
		}
	}
}

func TestClient_AKey_TogglesShowAllBeforeFetch(t *testing.T) {
	mc := newStubClient()
	var gotAll bool
	mc.fetchContainers = func(all bool) tea.Cmd {
		gotAll = all
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, nil)
	m.showAll = false
	update(m, runeKey("A"))
	if !gotAll {
		t.Error("want FetchContainers called with all=true after toggling from false")
	}
}

func TestClient_ConfirmStop_CallsStopContainerWithID(t *testing.T) {
	mc := newStubClient()
	var gotID string
	mc.stopContainer = func(id string) tea.Cmd {
		gotID = id
		return func() tea.Msg { return nil }
	}
	m := confirming("stop", runningContainer)
	m.client = mc
	update(m, runeKey("y"))
	if gotID != runningContainer.ID {
		t.Errorf("want StopContainer(%q), got %q", runningContainer.ID, gotID)
	}
}

func TestClient_ConfirmStart_CallsStartContainerWithID(t *testing.T) {
	mc := newStubClient()
	var gotID string
	mc.startContainer = func(id string) tea.Cmd {
		gotID = id
		return func() tea.Msg { return nil }
	}
	m := confirming("start", stoppedContainer)
	m.client = mc
	update(m, runeKey("y"))
	if gotID != stoppedContainer.ID {
		t.Errorf("want StartContainer(%q), got %q", stoppedContainer.ID, gotID)
	}
}

func TestClient_ConfirmDelete_CallsDeleteContainerWithID(t *testing.T) {
	mc := newStubClient()
	var gotID string
	mc.deleteContainer = func(id string) tea.Cmd {
		gotID = id
		return func() tea.Msg { return nil }
	}
	m := confirming("delete", stoppedContainer)
	m.client = mc
	update(m, runeKey("y"))
	if gotID != stoppedContainer.ID {
		t.Errorf("want DeleteContainer(%q), got %q", stoppedContainer.ID, gotID)
	}
}

func TestClient_TKey_CallsFetchStatsWithID(t *testing.T) {
	mc := newStubClient()
	var gotID string
	mc.fetchStats = func(id string) tea.Cmd {
		gotID = id
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, []docker.Container{runningContainer})
	update(m, runeKey("t"))
	if gotID != runningContainer.ID {
		t.Errorf("want FetchStats(%q), got %q", runningContainer.ID, gotID)
	}
}

func TestClient_LKey_CallsStartLogsWithID(t *testing.T) {
	mc := newStubClient()
	var gotID string
	mc.startLogs = func(_ context.Context, id string, _ string, _ int) tea.Cmd {
		gotID = id
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, []docker.Container{runningContainer})
	update(m, runeKey("l"))
	if gotID != runningContainer.ID {
		t.Errorf("want StartLogs(%q), got %q", runningContainer.ID, gotID)
	}
}

func TestClient_IKey_CallsInspectContainerWithID(t *testing.T) {
	mc := newStubClient()
	var gotID string
	mc.inspectContainer = func(id string) tea.Cmd {
		gotID = id
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, []docker.Container{runningContainer})
	update(m, runeKey("i"))
	if gotID != runningContainer.ID {
		t.Errorf("want InspectContainer(%q), got %q", runningContainer.ID, gotID)
	}
}

func TestClient_EKey_CallsCheckShellAvailableWithID(t *testing.T) {
	mc := newStubClient()
	var gotID string
	mc.checkShellAvail = func(id string) tea.Cmd {
		gotID = id
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, []docker.Container{runningContainer})
	update(m, runeKey("e"))
	if gotID != runningContainer.ID {
		t.Errorf("want CheckShellAvailable(%q), got %q", runningContainer.ID, gotID)
	}
}

func TestClient_ShellAvailableMsg_CallsExecContainer(t *testing.T) {
	mc := newStubClient()
	var gotID, gotShell string
	mc.execContainer = func(id, shell string) tea.Cmd {
		gotID = id
		gotShell = shell
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, nil)
	update(m, docker.ShellAvailableMsg{ID: runningContainer.ID, Available: true, Shell: "bash"})
	if gotID != runningContainer.ID {
		t.Errorf("want ExecContainer(%q), got %q", runningContainer.ID, gotID)
	}
	if gotShell != "bash" {
		t.Errorf("want shell=%q, got %q", "bash", gotShell)
	}
}

func TestClient_StopMsg_FetchesContainers(t *testing.T) {
	mc := newStubClient()
	fetched := false
	mc.fetchContainers = func(_ bool) tea.Cmd {
		fetched = true
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, []docker.Container{runningContainer})
	m.op = OpStopping
	update(m, docker.StopMsg{})
	if !fetched {
		t.Error("want FetchContainers called after StopMsg")
	}
}

func TestClient_StartMsg_FetchesContainers(t *testing.T) {
	mc := newStubClient()
	fetched := false
	mc.fetchContainers = func(_ bool) tea.Cmd {
		fetched = true
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, []docker.Container{stoppedContainer})
	m.op = OpStarting
	update(m, docker.StartMsg{})
	if !fetched {
		t.Error("want FetchContainers called after StartMsg")
	}
}

func TestClient_ExecDoneMsg_FetchesContainers(t *testing.T) {
	mc := newStubClient()
	fetched := false
	mc.fetchContainers = func(_ bool) tea.Cmd {
		fetched = true
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, nil)
	update(m, docker.ExecDoneMsg{})
	if !fetched {
		t.Error("want FetchContainers called after successful ExecDoneMsg")
	}
}

func TestClient_ShellUnavailableMsg_SetsError(t *testing.T) {
	mc := newStubClient()
	m := modelWithMock(mc, nil)
	got := update(m, docker.ShellAvailableMsg{ID: runningContainer.ID, Available: false})
	if got.err == nil {
		t.Error("want err set when shell not available")
	}
}

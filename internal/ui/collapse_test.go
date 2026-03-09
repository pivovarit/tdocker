package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

func composeLabels(project string) docker.Labels {
	return docker.Labels{"com.docker.compose.project": project}
}

func TestCollapseSummary(t *testing.T) {
	tests := []struct {
		name           string
		project        string
		containers     []docker.Container
		wantNames      string
		wantState      string
		wantID         string
		wantHasProject bool
	}{
		{
			name:    "single state - all running",
			project: "myapp",
			containers: []docker.Container{
				{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
				{ID: "a2", Names: "db", State: "running", Labels: composeLabels("myapp")},
			},
			wantNames:      "▶ myapp (2 running)",
			wantState:      "collapsed",
			wantID:         "",
			wantHasProject: true,
		},
		{
			name:    "mixed states - running and exited",
			project: "myapp",
			containers: []docker.Container{
				{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
				{ID: "a2", Names: "worker", State: "running", Labels: composeLabels("myapp")},
				{ID: "a3", Names: "db", State: "exited", Labels: composeLabels("myapp")},
			},
			wantNames:      "▶ myapp (2 running, 1 exited)",
			wantState:      "collapsed",
			wantID:         "",
			wantHasProject: true,
		},
		{
			name:    "state ordering - running first, paused second, rest alphabetical",
			project: "stack",
			containers: []docker.Container{
				{ID: "a1", State: "exited", Labels: composeLabels("stack")},
				{ID: "a2", State: "paused", Labels: composeLabels("stack")},
				{ID: "a3", State: "running", Labels: composeLabels("stack")},
				{ID: "a4", State: "dead", Labels: composeLabels("stack")},
				{ID: "a5", State: "running", Labels: composeLabels("stack")},
			},
			wantNames:      "▶ stack (2 running, 1 paused, 1 dead, 1 exited)",
			wantState:      "collapsed",
			wantID:         "",
			wantHasProject: true,
		},
		{
			name:    "single container",
			project: "solo",
			containers: []docker.Container{
				{ID: "s1", Names: "app", State: "running", Labels: composeLabels("solo")},
			},
			wantNames:      "▶ solo (1 running)",
			wantState:      "collapsed",
			wantID:         "",
			wantHasProject: true,
		},
		{
			name:    "all same non-running state",
			project: "stopped",
			containers: []docker.Container{
				{ID: "x1", State: "exited", Labels: composeLabels("stopped")},
				{ID: "x2", State: "exited", Labels: composeLabels("stopped")},
				{ID: "x3", State: "exited", Labels: composeLabels("stopped")},
			},
			wantNames:      "▶ stopped (3 exited)",
			wantState:      "collapsed",
			wantID:         "",
			wantHasProject: true,
		},
		{
			name:    "other fields are empty",
			project: "app",
			containers: []docker.Container{
				{ID: "c1", Names: "web", Image: "nginx", State: "running", Status: "Up 5m", Ports: "80/tcp", Labels: composeLabels("app")},
			},
			wantNames:      "▶ app (1 running)",
			wantState:      "collapsed",
			wantID:         "",
			wantHasProject: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collapseSummary(tt.project, tt.containers)

			if got.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", got.ID, tt.wantID)
			}
			if got.Names != tt.wantNames {
				t.Errorf("Names = %q, want %q", got.Names, tt.wantNames)
			}
			if got.State != tt.wantState {
				t.Errorf("State = %q, want %q", got.State, tt.wantState)
			}
			if got.Image != "" {
				t.Errorf("Image = %q, want empty", got.Image)
			}
			if got.Command != "" {
				t.Errorf("Command = %q, want empty", got.Command)
			}
			if got.Status != "" {
				t.Errorf("Status = %q, want empty", got.Status)
			}
			if got.RunningFor != "" {
				t.Errorf("RunningFor = %q, want empty", got.RunningFor)
			}
			if got.Ports != "" {
				t.Errorf("Ports = %q, want empty", got.Ports)
			}
			if tt.wantHasProject {
				if got.ComposeProject() != tt.project {
					t.Errorf("ComposeProject() = %q, want %q", got.ComposeProject(), tt.project)
				}
			}
		})
	}
}

func composeServiceLabels(project, service string) docker.Labels {
	return docker.Labels{
		"com.docker.compose.project": project,
		"com.docker.compose.service": service,
	}
}

func TestFiltered_CollapsedGroupEmitsSummary(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", Image: "nginx", State: "running", Labels: composeServiceLabels("myapp", "web")},
		{ID: "a2", Names: "db", Image: "postgres", State: "running", Labels: composeServiceLabels("myapp", "db")},
		{ID: "s1", Names: "standalone", Image: "alpine", State: "running"},
	}
	m := modelWithSorted(containers)
	m.collapsedProjects["myapp"] = true

	got := m.filtered()
	if len(got) != 2 {
		t.Fatalf("want 2 rows (1 summary + 1 standalone), got %d", len(got))
	}
	if got[0].State != "collapsed" {
		t.Errorf("first row state = %q, want %q", got[0].State, "collapsed")
	}
	if got[0].ComposeProject() != "myapp" {
		t.Errorf("summary project = %q, want %q", got[0].ComposeProject(), "myapp")
	}
	if got[1].Names != "standalone" {
		t.Errorf("second row = %q, want %q", got[1].Names, "standalone")
	}
}

func TestFiltered_ExpandedGroupPassesThrough(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
		{ID: "a2", Names: "db", State: "running", Labels: composeLabels("myapp")},
		{ID: "s1", Names: "standalone", State: "running"},
	}
	m := modelWithSorted(containers)

	got := m.filtered()
	if len(got) != 3 {
		t.Fatalf("want 3, got %d", len(got))
	}
}

func TestFiltered_FilterAutoExpandsMatchingGroups(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", Image: "nginx", State: "running", Labels: composeServiceLabels("myapp", "web")},
		{ID: "a2", Names: "db", Image: "postgres", State: "running", Labels: composeServiceLabels("myapp", "db")},
		{ID: "s1", Names: "standalone", Image: "alpine", State: "running"},
	}
	m := modelWithSorted(containers)
	m.collapsedProjects["myapp"] = true
	m.filterQuery = "web"

	got := m.filtered()
	if len(got) != 1 {
		t.Fatalf("want 1 (matched container), got %d", len(got))
	}
	if got[0].Names != "web" {
		t.Errorf("got %q, want %q", got[0].Names, "web")
	}
}

func TestFiltered_FilterNoMatchesHidesCollapsedGroup(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
		{ID: "a2", Names: "db", State: "running", Labels: composeLabels("myapp")},
	}
	m := modelWithSorted(containers)
	m.collapsedProjects["myapp"] = true
	m.filterQuery = "zzznomatch"

	got := m.filtered()
	if len(got) != 0 {
		t.Errorf("want 0, got %d", len(got))
	}
}

func TestFiltered_MultipleCollapsedGroups(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("app1")},
		{ID: "a2", Names: "db", State: "running", Labels: composeLabels("app1")},
		{ID: "b1", Names: "api", State: "exited", Labels: composeLabels("app2")},
		{ID: "b2", Names: "worker", State: "running", Labels: composeLabels("app2")},
	}
	m := modelWithSorted(containers)
	m.collapsedProjects["app1"] = true
	m.collapsedProjects["app2"] = true

	got := m.filtered()
	if len(got) != 2 {
		t.Fatalf("want 2 summaries, got %d", len(got))
	}
	if got[0].State != "collapsed" || got[1].State != "collapsed" {
		t.Errorf("both rows should be collapsed summaries, got states %q and %q", got[0].State, got[1].State)
	}
	projects := map[string]bool{got[0].ComposeProject(): true, got[1].ComposeProject(): true}
	if !projects["app1"] || !projects["app2"] {
		t.Errorf("want projects app1 and app2, got %v", projects)
	}
}

func TestLeftArrow_CollapsesComposeGroup(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
		{ID: "a2", Names: "db", State: "running", Labels: composeLabels("myapp")},
		{ID: "s1", Names: "standalone", State: "running"},
	}
	m := modelWithSorted(containers)
	got := update(m, tea.KeyPressMsg{Code: tea.KeyLeft})
	if !got.collapsedProjects["myapp"] {
		t.Fatal("want myapp collapsed after left arrow")
	}
	filtered := got.filtered()
	if len(filtered) != 2 {
		t.Fatalf("want 2 rows (summary + standalone), got %d", len(filtered))
	}
	if filtered[0].State != "collapsed" {
		t.Errorf("first row state = %q, want collapsed", filtered[0].State)
	}
}

func TestRightArrow_ExpandsCollapsedGroup(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
		{ID: "a2", Names: "db", State: "running", Labels: composeLabels("myapp")},
		{ID: "s1", Names: "standalone", State: "running"},
	}
	m := modelWithSorted(containers)
	m.collapsedProjects["myapp"] = true
	m = m.rebuildTable("")
	got := update(m, tea.KeyPressMsg{Code: tea.KeyRight})
	if got.collapsedProjects["myapp"] {
		t.Fatal("want myapp expanded after right arrow")
	}
	filtered := got.filtered()
	if len(filtered) != 3 {
		t.Fatalf("want 3 rows after expand, got %d", len(filtered))
	}
}

func TestLeftArrow_NoopOnNonComposeContainer(t *testing.T) {
	containers := []docker.Container{
		{ID: "s1", Names: "standalone", State: "running"},
	}
	m := modelWithSorted(containers)
	got := update(m, tea.KeyPressMsg{Code: tea.KeyLeft})
	if len(got.collapsedProjects) != 0 {
		t.Errorf("want no collapsed projects, got %v", got.collapsedProjects)
	}
}

func TestLeftArrow_NoopOnAlreadyCollapsedRow(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
		{ID: "a2", Names: "db", State: "running", Labels: composeLabels("myapp")},
	}
	m := modelWithSorted(containers)
	m.collapsedProjects["myapp"] = true
	m = m.rebuildTable("")
	got := update(m, tea.KeyPressMsg{Code: tea.KeyLeft})
	if !got.collapsedProjects["myapp"] {
		t.Fatal("want myapp still collapsed")
	}
}

func TestRightArrow_NoopOnRegularContainer(t *testing.T) {
	containers := []docker.Container{
		{ID: "s1", Names: "standalone", State: "running"},
	}
	m := modelWithSorted(containers)
	got := update(m, tea.KeyPressMsg{Code: tea.KeyRight})
	if len(got.collapsedProjects) != 0 {
		t.Errorf("want no collapsed projects changed, got %v", got.collapsedProjects)
	}
}

func TestCollapse_CursorLandsOnSummaryRow(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
		{ID: "a2", Names: "db", State: "running", Labels: composeLabels("myapp")},
		{ID: "s1", Names: "standalone", State: "running"},
	}
	m := modelWithSorted(containers)
	m.table.SetCursor(0)
	got := update(m, tea.KeyPressMsg{Code: tea.KeyLeft})
	filtered := got.filtered()
	cursor := got.table.Cursor()
	if cursor < 0 || cursor >= len(filtered) {
		t.Fatalf("cursor %d out of range [0, %d)", cursor, len(filtered))
	}
	if filtered[cursor].State != "collapsed" {
		t.Errorf("cursor should be on summary row, got state=%q", filtered[cursor].State)
	}
}

func TestOperations_NoopOnCollapsedRow(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
		{ID: "a2", Names: "db", State: "running", Labels: composeLabels("myapp")},
	}
	m := modelWithSorted(containers)
	m.collapsedProjects["myapp"] = true
	m = m.rebuildTable("")

	keys := []string{"D", "P", "N", "e", "x", "l", "i", "c", "t"}
	for _, k := range keys {
		got := update(m, runeKey(k))
		if got.op.kind != OpNone {
			t.Errorf("key %q: want OpNone on collapsed row, got %v", k, got.op.kind)
		}
		if got.logs.visible {
			t.Errorf("key %q: want logs not opened on collapsed row", k)
		}
		if got.inspect.visible {
			t.Errorf("key %q: want inspect not opened on collapsed row", k)
		}
		if got.stats.visible {
			t.Errorf("key %q: want stats not opened on collapsed row", k)
		}
	}
}

func TestComposeStop_OnCollapsedRow(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
		{ID: "a2", Names: "db", State: "running", Labels: composeLabels("myapp")},
	}
	m := modelWithSorted(containers)
	m.collapsedProjects["myapp"] = true
	m = m.rebuildTable("")

	got := update(m, runeKey("S"))
	if got.op.kind != OpConfirming {
		t.Fatalf("want OpConfirming, got %v", got.op.kind)
	}
	if got.op.action != "compose-stop" {
		t.Errorf("want action compose-stop, got %q", got.op.action)
	}
	if got.op.id != "myapp" {
		t.Errorf("want id=myapp (project name), got %q", got.op.id)
	}
}

func TestComposeRestart_OnCollapsedRow(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
		{ID: "a2", Names: "db", State: "running", Labels: composeLabels("myapp")},
	}
	m := modelWithSorted(containers)
	m.collapsedProjects["myapp"] = true
	m = m.rebuildTable("")

	got := update(m, runeKey("R"))
	if got.op.kind != OpConfirming {
		t.Fatalf("want OpConfirming, got %v", got.op.kind)
	}
	if got.op.action != "compose-restart" {
		t.Errorf("want action compose-restart, got %q", got.op.action)
	}
	if got.op.id != "myapp" {
		t.Errorf("want id=myapp (project name), got %q", got.op.id)
	}
}

func TestComposeStart_OnCollapsedRowAllExited(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "exited", Labels: composeLabels("myapp")},
		{ID: "a2", Names: "db", State: "exited", Labels: composeLabels("myapp")},
	}
	m := modelWithSorted(containers)
	m.collapsedProjects["myapp"] = true
	m = m.rebuildTable("")

	got := update(m, runeKey("S"))
	if got.op.kind != OpConfirming {
		t.Fatalf("want OpConfirming, got %v", got.op.kind)
	}
	if got.op.action != "compose-start" {
		t.Errorf("want action compose-start, got %q", got.op.action)
	}

	got = update(m, runeKey("R"))
	if got.op.action != "compose-start" {
		t.Errorf("want action compose-start for R key too, got %q", got.op.action)
	}
}

func TestComposeStop_ConfirmCallsClient(t *testing.T) {
	mc := newStubClient()
	var gotProject string
	mc.stopCompose = func(project string) tea.Cmd {
		gotProject = project
		return func() tea.Msg { return nil }
	}
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
	}
	m := modelWithMock(mc, containers)
	m.op = operationState{kind: OpConfirming, id: "myapp", action: "compose-stop", name: "▶ myapp (1 running)"}
	update(m, runeKey("y"))
	if gotProject != "myapp" {
		t.Errorf("want StopCompose(%q), got %q", "myapp", gotProject)
	}
}

func TestExpand_CursorLandsOnFirstContainerInGroup(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
		{ID: "a2", Names: "db", State: "running", Labels: composeLabels("myapp")},
		{ID: "s1", Names: "standalone", State: "running"},
	}
	m := modelWithSorted(containers)
	m.collapsedProjects["myapp"] = true
	m = m.rebuildTable("")
	m.table.SetCursor(0)
	got := update(m, tea.KeyPressMsg{Code: tea.KeyRight})
	filtered := got.filtered()
	cursor := got.table.Cursor()
	if cursor < 0 || cursor >= len(filtered) {
		t.Fatalf("cursor %d out of range [0, %d)", cursor, len(filtered))
	}
	c := filtered[cursor]
	if c.ComposeProject() != "myapp" {
		t.Errorf("cursor should be on first myapp container, got project=%q name=%q", c.ComposeProject(), c.Names)
	}
}

func TestBuildTableName_CollapsedRow(t *testing.T) {
	containers := []docker.Container{
		{Names: "▶ myapp (2 running)", State: "collapsed", Labels: composeLabels("myapp")},
	}
	got := buildTableName(containers, 0)
	want := "▶ myapp (2 running)"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestHelpBar_CollapsedRowShowsExpandHint(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
		{ID: "a2", Names: "db", State: "running", Labels: composeLabels("myapp")},
	}
	m := modelWithSorted(containers)
	m.collapsedProjects["myapp"] = true
	m = m.rebuildTable("")

	bar := m.helpBar()
	if !strings.Contains(bar, "expand") {
		t.Errorf("want help bar to mention expand, got %q", bar)
	}
}

func TestCollapseExpand_FullCycle(t *testing.T) {
	containers := []docker.Container{
		{ID: "s1", Names: "solo", State: "running"},
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
		{ID: "a2", Names: "db", State: "exited", Labels: composeLabels("myapp")},
	}
	m := modelWithSorted(containers)
	m.table.SetCursor(1)

	m = update(m, tea.KeyPressMsg{Code: tea.KeyLeft})
	filtered := m.filtered()
	if len(filtered) != 2 {
		t.Fatalf("after collapse: want 2 rows, got %d", len(filtered))
	}
	if filtered[1].State != "collapsed" {
		t.Error("want second row to be collapsed summary")
	}

	m.table.SetCursor(1)
	m = update(m, tea.KeyPressMsg{Code: tea.KeyRight})
	filtered = m.filtered()
	if len(filtered) != 3 {
		t.Fatalf("after expand: want 3 rows, got %d", len(filtered))
	}

	m.table.SetCursor(1)
	m = update(m, tea.KeyPressMsg{Code: tea.KeyLeft})
	m.filterQuery = "web"
	filtered = m.filtered()
	if len(filtered) != 1 {
		t.Fatalf("with filter: want 1 match, got %d", len(filtered))
	}
	if filtered[0].ID != "a1" {
		t.Error("want matched container, not summary")
	}

	m.filterQuery = ""
	filtered = m.filtered()
	if len(filtered) != 2 {
		t.Fatalf("after clear filter: want 2 rows (collapsed), got %d", len(filtered))
	}
}

func TestCollapse_PersistsAcrossRefresh(t *testing.T) {
	containers := []docker.Container{
		{ID: "a1", Names: "web", State: "running", Labels: composeLabels("myapp")},
		{ID: "a2", Names: "db", State: "running", Labels: composeLabels("myapp")},
	}
	m := modelWithSorted(containers)
	m.collapsedProjects["myapp"] = true

	m = update(m, docker.ContainersMsg(containers))
	if !m.collapsedProjects["myapp"] {
		t.Error("want collapsed state preserved after refresh")
	}
	filtered := m.filtered()
	if len(filtered) != 1 || filtered[0].State != "collapsed" {
		t.Error("want collapsed summary after refresh")
	}
}

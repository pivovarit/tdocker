package ui

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

func joined(lines []string) string { return strings.Join(lines, "\n") }

func TestBuildDiagnosticLines_AlwaysSections(t *testing.T) {
	c := docker.Container{ID: "abc", Names: "web", State: docker.StateRunning, Status: "Up 2 minutes"}
	data := &docker.InspectData{
		State:         docker.ContainerState{Status: "running", StartedAt: time.Now().Add(-2 * time.Minute)},
		RestartPolicy: docker.RestartPolicy{Name: "no"},
	}
	lines := buildDiagnosticLines(c, data, nil, nil, nil, 80, time.Now())
	out := joined(lines)
	for _, want := range []string{"Status", "Restarts", "Recent events", "Log tail", "Image", "Ports", "Mounts", "Networks", "Environment"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing section %q in output:\n%s", want, out)
		}
	}
}

func TestBuildDiagnosticLines_DiagnosisBlockHidden_WhenHealthy(t *testing.T) {
	c := docker.Container{ID: "abc", Names: "web", State: docker.StateRunning}
	data := &docker.InspectData{State: docker.ContainerState{Status: "running"}}
	lines := buildDiagnosticLines(c, data, nil, nil, nil, 80, time.Now())
	out := joined(lines)
	if strings.Contains(out, "Diagnosis") {
		t.Errorf("Diagnosis block should be hidden when healthy:\n%s", out)
	}
}

func TestBuildDiagnosticLines_DiagnosisBlockShown_WhenOOM(t *testing.T) {
	c := docker.Container{ID: "abc", Names: "web", State: "exited"}
	data := &docker.InspectData{
		State: docker.ContainerState{Status: "exited", ExitCode: 137, OOMKilled: true},
	}
	lines := buildDiagnosticLines(c, data, nil, nil, nil, 80, time.Now())
	out := joined(lines)
	if !strings.Contains(out, "Diagnosis") {
		t.Errorf("Diagnosis block should be shown when OOM:\n%s", out)
	}
	if !strings.Contains(out, "OOM") {
		t.Errorf("output should mention OOM:\n%s", out)
	}
}

func TestBuildDiagnosticLines_HealthSection_OnlyWhenHealthcheck(t *testing.T) {
	c := docker.Container{ID: "abc", Names: "web"}
	withHC := &docker.InspectData{
		State:       docker.ContainerState{Status: "running"},
		Healthcheck: &docker.Healthcheck{Test: []string{"CMD", "curl", "/"}, Interval: 30 * time.Second, Retries: 3},
	}
	linesHC := buildDiagnosticLines(c, withHC, nil, nil, nil, 80, time.Now())
	if !strings.Contains(joined(linesHC), "Health") {
		t.Errorf("Health section should appear when Healthcheck configured:\n%s", joined(linesHC))
	}
}

func TestBuildDiagnosticLines_EmptyEventsPlaceholder(t *testing.T) {
	c := docker.Container{ID: "abc", Names: "web"}
	data := &docker.InspectData{State: docker.ContainerState{Status: "running"}}
	lines := buildDiagnosticLines(c, data, nil, nil, nil, 80, time.Now())
	if !strings.Contains(joined(lines), "no events in last hour") {
		t.Errorf("expected empty-events placeholder:\n%s", joined(lines))
	}
}

func TestBuildDiagnosticLines_EmptyLogsPlaceholder(t *testing.T) {
	c := docker.Container{ID: "abc", Names: "web"}
	data := &docker.InspectData{State: docker.ContainerState{Status: "running"}}
	lines := buildDiagnosticLines(c, data, nil, nil, nil, 80, time.Now())
	if !strings.Contains(joined(lines), "no logs") {
		t.Errorf("expected empty-logs placeholder:\n%s", joined(lines))
	}
}

func TestBuildDiagnosticLines_EnvironmentLast(t *testing.T) {
	c := docker.Container{ID: "abc", Names: "web"}
	data := &docker.InspectData{
		State: docker.ContainerState{Status: "running"},
		Env:   []string{"FOO=bar"},
	}
	lines := buildDiagnosticLines(c, data, nil, nil, nil, 80, time.Now())
	out := joined(lines)

	// Use LastIndex for the Environment header to avoid matching an env var that happens
	// to contain the substring.
	envIdx := strings.LastIndex(out, "Environment")
	if envIdx < 0 {
		t.Fatalf("missing Environment section in output:\n%s", out)
	}

	for _, earlier := range []string{"Image", "Ports", "Mounts", "Networks"} {
		idx := strings.Index(out, earlier)
		if idx < 0 {
			t.Fatalf("missing section %q in output:\n%s", earlier, out)
		}
		if idx > envIdx {
			t.Errorf("section %q (at %d) should appear BEFORE Environment (at %d)", earlier, idx, envIdx)
		}
	}
}

func TestTruncateForWidth_RuneSafe(t *testing.T) {
	// Mix of multibyte runes. If byte-sliced at max-1, this would split a codepoint.
	in := "αβγδεζηθικλμνξοπ"
	out := truncateForWidth(in, 5)
	if !utf8.ValidString(out) {
		t.Errorf("truncateForWidth produced invalid UTF-8: %q", out)
	}
	// Also check it stays within the max rune count.
	if utf8.RuneCountInString(out) > 5 {
		t.Errorf("truncateForWidth exceeded max rune count: %d", utf8.RuneCountInString(out))
	}
	// And that it actually truncated.
	if utf8.RuneCountInString(out) == utf8.RuneCountInString(in) {
		t.Errorf("truncateForWidth did not truncate long input: %q", out)
	}
}

func TestDiagnosticPanel_RefreshKey(t *testing.T) {
	stub := newStubClient()
	app := newWithClient(stub, "test")
	app.diagnostic.visible = true
	app.diagnostic.containerID = "abc"
	app.diagnostic.container = "web"

	stub.lastInspectContainerID = ""
	stub.lastFetchContainerEventsID = ""
	stub.lastStartLogsID = ""

	_, cmd := app.handleDiagnosticKey(tea.KeyPressMsg{Code: 'r', Text: "r"})

	// Invoke the batch to trigger the individual commands
	if cmd != nil {
		if batchMsg, ok := cmd().(tea.BatchMsg); ok {
			for _, inner := range batchMsg {
				if inner != nil {
					inner()
				}
			}
		}
	}

	if stub.lastInspectContainerID != "abc" {
		t.Errorf("r should fire InspectContainer; got %q", stub.lastInspectContainerID)
	}
	if stub.lastFetchContainerEventsID != "abc" {
		t.Errorf("r should fire FetchContainerEvents; got %q", stub.lastFetchContainerEventsID)
	}
	if stub.lastStartLogsID != "abc" {
		t.Errorf("r should fire StartLogs; got %q", stub.lastStartLogsID)
	}
}

func TestDiagnosticPanel_OpenFiresEventsBackfill(t *testing.T) {
	stub := newStubClient()
	app := newWithClient(stub, "test")
	app.containers = []docker.Container{{ID: "abc", Names: "web", State: docker.StateRunning}}
	app.containersByID = indexContainers(app.containers)
	app.sorted = docker.Sort(app.containers)
	app = app.rebuildTable("")
	app.table.SetCursor(0)

	stub.lastFetchContainerEventsID = ""
	_, cmd := app.handleMainKey(tea.KeyPressMsg{Code: 'I', Text: "I"})

	// Invoke the batch to trigger the individual commands
	if cmd != nil {
		if batchMsg, ok := cmd().(tea.BatchMsg); ok {
			for _, inner := range batchMsg {
				if inner != nil {
					inner()
				}
			}
		}
	}

	if stub.lastFetchContainerEventsID != "abc" {
		t.Errorf("FetchContainerEvents not called; got %q", stub.lastFetchContainerEventsID)
	}
}

func TestDiagnosticPanel_LiveEventAppends(t *testing.T) {
	stub := newStubClient()
	app := newWithClient(stub, "test")
	app.diagnostic.visible = true
	app.diagnostic.containerID = "abc"
	app.diagnostic.container = "web"

	ev := docker.Event{Type: "container", Action: "die", Actor: docker.EventActor{ID: "abc"}, Time: time.Now().Unix()}
	msg := docker.EventLineMsg{Event: ev, Gen: app.bgEventsGen}

	next, _ := app.Update(msg)
	na := next.(App)
	if len(na.diagnostic.events) != 1 {
		t.Errorf("expected 1 event appended, got %d", len(na.diagnostic.events))
	}
}

func TestDiagnosticPanel_LiveEventOtherContainerIgnored(t *testing.T) {
	stub := newStubClient()
	app := newWithClient(stub, "test")
	app.diagnostic.visible = true
	app.diagnostic.containerID = "abc"

	ev := docker.Event{Type: "container", Action: "die", Actor: docker.EventActor{ID: "other"}, Time: time.Now().Unix()}
	msg := docker.EventLineMsg{Event: ev, Gen: app.bgEventsGen}

	next, _ := app.Update(msg)
	na := next.(App)
	if len(na.diagnostic.events) != 0 {
		t.Errorf("expected events untouched, got %d", len(na.diagnostic.events))
	}
}

func TestDiagnosticPanel_OpenStartsLogs(t *testing.T) {
	stub := newStubClient()
	app := newWithClient(stub, "test")
	app.containers = []docker.Container{{ID: "abc", Names: "web", State: docker.StateRunning}}
	app.containersByID = indexContainers(app.containers)
	app.sorted = docker.Sort(app.containers)
	app = app.rebuildTable("")
	app.table.SetCursor(0)

	stub.lastStartLogsID = ""
	_, cmd := app.handleMainKey(tea.KeyPressMsg{Code: 'I', Text: "I"})

	// Unwrap the batch to invoke all inner commands.
	if cmd != nil {
		if msg := cmd(); msg != nil {
			if batch, ok := msg.(tea.BatchMsg); ok {
				for _, inner := range batch {
					if inner != nil {
						inner()
					}
				}
			}
		}
	}

	if stub.lastStartLogsID != "abc" {
		t.Errorf("StartLogs not called with container ID; got %q", stub.lastStartLogsID)
	}
}

func TestDiagnosticPanel_LogLineAppends(t *testing.T) {
	stub := newStubClient()
	app := newWithClient(stub, "test")
	app.diagnostic.visible = true
	app.diagnostic.containerID = "abc"
	app.diagnostic.logsGen = 1

	msg := docker.LogsLineMsg{Line: "hello", Gen: 1}
	next, _ := app.Update(msg)
	na := next.(App)
	if len(na.diagnostic.logs) != 1 || na.diagnostic.logs[0] != "hello" {
		t.Errorf("logs = %v, want [hello]", na.diagnostic.logs)
	}
}

func TestDiagnosticPanel_LogLineWrongGenIgnored(t *testing.T) {
	stub := newStubClient()
	app := newWithClient(stub, "test")
	app.diagnostic.visible = true
	app.diagnostic.containerID = "abc"
	app.diagnostic.logsGen = 2

	msg := docker.LogsLineMsg{Line: "stale", Gen: 1}
	next, _ := app.Update(msg)
	na := next.(App)
	if len(na.diagnostic.logs) != 0 {
		t.Errorf("stale gen log should be ignored, got %v", na.diagnostic.logs)
	}
}

func TestDiagnosticPanel_LifecycleEventRefiresInspect(t *testing.T) {
	cases := []struct {
		action       string
		shouldRefire bool
	}{
		{"start", true},
		{"die", true},
		{"restart", true},
		{"oom", true},
		{"health_status", true},
		{"pause", false},
		{"create", false},
	}
	for _, c := range cases {
		t.Run(c.action, func(t *testing.T) {
			stub := newStubClient()
			app := newWithClient(stub, "test")
			app.diagnostic.visible = true
			app.diagnostic.containerID = "abc"

			stub.lastInspectContainerID = ""
			ev := docker.Event{Type: "container", Action: c.action, Actor: docker.EventActor{ID: "abc"}, Time: time.Now().Unix()}
			msg := docker.EventLineMsg{Event: ev, Gen: app.bgEventsGen}

			_, cmd := app.Update(msg)
			if cmd != nil {
				if out := cmd(); out != nil {
					if batch, ok := out.(tea.BatchMsg); ok {
						for _, inner := range batch {
							if inner != nil {
								inner()
							}
						}
					}
				}
			}

			refired := stub.lastInspectContainerID == "abc"
			if refired != c.shouldRefire {
				t.Errorf("action %q: refire = %v, want %v", c.action, refired, c.shouldRefire)
			}
		})
	}
}

func TestDiagnosticPanel_INoOpOnStateDetail(t *testing.T) {
	stub := newStubClient()
	app := newWithClient(stub, "test")
	app.containers = []docker.Container{{ID: "abc", Names: "(detail)", State: docker.StateDetail}}
	app.containersByID = indexContainers(app.containers)
	app.sorted = docker.Sort(app.containers)
	app = app.rebuildTable("")
	app.table.SetCursor(0)

	_, _ = app.handleMainKey(tea.KeyPressMsg{Code: 'I', Text: "I"})

	if app.diagnostic.visible {
		t.Error("diagnostic panel should not open on StateDetail row")
	}
}

func TestDiagnosticPanel_EndToEnd_RendersDiagnosisForExitedContainer(t *testing.T) {
	stub := newStubClient()
	stub.inspectData = &docker.InspectData{
		State: docker.ContainerState{
			Status:   "exited",
			ExitCode: 1,
			Error:    "boom",
		},
		RestartPolicy: docker.RestartPolicy{Name: "no"},
	}
	stub.events = []docker.Event{
		{Type: "container", Action: "die", Actor: docker.EventActor{ID: "abc"}, Time: time.Now().Unix()},
	}

	app := newWithClient(stub, "test")
	app.width = 120
	app.height = 40
	app.containers = []docker.Container{{ID: "abc", Names: "web", State: "exited", Status: "Exited (1) 30 seconds ago"}}
	app.containersByID = indexContainers(app.containers)
	app.sorted = docker.Sort(app.containers)
	app = app.rebuildTable("")
	app.table.SetCursor(0)

	// Open panel via Update.
	m1, cmd := app.Update(tea.KeyPressMsg{Code: 'I', Text: "I"})
	app = m1.(App)
	if !app.diagnostic.visible {
		t.Fatal("panel should be visible after i")
	}

	// Drain the batch cmd to collect response messages.
	var messages []tea.Msg
	if cmd != nil {
		if out := cmd(); out != nil {
			if batch, ok := out.(tea.BatchMsg); ok {
				for _, inner := range batch {
					if inner != nil {
						messages = append(messages, inner())
					}
				}
			} else {
				messages = append(messages, out)
			}
		}
	}

	// Pump responses through Update.
	for _, m := range messages {
		if m == nil {
			continue
		}
		m2, _ := app.Update(m)
		app = m2.(App)
	}

	out := app.renderDiagnosticPanel()

	if !strings.Contains(out, "exited") || !strings.Contains(out, "1") {
		t.Errorf("expected rendered panel to show exit status, got:\n%s", out)
	}
	if !strings.Contains(out, "Diagnosis") {
		t.Errorf("expected Diagnosis block in rendered panel, got:\n%s", out)
	}
	if !strings.Contains(out, "boom") {
		t.Errorf("expected error message 'boom' in rendered panel, got:\n%s", out)
	}
}

func TestDiagnosticPanel_DeleteClosesPanel(t *testing.T) {
	stub := newStubClient()
	app := newWithClient(stub, "test")
	app.diagnostic.visible = true
	app.diagnostic.containerID = "abc"
	app.diagnostic.container = "web"
	app.containers = []docker.Container{{ID: "abc", Names: "web"}}
	app.containersByID = indexContainers(app.containers)
	app.sorted = docker.Sort(app.containers)

	msg := docker.DeleteMsg{ID: "abc"}
	next, _ := app.Update(msg)
	na := next.(App)
	if na.diagnostic.visible {
		t.Error("diagnostic panel should close when its container is deleted")
	}
	if na.warnMsg == "" {
		t.Error("warnMsg should be set after container deletion")
	}
}

func TestDiagnosticPanel_ContextSwitchClosesPanel(t *testing.T) {
	stub := newStubClient()
	app := newWithClient(stub, "test")
	app.diagnostic.visible = true
	app.diagnostic.containerID = "abc"

	msg := docker.ContextSwitchMsg{}
	next, _ := app.Update(msg)
	na := next.(App)
	if na.diagnostic.visible {
		t.Error("diagnostic panel should close on context switch")
	}
}

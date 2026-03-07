//go:build integration

package ui

import (
	"context"
	stdlog "log"
	"strings"
	"testing"

	"github.com/pivovarit/tdocker/internal/docker"
	"github.com/testcontainers/testcontainers-go"
)

func startAlpine(t *testing.T, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, string) {
	t.Helper()
	ctx := context.Background()
	allOpts := append([]testcontainers.ContainerCustomizer{testcontainers.WithCmd("sh", "-c", "trap 'exit 0' TERM; sleep 60 & wait"), testcontainers.WithLogger(stdlog.Default())}, opts...)
	c, err := testcontainers.Run(ctx, "alpine", allOpts...)
	if err != nil {
		t.Fatalf("start container: %v", err)
	}
	testcontainers.CleanupContainer(t, c)
	return c, c.GetContainerID()
}

func appWithRealContainers(t *testing.T) App {
	t.Helper()
	msg := docker.CLI{}.FetchContainers(true)()
	containers, ok := msg.(docker.ContainersMsg)
	if !ok {
		t.Fatalf("expected ContainersMsg, got %T", msg)
	}
	m := New("")
	result, _ := m.Update(containers)
	return result.(App)
}

func findContainer(app App, fullID string) (docker.Container, bool) {
	for _, c := range app.filtered() {
		if strings.HasPrefix(fullID, c.ID) {
			return c, true
		}
	}
	return docker.Container{}, false
}

func TestIntegration_App_ShowsRunningContainer(t *testing.T) {
	t.Parallel()
	_, id := startAlpine(t)

	app := appWithRealContainers(t)

	c, found := findContainer(app, id)
	if !found {
		t.Fatalf("container %q not found in app", id[:12])
	}
	if c.State != "running" {
		t.Errorf("want State=running, got %q", c.State)
	}
}

func TestIntegration_App_FilterMatchesRunningContainer(t *testing.T) {
	t.Parallel()
	_, id := startAlpine(t)

	app := appWithRealContainers(t)
	app.filterQuery = id[:8]

	_, found := findContainer(app, id)
	if !found {
		t.Errorf("container %q not found when filtering by ID prefix %q", id[:12], id[:8])
	}
}

func TestIntegration_App_StopUpdatesState(t *testing.T) {
	t.Parallel()
	_, id := startAlpine(t)

	app := appWithRealContainers(t)
	if _, found := findContainer(app, id); !found {
		t.Fatalf("container not found before stop")
	}

	msg := docker.CLI{}.StopContainer(id)()
	result, _ := app.Update(msg)
	app = result.(App)

	app = appWithRealContainers(t)
	c, found := findContainer(app, id)
	if !found {
		t.Fatalf("container not found after stop (with showAll=true)")
	}
	if c.State == "running" {
		t.Errorf("want container stopped, still running")
	}
}

func TestIntegration_App_ComposeGroupingVisible(t *testing.T) {
	t.Parallel()
	_, id1 := startAlpine(t, testcontainers.WithLabels(map[string]string{
		"com.docker.compose.project": "integ-proj",
		"com.docker.compose.service": "svc-a",
	}))
	_, id2 := startAlpine(t, testcontainers.WithLabels(map[string]string{
		"com.docker.compose.project": "integ-proj",
		"com.docker.compose.service": "svc-b",
	}))

	app := appWithRealContainers(t)
	app.width = 160

	filtered := app.filtered()

	idx1, idx2 := -1, -1
	for i, c := range filtered {
		if strings.HasPrefix(id1, c.ID) {
			idx1 = i
		}
		if strings.HasPrefix(id2, c.ID) {
			idx2 = i
		}
	}
	if idx1 < 0 || idx2 < 0 {
		t.Fatalf("containers not found (idx1=%d, idx2=%d)", idx1, idx2)
	}

	diff := idx1 - idx2
	if diff < -1 || diff > 1 {
		t.Errorf("compose containers should be adjacent, got indices %d and %d", idx1, idx2)
	}

	rows := buildTable(filtered, 160).Rows()
	name1 := rows[idx1][1]
	name2 := rows[idx2][1]
	hasTree := func(s string) bool {
		return strings.ContainsAny(s, "┬├└")
	}
	if !hasTree(name1) || !hasTree(name2) {
		t.Errorf("expected tree chars in compose rows, got %q and %q", name1, name2)
	}
}

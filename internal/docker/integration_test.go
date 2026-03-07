//go:build integration

package docker

import (
	"context"
	stdlog "log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/testcontainers/testcontainers-go"
)

var (
	sharedAlpineID        string
	sharedLabeledAlpineID string
	sharedStoppedAlpineID string
	sharedPauseID         string
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	cmd := testcontainers.WithCmd("sh", "-c", "trap 'exit 0' TERM; sleep 60 & wait")
	logger := testcontainers.WithLogger(stdlog.Default())

	running, err := testcontainers.Run(ctx, "alpine", cmd, logger)
	if err != nil {
		stdlog.Fatalf("shared alpine: %v", err)
	}
	sharedAlpineID = running.GetContainerID()

	labeled, err := testcontainers.Run(ctx, "alpine", cmd, logger, testcontainers.WithLabels(map[string]string{
		"com.docker.compose.project": "myapp",
		"com.docker.compose.service": "web",
	}))
	if err != nil {
		stdlog.Fatalf("shared labeled alpine: %v", err)
	}
	sharedLabeledAlpineID = labeled.GetContainerID()

	stopped, err := testcontainers.Run(ctx, "alpine", cmd, logger)
	if err != nil {
		stdlog.Fatalf("shared stopped alpine: %v", err)
	}
	sharedStoppedAlpineID = stopped.GetContainerID()
	if err := stopped.Stop(ctx, nil); err != nil {
		stdlog.Fatalf("stop shared alpine: %v", err)
	}

	out, err := exec.Command("docker", "run", "-d", "registry.k8s.io/pause:3.10").Output()
	if err != nil {
		stdlog.Fatalf("shared pause: %v", err)
	}
	sharedPauseID = strings.TrimSpace(string(out))

	code := m.Run()

	running.Terminate(ctx)
	labeled.Terminate(ctx)
	stopped.Terminate(ctx)
	exec.Command("docker", "rm", "-f", sharedPauseID).Run()
	os.Exit(code)
}

func alpine(t *testing.T, extraArgs ...testcontainers.ContainerCustomizer) (testcontainers.Container, string) {
	t.Helper()
	ctx := context.Background()
	opts := append([]testcontainers.ContainerCustomizer{testcontainers.WithCmd("sh", "-c", "trap 'exit 0' TERM; sleep 60 & wait"), testcontainers.WithLogger(stdlog.Default())}, extraArgs...)
	c, err := testcontainers.Run(ctx, "alpine", opts...)
	if err != nil {
		t.Fatalf("start container: %v", err)
	}
	testcontainers.CleanupContainer(t, c)
	return c, c.GetContainerID()
}

func findByID(containers ContainersMsg, fullID string) (Container, bool) {
	for _, c := range containers {
		if strings.HasPrefix(fullID, c.ID) {
			return c, true
		}
	}
	return Container{}, false
}

func fetchAll(t *testing.T) ContainersMsg {
	t.Helper()
	msg := CLI{}.FetchContainers(true)()
	containers, ok := msg.(ContainersMsg)
	if !ok {
		t.Fatalf("expected ContainersMsg, got %T", msg)
	}
	return containers
}

func TestIntegration_FetchContainers_RunningContainerAppears(t *testing.T) {
	t.Parallel()
	containers := fetchAll(t)
	c, found := findByID(containers, sharedAlpineID)
	if !found {
		t.Fatalf("container %q not found", sharedAlpineID[:12])
	}
	if c.State != "running" {
		t.Errorf("want State=running, got %q", c.State)
	}
}

func TestIntegration_FetchContainers_StoppedHiddenWithoutAll(t *testing.T) {
	t.Parallel()
	msg := CLI{}.FetchContainers(false)()
	containers, ok := msg.(ContainersMsg)
	if !ok {
		t.Fatalf("expected ContainersMsg, got %T", msg)
	}
	if _, found := findByID(containers, sharedStoppedAlpineID); found {
		t.Errorf("stopped container %q should not appear without showAll", sharedStoppedAlpineID[:12])
	}
}

func TestIntegration_FetchContainers_ShowAllIncludesStopped(t *testing.T) {
	t.Parallel()
	containers := fetchAll(t)
	if _, found := findByID(containers, sharedStoppedAlpineID); !found {
		t.Errorf("stopped container %q not found with showAll=true", sharedStoppedAlpineID[:12])
	}
}

func TestIntegration_FetchContainers_LabelsAreParsed(t *testing.T) {
	t.Parallel()
	containers := fetchAll(t)
	c, found := findByID(containers, sharedLabeledAlpineID)
	if !found {
		t.Fatalf("container %q not found", sharedLabeledAlpineID[:12])
	}
	if got := c.ComposeProject(); got != "myapp" {
		t.Errorf("want ComposeProject=%q, got %q", "myapp", got)
	}
	if got := c.ComposeService(); got != "web" {
		t.Errorf("want ComposeService=%q, got %q", "web", got)
	}
}

func TestIntegration_CheckShellAvailable_NoShell(t *testing.T) {
	t.Parallel()
	msg := CLI{}.CheckShellAvailable(sharedPauseID)()

	result, ok := msg.(ShellAvailableMsg)
	if !ok {
		t.Fatalf("want ShellAvailableMsg, got %T", msg)
	}
	if result.ID != sharedPauseID {
		t.Errorf("want ID=%q, got %q", sharedPauseID, result.ID)
	}
	if result.Available {
		t.Error("want Available=false for shell-less container")
	}
}

func TestIntegration_CheckShellAvailable_WithShell(t *testing.T) {
	t.Parallel()
	msg := CLI{}.CheckShellAvailable(sharedAlpineID)()

	result, ok := msg.(ShellAvailableMsg)
	if !ok {
		t.Fatalf("want ShellAvailableMsg, got %T", msg)
	}
	if result.ID != sharedAlpineID {
		t.Errorf("want ID=%q, got %q", sharedAlpineID, result.ID)
	}
	if !result.Available {
		t.Error("want Available=true for alpine container")
	}
}

func TestIntegration_StopContainer(t *testing.T) {
	t.Parallel()
	_, id := alpine(t)
	ctx := context.Background()

	msg := CLI{}.StopContainer(id)()
	stopMsg, ok := msg.(StopMsg)
	if !ok {
		t.Fatalf("expected StopMsg, got %T", msg)
	}
	if stopMsg.Err != nil {
		t.Fatalf("unexpected error: %v", stopMsg.Err)
	}

	containers := fetchAll(t)
	c, found := findByID(containers, id)
	if !found {
		t.Fatalf("container %q not found after stop", id[:12])
	}
	_ = ctx
	if c.State == "running" {
		t.Errorf("want container stopped, got State=%q", c.State)
	}
}

func TestIntegration_StartContainer(t *testing.T) {
	t.Parallel()
	ctr, id := alpine(t)
	ctx := context.Background()
	if err := ctr.Stop(ctx, nil); err != nil {
		t.Fatalf("stop: %v", err)
	}

	msg := CLI{}.StartContainer(id)()
	startMsg, ok := msg.(StartMsg)
	if !ok {
		t.Fatalf("expected StartMsg, got %T", msg)
	}
	if startMsg.Err != nil {
		t.Fatalf("unexpected error: %v", startMsg.Err)
	}

	containers := fetchAll(t)
	c, found := findByID(containers, id)
	if !found {
		t.Fatalf("container %q not found after start", id[:12])
	}
	if c.State != "running" {
		t.Errorf("want State=running, got %q", c.State)
	}
}

func TestIntegration_DeleteContainer(t *testing.T) {
	t.Parallel()
	ctr, id := alpine(t)
	ctx := context.Background()
	if err := ctr.Stop(ctx, nil); err != nil {
		t.Fatalf("stop: %v", err)
	}

	msg := CLI{}.DeleteContainer(id)()
	deleteMsg, ok := msg.(DeleteMsg)
	if !ok {
		t.Fatalf("expected DeleteMsg, got %T", msg)
	}
	if deleteMsg.Err != nil {
		t.Fatalf("unexpected error: %v", deleteMsg.Err)
	}
	if deleteMsg.ID != id {
		t.Errorf("want ID=%q, got %q", id, deleteMsg.ID)
	}

	containers := fetchAll(t)
	if _, found := findByID(containers, id); found {
		t.Errorf("container %q should be deleted but still appears", id[:12])
	}
}

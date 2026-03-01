package ui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pivovarit/tdocker/internal/docker"
)

var (
	runningContainer = docker.Container{ID: "run111", Names: "web", Image: "nginx", State: "running"}
	stoppedContainer = docker.Container{ID: "stop222", Names: "db", Image: "postgres", State: "exited"}
)

func TestUpdate_ShiftSOnRunningEntersStopConfirm(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	got := update(m, runeKey("S"))
	if got.op != OpConfirming {
		t.Fatal("want op=OpConfirming")
	}
	if got.confirmAction != "stop" {
		t.Errorf("want confirmAction=%q, got %q", "stop", got.confirmAction)
	}
	if got.confirmID != runningContainer.ID {
		t.Errorf("want confirmID=%q, got %q", runningContainer.ID, got.confirmID)
	}
}

func TestUpdate_ShiftSOnStoppedDoesNothing(t *testing.T) {
	m := modelWithSorted([]docker.Container{stoppedContainer})
	got := update(m, runeKey("S"))
	if got.op == OpConfirming {
		t.Error("want op!=OpConfirming for non-running container")
	}
}

func TestUpdate_SKeyOnStoppedEntersStartConfirm(t *testing.T) {
	m := modelWithSorted([]docker.Container{stoppedContainer})
	got := update(m, runeKey("s"))
	if got.op != OpConfirming {
		t.Fatal("want op=OpConfirming")
	}
	if got.confirmAction != "start" {
		t.Errorf("want confirmAction=%q, got %q", "start", got.confirmAction)
	}
	if got.confirmID != stoppedContainer.ID {
		t.Errorf("want confirmID=%q, got %q", stoppedContainer.ID, got.confirmID)
	}
}

func TestUpdate_SKeyOnRunningDoesNothing(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	got := update(m, runeKey("s"))
	if got.op == OpConfirming {
		t.Error("want op!=OpConfirming for running container")
	}
}

func TestUpdate_DKeyOnStoppedEntersDeleteConfirm(t *testing.T) {
	m := modelWithSorted([]docker.Container{stoppedContainer})
	got := update(m, runeKey("D"))
	if got.op != OpConfirming {
		t.Fatal("want op=OpConfirming")
	}
	if got.confirmAction != "delete" {
		t.Errorf("want confirmAction=%q, got %q", "delete", got.confirmAction)
	}
	if got.confirmID != stoppedContainer.ID {
		t.Errorf("want confirmID=%q, got %q", stoppedContainer.ID, got.confirmID)
	}
}

func TestUpdate_DKeyOnRunningDoesNothing(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	got := update(m, runeKey("D"))
	if got.op == OpConfirming {
		t.Error("want op!=OpConfirming for running container")
	}
}

func TestUpdate_ConfirmYSetsStoppingFlag(t *testing.T) {
	m := confirming("stop", runningContainer)
	got := update(m, runeKey("y"))
	if got.op == OpConfirming {
		t.Error("want op!=OpConfirming after y")
	}
	if got.op != OpStopping {
		t.Error("want op=OpStopping")
	}
}

func TestUpdate_ConfirmYSetsStartingFlag(t *testing.T) {
	m := confirming("start", stoppedContainer)
	got := update(m, runeKey("y"))
	if got.op == OpConfirming {
		t.Error("want op!=OpConfirming after y")
	}
	if got.op != OpStarting {
		t.Error("want op=OpStarting")
	}
}

func TestUpdate_ConfirmYSetsDeletingFlag(t *testing.T) {
	m := confirming("delete", stoppedContainer)
	got := update(m, runeKey("y"))
	if got.op == OpConfirming {
		t.Error("want op!=OpConfirming after y")
	}
	if got.op != OpDeleting {
		t.Error("want op=OpDeleting")
	}
}

func TestUpdate_ConfirmCancelKeys(t *testing.T) {
	cancelKeys := []tea.KeyMsg{
		runeKey("n"),
		runeKey("N"),
		{Type: tea.KeyEsc},
	}
	for _, key := range cancelKeys {
		m := confirming("stop", runningContainer)
		got := update(m, key)
		if got.op == OpConfirming {
			t.Errorf("key %v: want op!=OpConfirming", key)
		}
		if got.op == OpStopping {
			t.Errorf("key %v: want op!=OpStopping", key)
		}
	}
}

func TestUpdate_DeleteMsgRemovesContainerLocally(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer, stoppedContainer})
	got := update(m, docker.DeleteMsg{ID: stoppedContainer.ID})
	if got.op == OpDeleting {
		t.Error("want op!=OpDeleting")
	}
	if got.loading {
		t.Error("want no reload — smooth deletion should update locally")
	}
	if len(got.containers) != 1 {
		t.Fatalf("want 1 container remaining, got %d", len(got.containers))
	}
	if got.containers[0].ID != runningContainer.ID {
		t.Errorf("want remaining ID=%q, got %q", runningContainer.ID, got.containers[0].ID)
	}
}

func TestUpdate_DeleteMsgUnknownIDIsNoop(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer, stoppedContainer})
	got := update(m, docker.DeleteMsg{ID: "unknown"})
	if len(got.containers) != 2 {
		t.Errorf("want 2 containers unchanged, got %d", len(got.containers))
	}
}

func TestUpdate_DeleteMsgErrorPreservesContainers(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer, stoppedContainer})
	got := update(m, docker.DeleteMsg{ID: stoppedContainer.ID, Err: errors.New("denied")})
	if got.err == nil {
		t.Error("want err set")
	}
	if len(got.containers) != 2 {
		t.Errorf("want containers unchanged on error, got %d", len(got.containers))
	}
}

func TestUpdate_DeleteAllContainersLeavesEmptyList(t *testing.T) {
	m := modelWithSorted([]docker.Container{stoppedContainer})
	got := update(m, docker.DeleteMsg{ID: stoppedContainer.ID})
	if len(got.containers) != 0 {
		t.Errorf("want 0 containers, got %d", len(got.containers))
	}
}

func TestUpdate_StartMsgTriggersReload(t *testing.T) {
	m := modelWithSorted([]docker.Container{stoppedContainer})
	m.op = OpStarting
	result, cmd := m.Update(docker.StartMsg{})
	got := result.(App)
	if got.op == OpStarting {
		t.Error("want op!=OpStarting")
	}
	if !got.loading {
		t.Error("want loading=true to trigger reload")
	}
	if cmd == nil {
		t.Error("want non-nil fetch cmd")
	}
}

func TestUpdate_StartMsgErrorSetsErr(t *testing.T) {
	m := modelWithSorted([]docker.Container{stoppedContainer})
	m.op = OpStarting
	got, cmd := m.Update(docker.StartMsg{Err: errors.New("no such container")})
	if got.(App).err == nil {
		t.Error("want err set")
	}
	if got.(App).loading {
		t.Error("want no reload on error")
	}
	if cmd != nil {
		t.Error("want nil cmd on error")
	}
}

func TestUpdate_StopMsgTriggersReload(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.op = OpStopping
	result, cmd := m.Update(docker.StopMsg{})
	got := result.(App)
	if got.op == OpStopping {
		t.Error("want op!=OpStopping")
	}
	if !got.loading {
		t.Error("want loading=true to trigger reload")
	}
	if cmd == nil {
		t.Error("want non-nil fetch cmd")
	}
}

func TestUpdate_StopMsgErrorSetsErr(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.op = OpStopping
	got, _ := m.Update(docker.StopMsg{Err: errors.New("failed")})
	if got.(App).err == nil {
		t.Error("want err set")
	}
	if got.(App).loading {
		t.Error("want no reload on error")
	}
}

func update(m App, msg tea.Msg) App {
	result, _ := m.Update(msg)
	return result.(App)
}

func confirming(action string, c docker.Container) App {
	m := modelWithSorted([]docker.Container{c})
	m.op = OpConfirming
	m.confirmAction = action
	m.confirmID = c.ID
	m.confirmName = c.Names
	return m
}

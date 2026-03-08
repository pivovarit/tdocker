package ui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

var (
	runningContainer = docker.Container{ID: "run111", Names: "web", Image: "nginx", State: "running"}
	stoppedContainer = docker.Container{ID: "stop222", Names: "db", Image: "postgres", State: "exited"}
)

func TestUpdate_ShiftSOnRunningEntersStopConfirm(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	got := update(m, runeKey("S"))
	if got.op.kind != OpConfirming {
		t.Fatal("want op=OpConfirming")
	}
	if got.op.action != "stop" {
		t.Errorf("want confirmAction=%q, got %q", "stop", got.op.action)
	}
	if got.op.id != runningContainer.ID {
		t.Errorf("want confirmID=%q, got %q", runningContainer.ID, got.op.id)
	}
}

func TestUpdate_ShiftSOnStoppedEntersStartConfirm(t *testing.T) {
	m := modelWithSorted([]docker.Container{stoppedContainer})
	got := update(m, runeKey("S"))
	if got.op.kind != OpConfirming {
		t.Fatal("want op=OpConfirming")
	}
	if got.op.action != "start" {
		t.Errorf("want confirmAction=%q, got %q", "start", got.op.action)
	}
	if got.op.id != stoppedContainer.ID {
		t.Errorf("want confirmID=%q, got %q", stoppedContainer.ID, got.op.id)
	}
}

func TestUpdate_DKeyOnStoppedEntersDeleteConfirm(t *testing.T) {
	m := modelWithSorted([]docker.Container{stoppedContainer})
	got := update(m, runeKey("D"))
	if got.op.kind != OpConfirming {
		t.Fatal("want op=OpConfirming")
	}
	if got.op.action != "delete" {
		t.Errorf("want confirmAction=%q, got %q", "delete", got.op.action)
	}
	if got.op.id != stoppedContainer.ID {
		t.Errorf("want confirmID=%q, got %q", stoppedContainer.ID, got.op.id)
	}
}

func TestUpdate_DKeyOnRunningDoesNothing(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	got := update(m, runeKey("D"))
	if got.op.kind == OpConfirming {
		t.Error("want op!=OpConfirming for running container")
	}
}

func TestUpdate_ConfirmYSetsStoppingFlag(t *testing.T) {
	m := confirming("stop", runningContainer)
	got := update(m, runeKey("y"))
	if got.op.kind == OpConfirming {
		t.Error("want op!=OpConfirming after y")
	}
	if got.op.kind != OpStopping {
		t.Error("want op=OpStopping")
	}
}

func TestUpdate_ConfirmYSetsStartingFlag(t *testing.T) {
	m := confirming("start", stoppedContainer)
	got := update(m, runeKey("y"))
	if got.op.kind == OpConfirming {
		t.Error("want op!=OpConfirming after y")
	}
	if got.op.kind != OpStarting {
		t.Error("want op=OpStarting")
	}
}

func TestUpdate_ConfirmYSetsDeletingFlag(t *testing.T) {
	m := confirming("delete", stoppedContainer)
	got := update(m, runeKey("y"))
	if got.op.kind == OpConfirming {
		t.Error("want op!=OpConfirming after y")
	}
	if got.op.kind != OpDeleting {
		t.Error("want op=OpDeleting")
	}
}

func TestUpdate_ConfirmCancelKeys(t *testing.T) {
	cancelKeys := []tea.Msg{
		runeKey("n"),
		runeKey("N"),
		tea.KeyPressMsg{Code: tea.KeyEsc},
	}
	for _, key := range cancelKeys {
		m := confirming("stop", runningContainer)
		got := update(m, key)
		if got.op.kind == OpConfirming {
			t.Errorf("key %v: want op!=OpConfirming", key)
		}
		if got.op.kind == OpStopping {
			t.Errorf("key %v: want op!=OpStopping", key)
		}
	}
}

func TestUpdate_DeleteMsgRemovesContainerLocally(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer, stoppedContainer})
	got := update(m, docker.DeleteMsg{ID: stoppedContainer.ID})
	if got.op.kind == OpDeleting {
		t.Error("want op!=OpDeleting")
	}
	if got.fetch.loading {
		t.Error("want no reload - smooth deletion should update locally")
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
	m.op.kind = OpStarting
	result, cmd := m.Update(docker.StartMsg{})
	got := result.(App)
	if got.op.kind == OpStarting {
		t.Error("want op!=OpStarting")
	}
	if !got.fetch.loading {
		t.Error("want loading=true to trigger reload")
	}
	if cmd == nil {
		t.Error("want non-nil fetch cmd")
	}
}

func TestUpdate_StartMsgErrorSetsErr(t *testing.T) {
	m := modelWithSorted([]docker.Container{stoppedContainer})
	m.op.kind = OpStarting
	got, cmd := m.Update(docker.StartMsg{Err: errors.New("no such container")})
	if got.(App).err == nil {
		t.Error("want err set")
	}
	if got.(App).fetch.loading {
		t.Error("want no reload on error")
	}
	if cmd != nil {
		t.Error("want nil cmd on error")
	}
}

func TestUpdate_StopMsgTriggersReload(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.op.kind = OpStopping
	result, cmd := m.Update(docker.StopMsg{})
	got := result.(App)
	if got.op.kind == OpStopping {
		t.Error("want op!=OpStopping")
	}
	if !got.fetch.loading {
		t.Error("want loading=true to trigger reload")
	}
	if cmd == nil {
		t.Error("want non-nil fetch cmd")
	}
}

func TestUpdate_StopMsgErrorSetsErr(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.op.kind = OpStopping
	got, _ := m.Update(docker.StopMsg{Err: errors.New("failed")})
	if got.(App).err == nil {
		t.Error("want err set")
	}
	if got.(App).fetch.loading {
		t.Error("want no reload on error")
	}
}

func TestUpdate_ExecDoneMsgWithErrorSetsErr(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	got, cmd := m.Update(docker.ExecDoneMsg{Err: errors.New("shell not found")})
	if got.(App).err == nil {
		t.Error("want err surfaced in app")
	}
	if got.(App).fetch.loading {
		t.Error("want no loading triggered on exec error")
	}
	if cmd != nil {
		t.Error("want nil cmd on exec error")
	}
}

func TestUpdate_ExecDoneMsgSuccessTriggersReload(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	got, cmd := m.Update(docker.ExecDoneMsg{})
	if got.(App).err != nil {
		t.Errorf("want no error on success, got %v", got.(App).err)
	}
	if !got.(App).fetch.loading {
		t.Error("want loading=true after successful exec")
	}
	if cmd == nil {
		t.Error("want non-nil fetch cmd after successful exec")
	}
}

func update(m App, msg tea.Msg) App {
	result, _ := m.Update(msg)
	return result.(App)
}

func confirming(action string, c docker.Container) App {
	m := modelWithSorted([]docker.Container{c})
	m.op = operationState{kind: OpConfirming, action: action, id: c.ID, name: c.Names}
	return m
}

package ui

import (
	"testing"

	"github.com/pivovarit/tdocker/internal/docker"
)

var (
	cA = docker.Container{ID: "aaa", Names: "alpha", State: "running"}
	cB = docker.Container{ID: "bbb", Names: "beta", State: "running"}
	cC = docker.Container{ID: "ccc", Names: "gamma", State: "running"}
)

func TestSelection_PreservedOnContainersMsg(t *testing.T) {
	m := modelWithSorted([]docker.Container{cA, cB, cC})
	m.table.SetCursor(1)

	m = update(m, docker.ContainersMsg{cA, cB, cC})
	if got := m.table.Cursor(); got != 1 {
		t.Errorf("want cursor=1 (cB), got %d", got)
	}
}

func TestSelection_CursorFollowsContainerWhenNewOneInsertsBefore(t *testing.T) {
	m := modelWithSorted([]docker.Container{cB, cC})
	m.table.SetCursor(0)

	m = update(m, docker.ContainersMsg{cA, cB, cC})
	if got := m.table.Cursor(); got != 1 {
		t.Errorf("want cursor=1 (cB shifted after cA inserted before it), got %d", got)
	}
}

func TestSelection_ResetWhenSelectedContainerDisappears(t *testing.T) {
	m := modelWithSorted([]docker.Container{cA, cB, cC})
	m.table.SetCursor(1)

	m = update(m, docker.ContainersMsg{cA, cC})
	if got := m.table.Cursor(); got != 0 {
		t.Errorf("want cursor reset to 0 when selected container is gone, got %d", got)
	}
}

func TestSelection_PreservedAfterDelete(t *testing.T) {
	m := modelWithSorted([]docker.Container{cA, cB, cC})
	m.table.SetCursor(2)

	m = update(m, docker.DeleteMsg{ID: cB.ID})
	if got := m.table.Cursor(); got != 1 {
		t.Errorf("want cursor=1 (cC shifted after cB deleted), got %d", got)
	}
}

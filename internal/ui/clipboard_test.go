package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pivovarit/tdocker/internal/docker"
)

func TestUpdate_CKeySetsCopiedName(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	got, cmd := m.Update(runeKey("c"))
	if got.(Model).copiedName != runningContainer.Names {
		t.Errorf("want copiedName=%q, got %q", runningContainer.Names, got.(Model).copiedName)
	}
	if cmd == nil {
		t.Error("want non-nil clipboard cmd")
	}
}

func TestUpdate_CKeyOnStoppedContainerSetsCopiedName(t *testing.T) {
	m := modelWithSorted([]docker.Container{stoppedContainer})
	got := update(m, runeKey("c"))
	if got.copiedName != stoppedContainer.Names {
		t.Errorf("want copiedName=%q, got %q", stoppedContainer.Names, got.copiedName)
	}
}

func TestUpdate_CKeyOnEmptyListDoesNothing(t *testing.T) {
	m := modelWithSorted(nil)
	got := update(m, runeKey("c"))
	if got.copiedName != "" {
		t.Errorf("want copiedName empty, got %q", got.copiedName)
	}
}

func TestUpdate_AnyKeyClearsCopiedName(t *testing.T) {
	keys := []tea.Msg{
		runeKey("r"),
		runeKey("a"),
		runeKey("/"),
		tea.KeyMsg{Type: tea.KeyEsc},
		tea.KeyMsg{Type: tea.KeyUp},
	}
	for _, key := range keys {
		m := modelWithSorted([]docker.Container{runningContainer})
		m.copiedName = runningContainer.Names
		got := update(m, key)
		if got.copiedName != "" {
			t.Errorf("key %v: want copiedName cleared, got %q", key, got.copiedName)
		}
	}
}

func TestUpdate_CKeyAfterCKeyClears(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	m.copiedName = runningContainer.Names
	got := update(m, runeKey("c"))
	if got.copiedName != runningContainer.Names {
		t.Errorf("want copiedName reset to %q, got %q", runningContainer.Names, got.copiedName)
	}
}

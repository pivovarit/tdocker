package ui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pivovarit/tdocker/internal/docker"
)

func TestUpdate_CKeyDispatchesClipboardCmd(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	got, cmd := m.Update(runeKey("c"))
	if got.(App).copiedName != "" {
		t.Error("want copiedName empty until clipboardMsg arrives")
	}
	if cmd == nil {
		t.Error("want non-nil clipboard cmd")
	}
}

func TestUpdate_ClipboardMsgSetsCopiedName(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	got := update(m, clipboardMsg{name: runningContainer.Names})
	if got.copiedName != runningContainer.Names {
		t.Errorf("want copiedName=%q, got %q", runningContainer.Names, got.copiedName)
	}
}

func TestUpdate_ClipboardMsgErrorSetsErr(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	got := update(m, clipboardMsg{err: fmt.Errorf("pbcopy: command not found")})
	if got.err == nil {
		t.Error("want err set on clipboard failure")
	}
	if got.copiedName != "" {
		t.Error("want copiedName empty on failure")
	}
}

func TestUpdate_CKeyOnEmptyListDoesNothing(t *testing.T) {
	m := modelWithSorted(nil)
	got, cmd := m.Update(runeKey("c"))
	if got.(App).copiedName != "" {
		t.Errorf("want copiedName empty, got %q", got.(App).copiedName)
	}
	if cmd != nil {
		t.Error("want nil cmd when no container selected")
	}
}

func TestClipExec_SuccessReturnsMsgWithName(t *testing.T) {
	cmd := clipExec("web", "abc123", "cat")
	msg := cmd()
	got, ok := msg.(clipboardMsg)
	if !ok {
		t.Fatalf("want clipboardMsg, got %T", msg)
	}
	if got.name != "web" {
		t.Errorf("want name=%q, got %q", "web", got.name)
	}
	if got.err != nil {
		t.Errorf("want nil error, got %v", got.err)
	}
}

func TestClipExec_FailureReturnsMsgWithError(t *testing.T) {
	cmd := clipExec("web", "abc123", "__nonexistent_binary__")
	msg := cmd()
	got, ok := msg.(clipboardMsg)
	if !ok {
		t.Fatalf("want clipboardMsg, got %T", msg)
	}
	if got.err == nil {
		t.Error("want error when command not found")
	}
	if got.name != "" {
		t.Errorf("want empty name on failure, got %q", got.name)
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

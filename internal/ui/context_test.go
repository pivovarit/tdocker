package ui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pivovarit/tdocker/internal/docker"
)

var testContexts = []docker.DockerContext{
	{Name: "default", Current: true, Description: "Current DOCKER_HOST based"},
	{Name: "remote", Current: false, Description: "Remote server"},
}

func TestContextPicker_XKey_CallsFetchContexts(t *testing.T) {
	mc := newStubClient()
	fetched := false
	mc.fetchContexts = func() tea.Cmd {
		fetched = true
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, nil)
	update(m, runeKey("X"))
	if !fetched {
		t.Error("want FetchContexts called on X key")
	}
}

func TestContextPicker_ContextsMsg_SetsCurrentContext(t *testing.T) {
	m := modelWithSorted(nil)
	got := update(m, docker.ContextsMsg(testContexts))
	if got.currentContext != "default" {
		t.Errorf("want currentContext=%q, got %q", "default", got.currentContext)
	}
	if got.contextPickerVisible {
		t.Error("want contextPickerVisible=false when not requested")
	}
}

func TestContextPicker_ContextsMsg_OpensPicker(t *testing.T) {
	m := modelWithSorted(nil)
	m.contextPickerRequested = true
	got := update(m, docker.ContextsMsg(testContexts))
	if !got.contextPickerVisible {
		t.Error("want contextPickerVisible=true after ContextsMsg")
	}
}

func TestContextPicker_ContextsMsg_CursorOnCurrentContext(t *testing.T) {
	contexts := []docker.DockerContext{
		{Name: "other", Current: false},
		{Name: "active", Current: true},
	}
	m := modelWithSorted(nil)
	m.contextPickerRequested = true
	got := update(m, docker.ContextsMsg(contexts))
	if got.contextCursor != 1 {
		t.Errorf("want contextCursor=1 (active context), got %d", got.contextCursor)
	}
}

func TestContextPicker_EscClosesPicker(t *testing.T) {
	m := modelWithSorted(nil)
	m.contextPickerVisible = true
	m.contexts = testContexts
	got := update(m, tea.KeyMsg{Type: tea.KeyEsc})
	if got.contextPickerVisible {
		t.Error("want contextPickerVisible=false after esc")
	}
	if got.contexts != nil {
		t.Error("want contexts cleared after esc")
	}
}

func TestContextPicker_JKey_MovesCursorDown(t *testing.T) {
	m := modelWithSorted(nil)
	m.contextPickerVisible = true
	m.contexts = testContexts
	m.contextCursor = 0
	got := update(m, runeKey("j"))
	if got.contextCursor != 1 {
		t.Errorf("want contextCursor=1 after j, got %d", got.contextCursor)
	}
}

func TestContextPicker_DownKey_MovesCursorDown(t *testing.T) {
	m := modelWithSorted(nil)
	m.contextPickerVisible = true
	m.contexts = testContexts
	m.contextCursor = 0
	got := update(m, tea.KeyMsg{Type: tea.KeyDown})
	if got.contextCursor != 1 {
		t.Errorf("want contextCursor=1 after down, got %d", got.contextCursor)
	}
}

func TestContextPicker_KKey_MovesCursorUp(t *testing.T) {
	m := modelWithSorted(nil)
	m.contextPickerVisible = true
	m.contexts = testContexts
	m.contextCursor = 1
	got := update(m, runeKey("k"))
	if got.contextCursor != 0 {
		t.Errorf("want contextCursor=0 after k, got %d", got.contextCursor)
	}
}

func TestContextPicker_CursorDoesNotUnderflow(t *testing.T) {
	m := modelWithSorted(nil)
	m.contextPickerVisible = true
	m.contexts = testContexts
	m.contextCursor = 0
	got := update(m, runeKey("k"))
	if got.contextCursor != 0 {
		t.Errorf("want contextCursor=0 (no underflow), got %d", got.contextCursor)
	}
}

func TestContextPicker_CursorDoesNotOverflow(t *testing.T) {
	m := modelWithSorted(nil)
	m.contextPickerVisible = true
	m.contexts = testContexts
	m.contextCursor = len(testContexts) - 1
	got := update(m, runeKey("j"))
	if got.contextCursor != len(testContexts)-1 {
		t.Errorf("want contextCursor=%d (no overflow), got %d", len(testContexts)-1, got.contextCursor)
	}
}

func TestContextPicker_Enter_CallsSwitchContext(t *testing.T) {
	mc := newStubClient()
	var gotName string
	mc.switchContext = func(name string) tea.Cmd {
		gotName = name
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, nil)
	m.contextPickerVisible = true
	m.contexts = testContexts
	m.contextCursor = 1
	update(m, tea.KeyMsg{Type: tea.KeyEnter})
	if gotName != "remote" {
		t.Errorf("want SwitchContext(%q), got %q", "remote", gotName)
	}
}

func TestContextPicker_SwitchMsg_ClosesAndRefreshes(t *testing.T) {
	mc := newStubClient()
	mc.fetchContainers = func(_ bool) tea.Cmd {
		return func() tea.Msg { return nil }
	}
	m := modelWithMock(mc, nil)
	m.contextPickerVisible = true
	m.contexts = testContexts
	got, cmd := m.Update(docker.ContextSwitchMsg{})
	app := got.(App)
	if app.contextPickerVisible {
		t.Error("want contextPickerVisible=false after switch")
	}
	if !app.loading {
		t.Error("want loading=true after switch")
	}
	if cmd == nil {
		t.Error("want non-nil fetch cmd after switch")
	}
}

func TestContextPicker_SwitchMsg_ErrorSetsErr(t *testing.T) {
	m := modelWithSorted(nil)
	m.contextPickerVisible = true
	m.contexts = testContexts
	got := update(m, docker.ContextSwitchMsg{Err: errors.New("permission denied")})
	if got.err == nil {
		t.Error("want err set on switch failure")
	}
	if got.contextPickerVisible {
		t.Error("want contextPickerVisible=false even on error")
	}
}

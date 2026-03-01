package ui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
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
	if got.ctxPicker.current != "default" {
		t.Errorf("want ctxPicker.current=%q, got %q", "default", got.ctxPicker.current)
	}
	if got.ctxPicker.visible {
		t.Error("want ctxPicker.visible=false when not requested")
	}
}

func TestContextPicker_ContextsMsg_OpensPicker(t *testing.T) {
	m := modelWithSorted(nil)
	m.ctxPicker.requested = true
	got := update(m, docker.ContextsMsg(testContexts))
	if !got.ctxPicker.visible {
		t.Error("want ctxPicker.visible=true after ContextsMsg")
	}
}

func TestContextPicker_ContextsMsg_CursorOnCurrentContext(t *testing.T) {
	contexts := []docker.DockerContext{
		{Name: "other", Current: false},
		{Name: "active", Current: true},
	}
	m := modelWithSorted(nil)
	m.ctxPicker.requested = true
	got := update(m, docker.ContextsMsg(contexts))
	if got.ctxPicker.cursor != 1 {
		t.Errorf("want ctxPicker.cursor=1 (active context), got %d", got.ctxPicker.cursor)
	}
}

func TestContextPicker_EscClosesPicker(t *testing.T) {
	m := modelWithSorted(nil)
	m.ctxPicker.visible = true
	m.ctxPicker.contexts = testContexts
	got := update(m, tea.KeyPressMsg{Code: tea.KeyEsc})
	if got.ctxPicker.visible {
		t.Error("want ctxPicker.visible=false after esc")
	}
	if got.ctxPicker.contexts != nil {
		t.Error("want ctxPicker.contexts cleared after esc")
	}
}

func TestContextPicker_JKey_MovesCursorDown(t *testing.T) {
	m := modelWithSorted(nil)
	m.ctxPicker.visible = true
	m.ctxPicker.contexts = testContexts
	m.ctxPicker.cursor = 0
	got := update(m, runeKey("j"))
	if got.ctxPicker.cursor != 1 {
		t.Errorf("want ctxPicker.cursor=1 after j, got %d", got.ctxPicker.cursor)
	}
}

func TestContextPicker_DownKey_MovesCursorDown(t *testing.T) {
	m := modelWithSorted(nil)
	m.ctxPicker.visible = true
	m.ctxPicker.contexts = testContexts
	m.ctxPicker.cursor = 0
	got := update(m, tea.KeyPressMsg{Code: tea.KeyDown})
	if got.ctxPicker.cursor != 1 {
		t.Errorf("want ctxPicker.cursor=1 after down, got %d", got.ctxPicker.cursor)
	}
}

func TestContextPicker_KKey_MovesCursorUp(t *testing.T) {
	m := modelWithSorted(nil)
	m.ctxPicker.visible = true
	m.ctxPicker.contexts = testContexts
	m.ctxPicker.cursor = 1
	got := update(m, runeKey("k"))
	if got.ctxPicker.cursor != 0 {
		t.Errorf("want ctxPicker.cursor=0 after k, got %d", got.ctxPicker.cursor)
	}
}

func TestContextPicker_CursorDoesNotUnderflow(t *testing.T) {
	m := modelWithSorted(nil)
	m.ctxPicker.visible = true
	m.ctxPicker.contexts = testContexts
	m.ctxPicker.cursor = 0
	got := update(m, runeKey("k"))
	if got.ctxPicker.cursor != 0 {
		t.Errorf("want ctxPicker.cursor=0 (no underflow), got %d", got.ctxPicker.cursor)
	}
}

func TestContextPicker_CursorDoesNotOverflow(t *testing.T) {
	m := modelWithSorted(nil)
	m.ctxPicker.visible = true
	m.ctxPicker.contexts = testContexts
	m.ctxPicker.cursor = len(testContexts) - 1
	got := update(m, runeKey("j"))
	if got.ctxPicker.cursor != len(testContexts)-1 {
		t.Errorf("want ctxPicker.cursor=%d (no overflow), got %d", len(testContexts)-1, got.ctxPicker.cursor)
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
	m.ctxPicker.visible = true
	m.ctxPicker.contexts = testContexts
	m.ctxPicker.cursor = 1
	update(m, tea.KeyPressMsg{Code: tea.KeyEnter})
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
	m.ctxPicker.visible = true
	m.ctxPicker.contexts = testContexts
	got, cmd := m.Update(docker.ContextSwitchMsg{})
	app := got.(App)
	if app.ctxPicker.visible {
		t.Error("want ctxPicker.visible=false after switch")
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
	m.ctxPicker.visible = true
	m.ctxPicker.contexts = testContexts
	got := update(m, docker.ContextSwitchMsg{Err: errors.New("permission denied")})
	if got.err == nil {
		t.Error("want err set on switch failure")
	}
	if got.ctxPicker.visible {
		t.Error("want ctxPicker.visible=false even on error")
	}
}

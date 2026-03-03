package ui

import (
	"testing"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

var filterContainers = []docker.Container{
	{ID: "abc123def456", Names: "web-app", Image: "nginx:alpine", State: "running"},
	{ID: "111222333444", Names: "database", Image: "postgres:16", State: "running"},
	{ID: "deadbeef0000", Names: "cache", Image: "redis:7", State: "exited"},
}

func modelWithSorted(containers []docker.Container) App {
	m := newWithClient(newStubClient())
	m.sorted = containers
	m.containers = containers
	m.loading = false
	m.width = 120
	m.height = 60
	return m.rebuildTable("")
}

func runeKey(s string) tea.KeyPressMsg {
	r := []rune(s)
	return tea.KeyPressMsg{Code: unicode.ToLower(r[0]), Text: s}
}

func TestFiltered_EmptyQuery(t *testing.T) {
	m := modelWithSorted(filterContainers)
	if got := m.filtered(); len(got) != 3 {
		t.Errorf("want 3, got %d", len(got))
	}
}

func TestFiltered_ByName(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{"web", "web-app"},
		{"data", "database"},
		{"cache", "cache"},
	}
	for _, tc := range tests {
		m := modelWithSorted(filterContainers)
		m.filterQuery = tc.query
		got := m.filtered()
		if len(got) != 1 || got[0].Names != tc.want {
			t.Errorf("query=%q: want [%s], got %v", tc.query, tc.want, got)
		}
	}
}

func TestFiltered_ByImage(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{"nginx", "web-app"},
		{"postgres", "database"},
		{"redis", "cache"},
	}
	for _, tc := range tests {
		m := modelWithSorted(filterContainers)
		m.filterQuery = tc.query
		got := m.filtered()
		if len(got) != 1 || got[0].Names != tc.want {
			t.Errorf("query=%q: want [%s], got %v", tc.query, tc.want, got)
		}
	}
}

func TestFiltered_ByID(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{"abc123", "web-app"},
		{"111222", "database"},
		{"deadbeef", "cache"},
	}
	for _, tc := range tests {
		m := modelWithSorted(filterContainers)
		m.filterQuery = tc.query
		got := m.filtered()
		if len(got) != 1 || got[0].Names != tc.want {
			t.Errorf("query=%q: want [%s], got %v", tc.query, tc.want, got)
		}
	}
}

func TestFiltered_CaseInsensitive(t *testing.T) {
	for _, q := range []string{"NGINX", "Nginx", "nGiNx"} {
		m := modelWithSorted(filterContainers)
		m.filterQuery = q
		got := m.filtered()
		if len(got) != 1 || got[0].Names != "web-app" {
			t.Errorf("query=%q: want [web-app], got %v", q, got)
		}
	}
}

func TestFiltered_NoMatch(t *testing.T) {
	m := modelWithSorted(filterContainers)
	m.filterQuery = "zzznomatch"
	if got := m.filtered(); len(got) != 0 {
		t.Errorf("want 0, got %d", len(got))
	}
}

func TestFiltered_MultipleMatches(t *testing.T) {
	m := modelWithSorted(filterContainers)
	m.filterQuery = "1"
	got := m.filtered()
	if len(got) != 2 {
		t.Errorf("want 2, got %d", len(got))
	}
}

func TestUpdate_SlashEntersFilterMode(t *testing.T) {
	m := modelWithSorted(filterContainers)
	result, _ := m.Update(runeKey("/"))
	if !result.(App).filtering {
		t.Error("want filtering=true after /")
	}
}

func TestUpdate_FilteringTypingBuildsQuery(t *testing.T) {
	m := modelWithSorted(filterContainers)
	m.filtering = true
	for _, ch := range []string{"n", "g", "i", "n", "x"} {
		result, _ := m.Update(runeKey(ch))
		m = result.(App)
	}
	if m.filterQuery != "nginx" {
		t.Errorf("want filterQuery=%q, got %q", "nginx", m.filterQuery)
	}
}

func TestUpdate_FilteringBackspaceRemovesChar(t *testing.T) {
	m := modelWithSorted(filterContainers)
	m.filtering = true
	m.filterQuery = "foo"
	result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if got := result.(App).filterQuery; got != "fo" {
		t.Errorf("want %q, got %q", "fo", got)
	}
}

func TestUpdate_FilteringBackspaceMultibyte(t *testing.T) {
	m := modelWithSorted(filterContainers)
	m.filtering = true
	m.filterQuery = "日本語"
	result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if got := result.(App).filterQuery; got != "日本" {
		t.Errorf("want %q, got %q", "日本", got)
	}
}

func TestUpdate_FilteringBackspaceOnEmpty(t *testing.T) {
	m := modelWithSorted(filterContainers)
	m.filtering = true
	result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if got := result.(App).filterQuery; got != "" {
		t.Errorf("want empty query, got %q", got)
	}
}

func TestUpdate_FilteringEscExitsKeepsQuery(t *testing.T) {
	m := modelWithSorted(filterContainers)
	m.filtering = true
	m.filterQuery = "web"
	result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	got := result.(App)
	if got.filtering {
		t.Error("want filtering=false after esc")
	}
	if got.filterQuery != "web" {
		t.Errorf("want query preserved %q, got %q", "web", got.filterQuery)
	}
}

func TestUpdate_FilteringEnterExitsKeepsQuery(t *testing.T) {
	m := modelWithSorted(filterContainers)
	m.filtering = true
	m.filterQuery = "web"
	result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := result.(App)
	if got.filtering {
		t.Error("want filtering=false after enter")
	}
	if got.filterQuery != "web" {
		t.Errorf("want query preserved %q, got %q", "web", got.filterQuery)
	}
}

func TestUpdate_EscInNormalModeClearsQuery(t *testing.T) {
	m := modelWithSorted(filterContainers)
	m.filterQuery = "web"
	result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if got := result.(App).filterQuery; got != "" {
		t.Errorf("want empty query after esc, got %q", got)
	}
}

func TestUpdate_EscInNormalModeNoopWhenEmpty(t *testing.T) {
	m := modelWithSorted(filterContainers)
	result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if got := result.(App).filterQuery; got != "" {
		t.Errorf("want empty query, got %q", got)
	}
}

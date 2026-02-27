package ui

import (
	"testing"

	"github.com/pivovarit/tdocker/internal/docker"
)

func TestTrunc(t *testing.T) {
	tests := []struct {
		in   string
		max  int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello", 4, "hel…"},
		{"hello", 1, "…"},
		{"hello", 0, ""},
		{"/container_name", 20, "container_name"},
		{"/container_name", 6, "conta…"},
		{"café", 3, "ca…"},
		{"日本語", 2, "日…"}, // CJK
		{"日本語", 3, "日本語"},
	}

	for _, tc := range tests {
		got := trunc(tc.in, tc.max)
		if got != tc.want {
			t.Errorf("trunc(%q, %d) = %q, want %q", tc.in, tc.max, got, tc.want)
		}
	}
}

func TestBuildTableEmpty(t *testing.T) {
	tbl := buildTable(nil, 120)
	if len(tbl.Rows()) != 0 {
		t.Errorf("expected 0 rows, got %d", len(tbl.Rows()))
	}
}

func TestBuildTableRows(t *testing.T) {
	containers := []docker.Container{
		{ID: "abc123", Names: "my-app", Image: "nginx:alpine", State: "running", Status: "Up 2 hours", RunningFor: "2 hours ago", Ports: "80/tcp"},
		{ID: "def456", Names: "/other", Image: "postgres:16", State: "exited", Status: "Exited (0)", RunningFor: "1 day ago", Ports: ""},
	}
	tbl := buildTable(containers, 180)
	if len(tbl.Rows()) != 2 {
		t.Errorf("expected 2 rows, got %d", len(tbl.Rows()))
	}
	if got := tbl.Rows()[1][1]; got != "other" {
		t.Errorf("expected name %q, got %q", "other", got)
	}
}

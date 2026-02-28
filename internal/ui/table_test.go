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
		{"日本語", 2, "日…"},
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

func TestBuildTableName_NonCompose(t *testing.T) {
	containers := []docker.Container{
		{Names: "standalone"},
	}
	if got := buildTableName(containers, 0); got != "standalone" {
		t.Errorf("want %q, got %q", "standalone", got)
	}
}

func TestBuildTableName_SingleComposeContainer(t *testing.T) {
	containers := []docker.Container{
		{Names: "web", Labels: docker.Labels{"com.docker.compose.project": "myapp", "com.docker.compose.service": "web"}},
	}
	if got := buildTableName(containers, 0); got != "myapp/web" {
		t.Errorf("want %q, got %q", "myapp/web", got)
	}
}

func TestBuildTableName_TreeChars(t *testing.T) {
	labels := func(project, service string) docker.Labels {
		return docker.Labels{"com.docker.compose.project": project, "com.docker.compose.service": service}
	}
	containers := []docker.Container{
		{Names: "db", Labels: labels("app", "db")},
		{Names: "cache", Labels: labels("app", "cache")},
		{Names: "web", Labels: labels("app", "web")},
	}
	tests := []struct {
		i    int
		want string
	}{
		{0, "┬ app/db"},
		{1, "├ app/cache"},
		{2, "└ app/web"},
	}
	for _, tc := range tests {
		if got := buildTableName(containers, tc.i); got != tc.want {
			t.Errorf("i=%d: want %q, got %q", tc.i, tc.want, got)
		}
	}
}

func TestBuildTableName_TwoInGroup(t *testing.T) {
	labels := func(project, service string) docker.Labels {
		return docker.Labels{"com.docker.compose.project": project, "com.docker.compose.service": service}
	}
	containers := []docker.Container{
		{Names: "api", Labels: labels("proj", "api")},
		{Names: "db", Labels: labels("proj", "db")},
	}
	if got := buildTableName(containers, 0); got != "┬ proj/api" {
		t.Errorf("want %q, got %q", "┬ proj/api", got)
	}
	if got := buildTableName(containers, 1); got != "└ proj/db" {
		t.Errorf("want %q, got %q", "└ proj/db", got)
	}
}

func TestBuildTableName_AdjacentProjects(t *testing.T) {
	labels := func(project, service string) docker.Labels {
		return docker.Labels{"com.docker.compose.project": project, "com.docker.compose.service": service}
	}
	containers := []docker.Container{
		{Names: "a", Labels: labels("proj1", "a")},
		{Names: "b", Labels: labels("proj1", "b")},
		{Names: "c", Labels: labels("proj2", "c")},
		{Names: "d", Labels: labels("proj2", "d")},
	}
	if got := buildTableName(containers, 1); got != "└ proj1/b" {
		t.Errorf("want %q, got %q", "└ proj1/b", got)
	}
	if got := buildTableName(containers, 2); got != "┬ proj2/c" {
		t.Errorf("want %q, got %q", "┬ proj2/c", got)
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

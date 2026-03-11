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

func TestBuildTableName_ComposeNoTreeChars(t *testing.T) {
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
		{0, "app/db"},
		{1, "app/cache"},
		{2, "app/web"},
	}
	for _, tc := range tests {
		if got := buildTableName(containers, tc.i); got != tc.want {
			t.Errorf("i=%d: want %q, got %q", tc.i, tc.want, got)
		}
	}
}

func TestComposeTreeChar(t *testing.T) {
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
		{0, "┬"},
		{1, "├"},
		{2, "└"},
	}
	for _, tc := range tests {
		got := ansiRe.ReplaceAllString(composeTreeChar(containers, tc.i), "")
		if got != tc.want {
			t.Errorf("i=%d: want %q, got %q", tc.i, tc.want, got)
		}
	}
}

func TestComposeTreeChar_TwoInGroup(t *testing.T) {
	labels := func(project, service string) docker.Labels {
		return docker.Labels{"com.docker.compose.project": project, "com.docker.compose.service": service}
	}
	containers := []docker.Container{
		{Names: "api", Labels: labels("proj", "api")},
		{Names: "db", Labels: labels("proj", "db")},
	}
	if got := ansiRe.ReplaceAllString(composeTreeChar(containers, 0), ""); got != "┬" {
		t.Errorf("want %q, got %q", "┬", got)
	}
	if got := ansiRe.ReplaceAllString(composeTreeChar(containers, 1), ""); got != "└" {
		t.Errorf("want %q, got %q", "└", got)
	}
}

func TestComposeTreeChar_AdjacentProjects(t *testing.T) {
	labels := func(project, service string) docker.Labels {
		return docker.Labels{"com.docker.compose.project": project, "com.docker.compose.service": service}
	}
	containers := []docker.Container{
		{Names: "a", Labels: labels("proj1", "a")},
		{Names: "b", Labels: labels("proj1", "b")},
		{Names: "c", Labels: labels("proj2", "c")},
		{Names: "d", Labels: labels("proj2", "d")},
	}
	if got := ansiRe.ReplaceAllString(composeTreeChar(containers, 1), ""); got != "└" {
		t.Errorf("want %q, got %q", "└", got)
	}
	if got := ansiRe.ReplaceAllString(composeTreeChar(containers, 2), ""); got != "┬" {
		t.Errorf("want %q, got %q", "┬", got)
	}
}

func TestBuildTableName_ExpandedStandaloneContainer(t *testing.T) {
	containers := []docker.Container{
		{ID: "s1", Names: "solo", State: "running"},
		{State: "detail", Names: "└  Network  bridge (172.17.0.2)"},
	}
	got := buildTableName(containers, 0)
	want := "solo"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestBuildTableName_DetailRowPassesThrough(t *testing.T) {
	containers := []docker.Container{
		{State: "detail", Names: "│  Ports    80/tcp → 0.0.0.0:80"},
	}
	got := buildTableName(containers, 0)
	want := "│  Ports    80/tcp → 0.0.0.0:80"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestBuildTableName_NotExpandedStandaloneNoPrefix(t *testing.T) {
	containers := []docker.Container{
		{ID: "s1", Names: "solo", State: "running"},
		{ID: "s2", Names: "other", State: "running"},
	}
	got := buildTableName(containers, 0)
	want := "solo"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
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

package docker

import (
	"encoding/json"
	"testing"
)

func TestLabels_UnmarshalJSON_StringFormat(t *testing.T) {
	input := `{"ID":"abc","Labels":"com.docker.compose.project=myapp,com.docker.compose.service=web"}`
	var c Container
	if err := json.Unmarshal([]byte(input), &c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := c.ComposeProject(); got != "myapp" {
		t.Errorf("want ComposeProject=%q, got %q", "myapp", got)
	}
	if got := c.ComposeService(); got != "web" {
		t.Errorf("want ComposeService=%q, got %q", "web", got)
	}
}

func TestLabels_UnmarshalJSON_MapFormat(t *testing.T) {
	input := `{"ID":"abc","Labels":{"com.docker.compose.project":"myapp","com.docker.compose.service":"db"}}`
	var c Container
	if err := json.Unmarshal([]byte(input), &c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := c.ComposeProject(); got != "myapp" {
		t.Errorf("want ComposeProject=%q, got %q", "myapp", got)
	}
	if got := c.ComposeService(); got != "db" {
		t.Errorf("want ComposeService=%q, got %q", "db", got)
	}
}

func TestLabels_UnmarshalJSON_NoLabels(t *testing.T) {
	input := `{"ID":"abc","Labels":""}`
	var c Container
	if err := json.Unmarshal([]byte(input), &c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := c.ComposeProject(); got != "" {
		t.Errorf("want empty ComposeProject, got %q", got)
	}
}

func TestLabels_UnmarshalJSON_NullLabels(t *testing.T) {
	input := `{"ID":"abc","Labels":null}`
	var c Container
	if err := json.Unmarshal([]byte(input), &c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := c.ComposeProject(); got != "" {
		t.Errorf("want empty ComposeProject, got %q", got)
	}
}

func TestSort_ComposeGroupedWithinStateBucket(t *testing.T) {
	containers := []Container{
		{ID: "1", Names: "standalone", State: "running"},
		{ID: "2", Names: "web", State: "running", Labels: Labels{"com.docker.compose.project": "app", "com.docker.compose.service": "web"}},
		{ID: "3", Names: "db", State: "running", Labels: Labels{"com.docker.compose.project": "app", "com.docker.compose.service": "db"}},
		{ID: "4", Names: "cache", State: "running", Labels: Labels{"com.docker.compose.project": "infra", "com.docker.compose.service": "cache"}},
	}
	sorted := Sort(containers)

	if sorted[0].Names != "db" {
		t.Errorf("want db first (app project, alphabetical), got %s", sorted[0].Names)
	}
	if sorted[1].Names != "web" {
		t.Errorf("want web second, got %s", sorted[1].Names)
	}
	if sorted[2].Names != "cache" {
		t.Errorf("want cache third (infra project), got %s", sorted[2].Names)
	}
	if sorted[3].Names != "standalone" {
		t.Errorf("want standalone last (non-compose), got %s", sorted[3].Names)
	}
}

func TestSort_RunningBeforeStopped(t *testing.T) {
	containers := []Container{
		{ID: "1", Names: "stopped", State: "exited"},
		{ID: "2", Names: "running", State: "running"},
	}
	sorted := Sort(containers)
	if sorted[0].State != "running" {
		t.Errorf("want running first, got %s", sorted[0].State)
	}
}

func TestSort_ComposeSameProjectSortedByService(t *testing.T) {
	containers := []Container{
		{ID: "1", Names: "z-svc", State: "running", Labels: Labels{"com.docker.compose.project": "proj", "com.docker.compose.service": "z-svc"}},
		{ID: "2", Names: "a-svc", State: "running", Labels: Labels{"com.docker.compose.project": "proj", "com.docker.compose.service": "a-svc"}},
	}
	sorted := Sort(containers)
	if sorted[0].Names != "a-svc" {
		t.Errorf("want a-svc first, got %s", sorted[0].Names)
	}
}

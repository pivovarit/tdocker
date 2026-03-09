package docker

import (
	"encoding/json"
	"os/exec"
	"strings"
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

func TestExecErr_Nil(t *testing.T) {
	if got := execErr(nil); got != nil {
		t.Errorf("want nil, got %v", got)
	}
}

func TestExecErr_Exit126_FriendlyMessage(t *testing.T) {
	err := exec.Command("sh", "-c", "exit 126").Run()
	got := execErr(err)
	if got == nil {
		t.Fatal("want non-nil error")
	}
	if !strings.Contains(got.Error(), "shell not found") {
		t.Errorf("want friendly message, got %q", got.Error())
	}
	if !strings.Contains(got.Error(), "'x'") {
		t.Errorf("want reference to debug key, got %q", got.Error())
	}
}

func TestExecErr_Exit127_FriendlyMessage(t *testing.T) {
	err := exec.Command("sh", "-c", "exit 127").Run()
	got := execErr(err)
	if got == nil {
		t.Fatal("want non-nil error")
	}
	if !strings.Contains(got.Error(), "shell not found") {
		t.Errorf("want friendly message, got %q", got.Error())
	}
	if !strings.Contains(got.Error(), "'x'") {
		t.Errorf("want reference to debug key, got %q", got.Error())
	}
}

func TestExecErr_OtherExitCode_OriginalError(t *testing.T) {
	err := exec.Command("sh", "-c", "exit 1").Run()
	got := execErr(err)
	if got == nil {
		t.Fatal("want non-nil error")
	}
	if strings.Contains(got.Error(), "shell not found") {
		t.Errorf("want original error for exit 1, got %q", got.Error())
	}
}

func TestExpandInspectMsg_HasContainerID(t *testing.T) {
	msg := ExpandInspectMsg{ContainerID: "abc123", Data: nil, Err: nil}
	if msg.ContainerID != "abc123" {
		t.Errorf("want ContainerID=%q, got %q", "abc123", msg.ContainerID)
	}
}

func TestInspectData_Networks_ParsedFromJSON(t *testing.T) {
	raw := `[{
		"Image": "sha256:abc",
		"Config": {"Env": []},
		"Mounts": [],
		"NetworkSettings": {
			"Ports": {},
			"Networks": {
				"bridge": {"IPAddress": "172.17.0.2"},
				"mynet":  {"IPAddress": "10.0.0.5"}
			}
		}
	}]`
	data, err := parseInspectData([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data.Networks) != 2 {
		t.Fatalf("want 2 networks, got %d", len(data.Networks))
	}
	if data.Networks[0].Name != "bridge" || data.Networks[0].IPAddress != "172.17.0.2" {
		t.Errorf("first network: got {%q %q}", data.Networks[0].Name, data.Networks[0].IPAddress)
	}
	if data.Networks[1].Name != "mynet" || data.Networks[1].IPAddress != "10.0.0.5" {
		t.Errorf("second network: got {%q %q}", data.Networks[1].Name, data.Networks[1].IPAddress)
	}
}

package docker

import (
	"testing"
	"time"
)

func TestParseInspectData_StateFields(t *testing.T) {
	raw := []byte(`[{
		"Image": "sha256:abc",
		"State": {
			"Status": "exited",
			"ExitCode": 137,
			"Error": "OOM",
			"OOMKilled": true,
			"StartedAt": "2026-04-17T10:00:00Z",
			"FinishedAt": "2026-04-17T10:05:00Z"
		},
		"RestartCount": 3,
		"HostConfig": {
			"RestartPolicy": {"Name": "on-failure", "MaximumRetryCount": 5}
		},
		"Config": {
			"Env": ["FOO=bar"]
		},
		"Mounts": [],
		"NetworkSettings": {"Ports": {}, "Networks": {}}
	}]`)

	data, err := parseInspectData(raw)
	if err != nil {
		t.Fatalf("parseInspectData: %v", err)
	}
	if data.State.Status != "exited" {
		t.Errorf("Status = %q, want exited", data.State.Status)
	}
	if data.State.ExitCode != 137 {
		t.Errorf("ExitCode = %d, want 137", data.State.ExitCode)
	}
	if !data.State.OOMKilled {
		t.Error("OOMKilled should be true")
	}
	if data.State.Error != "OOM" {
		t.Errorf("Error = %q, want OOM", data.State.Error)
	}
	want := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	if !data.State.StartedAt.Equal(want) {
		t.Errorf("StartedAt = %v, want %v", data.State.StartedAt, want)
	}
	if data.RestartCount != 3 {
		t.Errorf("RestartCount = %d, want 3", data.RestartCount)
	}
	if data.RestartPolicy.Name != "on-failure" {
		t.Errorf("RestartPolicy.Name = %q, want on-failure", data.RestartPolicy.Name)
	}
	if data.RestartPolicy.MaximumRetryCount != 5 {
		t.Errorf("RestartPolicy.MaximumRetryCount = %d, want 5", data.RestartPolicy.MaximumRetryCount)
	}
}

func TestParseInspectData_HealthcheckAndHealth(t *testing.T) {
	raw := []byte(`[{
		"Image": "sha256:abc",
		"State": {
			"Status": "running",
			"Health": {
				"Status": "unhealthy",
				"FailingStreak": 4,
				"Log": [
					{"Start":"2026-04-17T10:00:00Z","End":"2026-04-17T10:00:01Z","ExitCode":1,"Output":"connection refused"}
				]
			}
		},
		"Config": {
			"Env": [],
			"Healthcheck": {"Test":["CMD","curl","http://localhost"],"Interval":30000000000,"Timeout":5000000000,"Retries":3,"StartPeriod":0}
		},
		"Mounts": [],
		"NetworkSettings": {"Ports": {}, "Networks": {}}
	}]`)

	data, err := parseInspectData(raw)
	if err != nil {
		t.Fatalf("parseInspectData: %v", err)
	}
	if data.Healthcheck == nil {
		t.Fatal("Healthcheck should not be nil")
	}
	if len(data.Healthcheck.Test) != 3 || data.Healthcheck.Test[1] != "curl" {
		t.Errorf("Healthcheck.Test = %v", data.Healthcheck.Test)
	}
	if data.Healthcheck.Interval != 30*time.Second {
		t.Errorf("Interval = %v, want 30s", data.Healthcheck.Interval)
	}
	if data.State.Health == nil {
		t.Fatal("State.Health should not be nil")
	}
	if data.State.Health.Status != "unhealthy" {
		t.Errorf("Health.Status = %q, want unhealthy", data.State.Health.Status)
	}
	if len(data.State.Health.Log) != 1 || data.State.Health.Log[0].Output != "connection refused" {
		t.Errorf("Health.Log[0].Output = %v", data.State.Health.Log)
	}
}

func TestParseInspectData_NoHealthcheck(t *testing.T) {
	raw := []byte(`[{
		"Image": "sha256:abc",
		"State": {"Status": "running"},
		"Config": {"Env": []},
		"Mounts": [],
		"NetworkSettings": {"Ports": {}, "Networks": {}}
	}]`)

	data, err := parseInspectData(raw)
	if err != nil {
		t.Fatalf("parseInspectData: %v", err)
	}
	if data.Healthcheck != nil {
		t.Errorf("Healthcheck should be nil, got %+v", data.Healthcheck)
	}
	if data.State.Health != nil {
		t.Errorf("State.Health should be nil, got %+v", data.State.Health)
	}
}

package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/pivovarit/tdocker/internal/docker"
)

func TestDiagnose_Healthy(t *testing.T) {
	data := &docker.InspectData{State: docker.ContainerState{Status: "running"}}
	d := diagnose(data, nil, time.Now())
	if d.Severity != "" {
		t.Errorf("Severity = %q, want empty (healthy)", d.Severity)
	}
}

func TestDiagnose_OOM(t *testing.T) {
	data := &docker.InspectData{
		State: docker.ContainerState{Status: "exited", ExitCode: 137, OOMKilled: true},
	}
	d := diagnose(data, nil, time.Now())
	if d.Severity != "warn" {
		t.Errorf("Severity = %q, want warn", d.Severity)
	}
	if !strings.Contains(d.Title, "OOM") {
		t.Errorf("Title = %q, want contains OOM", d.Title)
	}
}

func TestDiagnose_NonZeroExit(t *testing.T) {
	data := &docker.InspectData{
		State: docker.ContainerState{Status: "exited", ExitCode: 1, Error: "boom"},
	}
	d := diagnose(data, nil, time.Now())
	if d.Severity != "error" {
		t.Errorf("Severity = %q, want error", d.Severity)
	}
	foundErr := false
	for _, l := range d.Details {
		if strings.Contains(l, "boom") {
			foundErr = true
		}
	}
	if !foundErr {
		t.Errorf("Details should contain error message, got %v", d.Details)
	}
}

func TestDiagnose_Unhealthy(t *testing.T) {
	data := &docker.InspectData{
		State: docker.ContainerState{
			Status: "running",
			Health: &docker.Health{
				Status:        "unhealthy",
				FailingStreak: 4,
				Log: []docker.HealthLogEntry{
					{Output: "connection refused"},
				},
			},
		},
	}
	d := diagnose(data, nil, time.Now())
	if d.Severity != "error" {
		t.Errorf("Severity = %q, want error", d.Severity)
	}
	found := false
	for _, l := range d.Details {
		if strings.Contains(l, "connection refused") {
			found = true
		}
	}
	if !found {
		t.Errorf("Details should include last check output, got %v", d.Details)
	}
}

func TestDiagnose_RestartLoop(t *testing.T) {
	now := time.Now()
	events := []docker.Event{
		{Action: "die", Time: now.Add(-30 * time.Minute).Unix()},
		{Action: "die", Time: now.Add(-15 * time.Minute).Unix()},
		{Action: "die", Time: now.Add(-1 * time.Minute).Unix()},
	}
	data := &docker.InspectData{
		State:         docker.ContainerState{Status: "running"},
		RestartPolicy: docker.RestartPolicy{Name: "on-failure"},
	}
	d := diagnose(data, events, now)
	if d.Severity != "warn" {
		t.Errorf("Severity = %q, want warn", d.Severity)
	}
}

func TestDiagnose_RestartLoop_IgnoresOldEvents(t *testing.T) {
	now := time.Now()
	events := []docker.Event{
		{Action: "die", Time: now.Add(-2 * time.Hour).Unix()},
		{Action: "die", Time: now.Add(-90 * time.Minute).Unix()},
		{Action: "die", Time: now.Add(-30 * time.Minute).Unix()},
	}
	data := &docker.InspectData{State: docker.ContainerState{Status: "running"}}
	d := diagnose(data, events, now)
	if d.Severity != "" {
		t.Errorf("Severity = %q, want empty (only 1 in window)", d.Severity)
	}
}

func TestDiagnose_ErrorBeatsWarn(t *testing.T) {
	now := time.Now()
	events := []docker.Event{
		{Action: "die", Time: now.Add(-30 * time.Minute).Unix()},
		{Action: "die", Time: now.Add(-15 * time.Minute).Unix()},
		{Action: "die", Time: now.Add(-1 * time.Minute).Unix()},
	}
	data := &docker.InspectData{
		State: docker.ContainerState{Status: "exited", ExitCode: 1, Error: "boom"},
	}
	d := diagnose(data, events, now)
	if d.Severity != "error" {
		t.Errorf("Severity = %q, want error (error beats warn)", d.Severity)
	}
}

func TestDiagnose_OOMAloneIsWarn(t *testing.T) {
	data := &docker.InspectData{
		State: docker.ContainerState{Status: "exited", ExitCode: 137, OOMKilled: true},
	}
	d := diagnose(data, nil, time.Now())
	if d.Severity != "warn" {
		t.Errorf("Severity = %q, want warn (OOM alone)", d.Severity)
	}
	exitMentions := 0
	for _, l := range d.Details {
		if strings.Contains(l, "Exited with code") {
			exitMentions++
		}
	}
	if exitMentions != 0 {
		t.Errorf("OOM should not also produce 'Exited with code' line; got %v", d.Details)
	}
}

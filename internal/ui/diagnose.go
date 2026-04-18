package ui

import (
	"fmt"
	"time"

	"github.com/pivovarit/tdocker/internal/docker"
)

type Diagnosis struct {
	Severity string
	Title    string
	Details  []string
}

const diagnoseRestartWindow = time.Hour
const diagnoseRestartLoopThreshold = 3
const diagnoseHealthOutputMax = 120

func diagnose(data *docker.InspectData, events []docker.Event, now time.Time) Diagnosis {
	if data == nil {
		return Diagnosis{}
	}

	var details []string
	severity := ""

	upgrade := func(to string) {
		if to == "error" || severity == "" {
			severity = to
		}
	}

	if data.State.OOMKilled {
		details = append(details, "OOM killed. Check memory limit vs. workload.")
		upgrade("warn")
	}

	if data.State.Status == "exited" && data.State.ExitCode != 0 && !data.State.OOMKilled {
		details = append(details, fmt.Sprintf("Exited with code %d.", data.State.ExitCode))
		if data.State.Error != "" {
			details = append(details, "Last error: "+data.State.Error)
		}
		upgrade("error")
	}

	if data.State.Health != nil && data.State.Health.Status == "unhealthy" {
		line := fmt.Sprintf("Healthcheck failing (streak: %d).", data.State.Health.FailingStreak)
		details = append(details, line)
		if n := len(data.State.Health.Log); n > 0 {
			out := data.State.Health.Log[n-1].Output
			if len(out) > diagnoseHealthOutputMax {
				out = out[:diagnoseHealthOutputMax] + "…"
			}
			if out != "" {
				details = append(details, "Last check: "+out)
			}
		}
		upgrade("error")
	}

	if n := countDieEvents(events, now.Add(-diagnoseRestartWindow)); n >= diagnoseRestartLoopThreshold {
		policy := data.RestartPolicy.Name
		if policy == "" {
			policy = "no"
		}
		details = append(details, fmt.Sprintf("Restart loop: %d deaths in last hour. Policy: %s.", n, policy))
		upgrade("warn")
	}

	if severity == "" {
		return Diagnosis{}
	}
	return Diagnosis{Severity: severity, Title: titleFor(data), Details: details}
}

func titleFor(data *docker.InspectData) string {
	switch {
	case data.State.OOMKilled:
		return "OOM killed"
	case data.State.Status == "exited" && data.State.ExitCode != 0:
		return fmt.Sprintf("Exited with code %d", data.State.ExitCode)
	case data.State.Health != nil && data.State.Health.Status == "unhealthy":
		return "Unhealthy"
	default:
		return "Restart loop"
	}
}

func countDieEvents(events []docker.Event, since time.Time) int {
	cutoff := since.Unix()
	n := 0
	for _, ev := range events {
		if ev.Action == "die" && ev.Time >= cutoff {
			n++
		}
	}
	return n
}

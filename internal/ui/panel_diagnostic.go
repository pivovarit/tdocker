package ui

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

func buildDiagnosticLines(
	c docker.Container,
	data *docker.InspectData,
	events []docker.Event,
	eventsErr error,
	logs []string,
	width int,
	now time.Time,
) []string {
	var lines []string
	section := func(title string) { lines = append(lines, inspectSectionStyle.Render(title)) }
	blank := func() { lines = append(lines, "") }
	val := func(s string) { lines = append(lines, "  "+inspectValueStyle.Render(s)) }
	kv := func(k, v string) { lines = append(lines, "  "+keyStyle.Render(k)+inspectValueStyle.Render(v)) }

	section(statusLine(c, data, now))
	lines = append(lines, inspectSectionStyle.Render(restartsLine(data)))
	blank()

	diag := diagnose(data, events, now)
	if diag.Severity != "" {
		style := diagnosisWarnStyle
		if diag.Severity == "error" {
			style = diagnosisErrorStyle
		}
		lines = append(lines, style.Render("⚠ Diagnosis: "+diag.Title))
		for _, d := range diag.Details {
			val(d)
		}
		blank()
	}

	section("Recent events")
	switch {
	case eventsErr != nil:
		val("(events unavailable: " + firstLine(eventsErr.Error()) + ")")
	case len(events) == 0:
		val("(no events in last hour)")
	default:
		for _, ev := range events {
			val(formatEventLine(ev))
		}
	}
	blank()

	section("Log tail")
	if len(logs) == 0 {
		val("(no logs)")
	} else {
		for _, l := range logs {
			val(truncateForWidth(l, width-4))
		}
	}
	blank()

	if data != nil && data.Healthcheck != nil {
		section("Health")
		kv("test: ", strings.Join(data.Healthcheck.Test, " "))
		kv("interval: ", data.Healthcheck.Interval.String())
		kv("retries: ", fmt.Sprintf("%d", data.Healthcheck.Retries))
		if data.State.Health != nil {
			kv("status: ", data.State.Health.Status)
			kv("failing: ", fmt.Sprintf("%d", data.State.Health.FailingStreak))
			for i, le := range data.State.Health.Log {
				val(fmt.Sprintf("  check[%d] exit=%d: %s", i, le.ExitCode, truncateForWidth(le.Output, width-20)))
			}
		}
		blank()
	}

	if data != nil {
		var envLines []string
		inEnv := false

		envSection := func(title string) { envLines = append(envLines, inspectSectionStyle.Render(title)) }
		envKV := func(k, v string) { envLines = append(envLines, "  "+keyStyle.Render(k)+inspectValueStyle.Render(v)) }
		envVal := func(s string) { envLines = append(envLines, "  "+inspectValueStyle.Render(s)) }

		for _, l := range data.Lines(width) {
			if l.Kind == docker.InspectLineSection {
				inEnv = l.Key == "Environment"
			}

			if inEnv {
				switch l.Kind {
				case docker.InspectLineSection:
					envSection(l.Key)
				case docker.InspectLineKeyValue:
					envKV(l.Key, l.Value)
				case docker.InspectLineValue:
					envVal(l.Value)
				case docker.InspectLineBlank:
					envLines = append(envLines, "")
				}
				continue
			}

			switch l.Kind {
			case docker.InspectLineSection:
				section(l.Key)
			case docker.InspectLineKeyValue:
				kv(l.Key, l.Value)
			case docker.InspectLineValue:
				val(l.Value)
			case docker.InspectLineBlank:
				lines = append(lines, "")
			}
		}

		lines = append(lines, envLines...)
	}

	return lines
}

func statusLine(c docker.Container, data *docker.InspectData, now time.Time) string {
	if data == nil {
		return "Status: " + c.Status
	}
	switch data.State.Status {
	case "running":
		uptime := ""
		if !data.State.StartedAt.IsZero() {
			uptime = "  uptime: " + compactDuration(now.Sub(data.State.StartedAt))
		}
		return "Status: running" + uptime
	case "exited":
		exit := fmt.Sprintf("Status: exited (%d)", data.State.ExitCode)
		if !data.State.FinishedAt.IsZero() {
			exit += "  died " + compactDuration(now.Sub(data.State.FinishedAt)) + " ago"
		}
		if data.State.OOMKilled {
			exit += "  ⚠ OOM"
		}
		return exit
	case "restarting":
		return "Status: restarting  attempt " + fmt.Sprintf("%d", data.RestartCount)
	case "paused":
		return "Status: paused"
	default:
		if data.State.Status == "" {
			return "Status: " + c.Status
		}
		return "Status: " + data.State.Status
	}
}

func restartsLine(data *docker.InspectData) string {
	if data == nil {
		return "Restarts: —  Policy: —"
	}
	policy := data.RestartPolicy.Name
	if policy == "" {
		policy = "no"
	}
	out := fmt.Sprintf("Restarts: %d  Policy: %s", data.RestartCount, policy)
	if data.RestartPolicy.MaximumRetryCount > 0 {
		out += fmt.Sprintf(" (max %d)", data.RestartPolicy.MaximumRetryCount)
	}
	return out
}

func compactDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		h := int(d.Hours())
		m := int((d - time.Duration(h)*time.Hour).Minutes())
		return fmt.Sprintf("%dh %dm", h, m)
	default:
		return fmt.Sprintf("%dd", int(d.Hours())/24)
	}
}

func formatEventLine(ev docker.Event) string {
	extras := ""
	if code := ev.Actor.Attributes["exitCode"]; code != "" {
		extras += "  exit=" + code
	}
	return fmt.Sprintf("%s  %-8s %s%s", ev.Timestamp(), ev.Action, ev.Name(), extras)
}

func truncateForWidth(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n >= max-1 {
			break
		}
		b.WriteRune(r)
		n++
	}
	b.WriteRune('…')
	return b.String()
}

type diagnosticState struct {
	visible     bool
	container   string
	containerID string
	data        *docker.InspectData
	events      []docker.Event
	eventsErr   error
	logs        []string
	logsCancel  context.CancelFunc
	logsGen     int
	loading     bool
	err         error
	lines       []string
	scroll      scrollState
}

func (m App) closeDiagnostic() App {
	if m.diagnostic.logsCancel != nil {
		m.diagnostic.logsCancel()
	}
	m.diagnostic = diagnosticState{}
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) handleDiagnosticKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.Code {
	case tea.KeyEsc, 'i':
		m = m.closeDiagnostic()
		return m, nil
	case 'r':
		return m.refreshDiagnostic()
	default:
		m.diagnostic.scroll, _ = m.diagnostic.scroll.handleKey(msg, len(m.diagnostic.lines), m.diagnosticPanelHeight()-2)
	}
	return m, nil
}

func (m App) refreshDiagnostic() (tea.Model, tea.Cmd) {
	if m.diagnostic.logsCancel != nil {
		m.diagnostic.logsCancel()
	}
	m.diagnostic.logsGen++
	m.diagnostic.logs = nil
	m.diagnostic.events = nil
	m.diagnostic.loading = true
	ctx, cancel := context.WithCancel(context.Background())
	m.diagnostic.logsCancel = cancel
	id := m.diagnostic.containerID
	return m, tea.Batch(
		m.client.InspectContainer(id),
		m.client.FetchContainerEvents(id, time.Hour),
		m.client.StartLogs(ctx, id, "50", false, "", m.diagnostic.logsGen),
	)
}

func (m App) renderDiagnosticPanel() string {
	return m.renderPanel(" Diagnostic: "+m.diagnostic.container, func(b *strings.Builder) {
		maxLines := m.diagnosticPanelHeight() - 2
		if m.diagnostic.err != nil {
			b.WriteString(errorStyle.Render(m.diagnostic.err.Error()))
			b.WriteString("\n")
			panelPad(b, 1, maxLines)
			return
		}
		if m.diagnostic.loading {
			b.WriteString(emptyStyle.Render("Loading…"))
			b.WriteString("\n")
			panelPad(b, 1, maxLines)
			return
		}
		start := m.diagnostic.scroll.offset
		end := start + maxLines
		if end > len(m.diagnostic.lines) {
			end = len(m.diagnostic.lines)
		}
		for _, line := range m.diagnostic.lines[start:end] {
			b.WriteString(line)
			b.WriteString("\n")
		}
		panelPad(b, end-start, maxLines)
	})
}

func (m App) rebuildDiagnosticLines() App {
	m.diagnostic.lines = buildDiagnosticLines(
		docker.Container{Names: m.diagnostic.container, ID: m.diagnostic.containerID},
		m.diagnostic.data, m.diagnostic.events, m.diagnostic.eventsErr, m.diagnostic.logs, m.width, time.Now(),
	)
	return m
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

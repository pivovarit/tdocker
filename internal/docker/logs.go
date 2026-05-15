package docker

import (
	"context"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type (
	LogsLineMsg struct {
		Line string
		Next tea.Cmd
		Gen  int
	}
	LogsEndMsg struct {
		Err error
		Gen int
	}
)

func (CLI) SupportsGrep() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutDebug)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", "logs", "--help").CombinedOutput()
		if err != nil {
			return GrepSupportMsg{Available: false}
		}
		return GrepSupportMsg{Available: strings.Contains(string(out), "--grep")}
	}
}

func (CLI) StartLogs(ctx context.Context, id string, tail string, timestamps bool, grep string, gen int) tea.Cmd {
	args := []string{"logs", "--follow", "--tail", tail}
	if timestamps {
		args = append(args, "--timestamps")
	}
	if grep != "" {
		args = append(args, "--grep", grep)
	}
	args = append(args, id)
	return streamCmd(ctx, exec.CommandContext(ctx, "docker", args...),
		func(line string, next tea.Cmd) tea.Msg {
			return LogsLineMsg{Line: line, Next: next, Gen: gen}
		},
		func(err error) tea.Msg { return LogsEndMsg{Err: err, Gen: gen} },
	)
}

func (CLI) StartComposeLogs(ctx context.Context, project string, tail string, timestamps bool, gen int) tea.Cmd {
	args := []string{"compose", "-p", project, "logs", "--follow", "--tail", tail}
	if timestamps {
		args = append(args, "--timestamps")
	}
	return streamCmd(ctx, exec.CommandContext(ctx, "docker", args...),
		func(line string, next tea.Cmd) tea.Msg {
			return LogsLineMsg{Line: line, Next: next, Gen: gen}
		},
		func(err error) tea.Msg { return LogsEndMsg{Err: err, Gen: gen} },
	)
}

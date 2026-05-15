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
	LogsOpts struct {
		ContainerID    string
		ComposeProject string
		Tail           string
		Timestamps     bool
		Grep           string
		Gen            int
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

func (CLI) StartLogs(ctx context.Context, opts LogsOpts) tea.Cmd {
	var args []string
	if opts.ComposeProject != "" {
		args = []string{"compose", "-p", opts.ComposeProject, "logs", "--follow", "--tail", opts.Tail}
	} else {
		args = []string{"logs", "--follow", "--tail", opts.Tail}
		if opts.Grep != "" {
			args = append(args, "--grep", opts.Grep)
		}
		args = append(args, opts.ContainerID)
	}
	if opts.Timestamps {
		args = append(args, "--timestamps")
	}
	return streamCmd(ctx, exec.CommandContext(ctx, "docker", args...),
		func(line string, next tea.Cmd) tea.Msg {
			return LogsLineMsg{Line: line, Next: next, Gen: opts.Gen}
		},
		func(err error) tea.Msg { return LogsEndMsg{Err: err, Gen: opts.Gen} },
	)
}

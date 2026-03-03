package docker

import (
	"context"
	"io"
	"os/exec"

	tea "charm.land/bubbletea/v2"
)

func (CLI) CheckShellAvailable(id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutDebug)
		defer cancel()
		for _, shell := range []string{"bash", "sh"} {
			cmd := exec.CommandContext(ctx, "docker", "exec", id, shell, "-c", "exit 0")
			cmd.Stdout = io.Discard
			cmd.Stderr = io.Discard
			if cmd.Run() == nil {
				return ShellAvailableMsg{ID: id, Available: true, Shell: shell}
			}
		}
		return ShellAvailableMsg{ID: id, Available: false}
	}
}

func (CLI) ExecContainer(id, shell string) tea.Cmd {
	return tea.ExecProcess(
		exec.Command("docker", "exec", "-it", id, shell),
		func(err error) tea.Msg { return ExecDoneMsg{Err: execErr(err)} },
	)
}

func (CLI) CheckDebugAvailable(id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutDebug)
		defer cancel()
		cmd := exec.CommandContext(ctx, "docker", "debug", "--help")
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		return DebugAvailableMsg{ID: id, Available: cmd.Run() == nil}
	}
}

func (CLI) DebugContainer(id string) tea.Cmd {
	return tea.ExecProcess(
		exec.Command("docker", "debug", id),
		func(err error) tea.Msg { return ExecDoneMsg{Err: err} },
	)
}

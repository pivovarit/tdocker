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
		cmd := exec.CommandContext(ctx, "docker", "exec", id, "sh", "-c", "exit 0")
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		return ShellAvailableMsg{ID: id, Available: cmd.Run() == nil}
	}
}

func (CLI) ExecContainer(id string) tea.Cmd {
	return tea.ExecProcess(
		exec.Command("docker", "exec", "-it", id, "sh"),
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

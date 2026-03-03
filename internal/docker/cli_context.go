package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (CLI) FetchContexts() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutContext)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", "context", "ls", "--format", "{{json .}}").CombinedOutput()
		if err != nil {
			if isDaemonUnavailable(out) {
				return ErrMsg{fmt.Errorf("docker context ls: %w\n%s", ErrDaemonUnavailable, strings.TrimSpace(string(out)))}
			}
			return ErrMsg{fmt.Errorf("docker context ls: %w\n%s", err, strings.TrimSpace(string(out)))}
		}
		var contexts []DockerContext
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if line == "" {
				continue
			}
			var c DockerContext
			if err := json.Unmarshal([]byte(line), &c); err == nil {
				contexts = append(contexts, c)
			}
		}
		return ContextsMsg(contexts)
	}
}

func (CLI) SwitchContext(name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutContext)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", "context", "use", name).CombinedOutput()
		if err != nil {
			return ContextSwitchMsg{Err: fmt.Errorf("docker context use: %w\n%s", err, strings.TrimSpace(string(out)))}
		}
		return ContextSwitchMsg{}
	}
}

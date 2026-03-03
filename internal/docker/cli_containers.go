package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type CLI struct{}

func (CLI) FetchContainers(all bool) tea.Cmd {
	return func() tea.Msg {
		args := []string{"ps", "--format", "{{json .}}"}
		if all {
			args = append(args, "-a")
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeoutFetch)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
		if err != nil {
			if isDaemonUnavailable(out) {
				return ErrMsg{fmt.Errorf("docker ps: %w\n%s", ErrDaemonUnavailable, strings.TrimSpace(string(out)))}
			}
			return ErrMsg{fmt.Errorf("docker ps: %w\n%s", err, strings.TrimSpace(string(out)))}
		}
		var containers []Container
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if line == "" {
				continue
			}
			var c Container
			if err := json.Unmarshal([]byte(line), &c); err == nil {
				containers = append(containers, c)
			}
		}
		return ContainersMsg(containers)
	}
}

func (CLI) StopContainer(id string) tea.Cmd {
	return runContainerCmd(id, timeoutStop, "stop", func(err error) tea.Msg { return StopMsg{Err: err} })
}

func (CLI) StartContainer(id string) tea.Cmd {
	return runContainerCmd(id, timeoutStart, "start", func(err error) tea.Msg { return StartMsg{Err: err} })
}

func (CLI) RestartContainer(id string) tea.Cmd {
	return runContainerCmd(id, timeoutRestart, "restart", func(err error) tea.Msg { return RestartMsg{Err: err} })
}

func (CLI) DeleteContainer(id string) tea.Cmd {
	return runContainerCmd(id, timeoutRM, "rm", func(err error) tea.Msg { return DeleteMsg{ID: id, Err: err} })
}

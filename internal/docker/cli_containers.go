package docker

import (
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
		out, err := exec.Command("docker", args...).CombinedOutput()
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
	return runContainerCmd(id, "stop", func(err error) tea.Msg { return StopMsg{Err: err} })
}

func (CLI) StartContainer(id string) tea.Cmd {
	return runContainerCmd(id, "start", func(err error) tea.Msg { return StartMsg{Err: err} })
}

func (CLI) RestartContainer(id string) tea.Cmd {
	return runContainerCmd(id, "restart", func(err error) tea.Msg { return RestartMsg{Err: err} })
}

func (CLI) DeleteContainer(id string) tea.Cmd {
	return runContainerCmd(id, "rm", func(err error) tea.Msg { return DeleteMsg{ID: id, Err: err} })
}

func (CLI) PauseContainer(id string) tea.Cmd {
	return runContainerCmd(id, "pause", func(err error) tea.Msg { return PauseMsg{Err: err} })
}

func (CLI) UnpauseContainer(id string) tea.Cmd {
	return runContainerCmd(id, "unpause", func(err error) tea.Msg { return UnpauseMsg{Err: err} })
}

func (CLI) RenameContainer(id, newName string) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("docker", "rename", id, newName).CombinedOutput()
		if err != nil {
			return RenameMsg{Err: fmt.Errorf("docker rename: %w\n%s", err, strings.TrimSpace(string(out)))}
		}
		return RenameMsg{}
	}
}

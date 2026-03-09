package docker

import (
	"encoding/json"
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
				return ErrMsg{cmdErr("ps", out, ErrDaemonUnavailable)}
			}
			return ErrMsg{cmdErr("ps", out, err)}
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
			return RenameMsg{Err: cmdErr("rename", out, err)}
		}
		return RenameMsg{}
	}
}

func (CLI) StopCompose(project string) tea.Cmd {
	return runComposeCmd(project, "stop", func(err error) tea.Msg { return ComposeStopMsg{Err: err} })
}

func (CLI) StartCompose(project string) tea.Cmd {
	return runComposeCmd(project, "start", func(err error) tea.Msg { return ComposeStartMsg{Err: err} })
}

func (CLI) RestartCompose(project string) tea.Cmd {
	return runComposeCmd(project, "restart", func(err error) tea.Msg { return ComposeRestartMsg{Err: err} })
}

func runComposeCmd(project string, subcmd string, mkMsg func(error) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("docker", "compose", "-p", project, subcmd).CombinedOutput()
		if err != nil {
			return mkMsg(cmdErr("compose "+subcmd, out, err))
		}
		return mkMsg(nil)
	}
}

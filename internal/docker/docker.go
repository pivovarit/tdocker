package docker

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Container struct {
	ID         string `json:"ID"`
	Names      string `json:"Names"`
	Image      string `json:"Image"`
	State      string `json:"State"`
	Status     string `json:"Status"`
	RunningFor string `json:"RunningFor"`
	Ports      string `json:"Ports"`
}

type (
	ContainersMsg []Container
	ErrMsg        struct{ Err error }
	StopMsg       struct{ Err error }
	StartMsg      struct{ Err error }
	DeleteMsg     struct {
		ID  string
		Err error
	}
)

func FetchContainers(all bool) tea.Cmd {
	return func() tea.Msg {
		args := []string{"ps", "--format", "{{json .}}"}
		if all {
			args = append(args, "-a")
		}
		out, err := exec.Command("docker", args...).CombinedOutput()
		if err != nil {
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

func StopContainer(id string) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("docker", "stop", id).CombinedOutput()
		if err != nil {
			return StopMsg{fmt.Errorf("docker stop: %w\n%s", err, strings.TrimSpace(string(out)))}
		}
		return StopMsg{}
	}
}

func StartContainer(id string) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("docker", "start", id).CombinedOutput()
		if err != nil {
			return StartMsg{fmt.Errorf("docker start: %w\n%s", err, strings.TrimSpace(string(out)))}
		}
		return StartMsg{}
	}
}

func DeleteContainer(id string) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("docker", "rm", id).CombinedOutput()
		if err != nil {
			return DeleteMsg{ID: id, Err: fmt.Errorf("docker rm: %w\n%s", err, strings.TrimSpace(string(out)))}
		}
		return DeleteMsg{ID: id}
	}
}

func Sort(containers []Container) []Container {
	sorted := make([]Container, len(containers))
	copy(sorted, containers)
	sort.SliceStable(sorted, func(i, j int) bool {
		ri, rj := sorted[i].State == "running", sorted[j].State == "running"
		return ri && !rj
	})
	return sorted
}

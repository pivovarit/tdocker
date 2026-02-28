package docker

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Labels map[string]string

func (l *Labels) UnmarshalJSON(data []byte) error {
	var m map[string]string
	if err := json.Unmarshal(data, &m); err == nil {
		*l = m
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*l = make(Labels)
	for _, pair := range strings.Split(s, ",") {
		if k, v, ok := strings.Cut(pair, "="); ok {
			(*l)[k] = v
		}
	}
	return nil
}

type Container struct {
	ID         string `json:"ID"`
	Names      string `json:"Names"`
	Image      string `json:"Image"`
	State      string `json:"State"`
	Status     string `json:"Status"`
	RunningFor string `json:"RunningFor"`
	Ports      string `json:"Ports"`
	Labels     Labels `json:"Labels"`
}

func (c Container) ComposeProject() string {
	return c.Labels["com.docker.compose.project"]
}

func (c Container) ComposeService() string {
	return c.Labels["com.docker.compose.service"]
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
		ci, cj := sorted[i], sorted[j]
		ri, rj := ci.State == "running", cj.State == "running"
		if ri != rj {
			return ri
		}
		pi, pj := ci.ComposeProject(), cj.ComposeProject()
		if pi != pj {
			if pi == "" {
				return false
			}
			if pj == "" {
				return true
			}
			return pi < pj
		}
		si, sj := ci.ComposeService(), cj.ComposeService()
		if si != sj {
			return si < sj
		}
		return ci.Names < cj.Names
	})
	return sorted
}

package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	timeoutFetch   = 10 * time.Second
	timeoutStop    = 30 * time.Second
	timeoutStart   = 15 * time.Second
	timeoutRM      = 10 * time.Second
	timeoutDebug   = 5 * time.Second
	timeoutContext = 10 * time.Second
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
	ExecDoneMsg       struct{}
	DebugAvailableMsg struct {
		ID        string
		Available bool
	}
	ContextsMsg      []DockerContext
	ContextSwitchMsg struct{ Err error }
)

type DockerContext struct {
	Name           string `json:"Name"`
	Current        bool   `json:"Current"`
	Description    string `json:"Description"`
	DockerEndpoint string `json:"DockerEndpoint"`
}

func FetchContainers(all bool) tea.Cmd {
	return func() tea.Msg {
		args := []string{"ps", "--format", "{{json .}}"}
		if all {
			args = append(args, "-a")
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeoutFetch)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
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
		ctx, cancel := context.WithTimeout(context.Background(), timeoutStop)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", "stop", id).CombinedOutput()
		if err != nil {
			return StopMsg{fmt.Errorf("docker stop: %w\n%s", err, strings.TrimSpace(string(out)))}
		}
		return StopMsg{}
	}
}

func StartContainer(id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutStart)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", "start", id).CombinedOutput()
		if err != nil {
			return StartMsg{fmt.Errorf("docker start: %w\n%s", err, strings.TrimSpace(string(out)))}
		}
		return StartMsg{}
	}
}

func DeleteContainer(id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutRM)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", "rm", id).CombinedOutput()
		if err != nil {
			return DeleteMsg{ID: id, Err: fmt.Errorf("docker rm: %w\n%s", err, strings.TrimSpace(string(out)))}
		}
		return DeleteMsg{ID: id}
	}
}

func ExecContainer(id string) tea.Cmd {
	return tea.ExecProcess(
		exec.Command("docker", "exec", "-it", id, "sh"),
		func(_ error) tea.Msg { return ExecDoneMsg{} },
	)
}

func CheckDebugAvailable(id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutDebug)
		defer cancel()
		cmd := exec.CommandContext(ctx, "docker", "debug", "--help")
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		return DebugAvailableMsg{ID: id, Available: cmd.Run() == nil}
	}
}

func DebugContainer(id string) tea.Cmd {
	return tea.ExecProcess(
		exec.Command("docker", "debug", id),
		func(_ error) tea.Msg { return ExecDoneMsg{} },
	)
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

func FetchContexts() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutContext)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", "context", "ls", "--format", "{{json .}}").CombinedOutput()
		if err != nil {
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

func SwitchContext(name string) tea.Cmd {
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

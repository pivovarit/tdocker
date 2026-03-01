package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

var ErrDaemonUnavailable = errors.New("docker daemon unavailable")

func isDaemonUnavailable(out []byte) bool {
	s := string(out)
	return strings.Contains(s, "Cannot connect to the Docker daemon") ||
		strings.Contains(s, "Is the docker daemon running") ||
		strings.Contains(s, "connection refused") ||
		(strings.Contains(s, "no such file or directory") && strings.Contains(s, "docker.sock"))
}

const (
	timeoutFetch   = 10 * time.Second
	timeoutStop    = 30 * time.Second
	timeoutStart   = 15 * time.Second
	timeoutRestart = 30 * time.Second
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
	RestartMsg    struct{ Err error }
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

func runContainerCmd(id string, timeout time.Duration, subcmd string, mkMsg func(error) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", subcmd, id).CombinedOutput()
		if err != nil {
			return mkMsg(fmt.Errorf("docker %s: %w\n%s", subcmd, err, strings.TrimSpace(string(out))))
		}
		return mkMsg(nil)
	}
}

func StopContainer(id string) tea.Cmd {
	return runContainerCmd(id, timeoutStop, "stop", func(err error) tea.Msg { return StopMsg{Err: err} })
}

func StartContainer(id string) tea.Cmd {
	return runContainerCmd(id, timeoutStart, "start", func(err error) tea.Msg { return StartMsg{Err: err} })
}

func RestartContainer(id string) tea.Cmd {
	return runContainerCmd(id, timeoutRestart, "restart", func(err error) tea.Msg { return RestartMsg{Err: err} })
}

func DeleteContainer(id string) tea.Cmd {
	return runContainerCmd(id, timeoutRM, "rm", func(err error) tea.Msg { return DeleteMsg{ID: id, Err: err} })
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
	slices.SortStableFunc(sorted, func(a, b Container) int {
		ra, rb := a.State == "running", b.State == "running"
		if ra != rb {
			if ra {
				return -1
			}
			return 1
		}
		pa, pb := a.ComposeProject(), b.ComposeProject()
		if pa != pb {
			if pa == "" {
				return 1
			}
			if pb == "" {
				return -1
			}
			return strings.Compare(pa, pb)
		}
		sa, sb := a.ComposeService(), b.ComposeService()
		if sa != sb {
			return strings.Compare(sa, sb)
		}
		return strings.Compare(a.Names, b.Names)
	})
	return sorted
}

func FetchContexts() tea.Cmd {
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

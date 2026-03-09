package docker

import (
	"encoding/json"
	"errors"
	"fmt"
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
	timeoutDebug   = 5 * time.Second
	timeoutContext = 10 * time.Second
)

type Labels map[string]string

func (l *Labels) UnmarshalJSON(data []byte) error {
	var m map[string]string
	if objErr := json.Unmarshal(data, &m); objErr == nil {
		*l = m
		return nil
	} else {
		var s string
		if strErr := json.Unmarshal(data, &s); strErr != nil {
			return errors.Join(objErr, strErr)
		}
		*l = make(Labels)
		for _, pair := range strings.Split(s, ",") {
			if k, v, ok := strings.Cut(pair, "="); ok {
				(*l)[k] = v
			}
		}
		return nil
	}
}

type Container struct {
	ID         string `json:"ID"`
	Names      string `json:"Names"`
	Image      string `json:"Image"`
	Command    string `json:"Command"`
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

type LifecycleMsg interface {
	GetErr() error
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
	ExecDoneMsg       struct{ Err error }
	ShellAvailableMsg struct {
		ID        string
		Available bool
		Shell     string
	}
	DebugAvailableMsg struct {
		ID        string
		Available bool
	}
	ContextsMsg       []DockerContext
	ContextSwitchMsg  struct{ Err error }
	PauseMsg          struct{ Err error }
	UnpauseMsg        struct{ Err error }
	RenameMsg         struct{ Err error }
	ComposeStopMsg    struct{ Err error }
	ComposeStartMsg   struct{ Err error }
	ComposeRestartMsg struct{ Err error }
	GrepSupportMsg    struct{ Available bool }
	ExpandInspectMsg  struct {
		ContainerID string
		Data        *InspectData
		Err         error
	}
)

func (m StopMsg) GetErr() error           { return m.Err }
func (m StartMsg) GetErr() error          { return m.Err }
func (m RestartMsg) GetErr() error        { return m.Err }
func (m PauseMsg) GetErr() error          { return m.Err }
func (m UnpauseMsg) GetErr() error        { return m.Err }
func (m RenameMsg) GetErr() error         { return m.Err }
func (m ComposeStopMsg) GetErr() error    { return m.Err }
func (m ComposeStartMsg) GetErr() error   { return m.Err }
func (m ComposeRestartMsg) GetErr() error { return m.Err }

type DockerContext struct {
	Name           string `json:"Name"`
	Current        bool   `json:"Current"`
	Description    string `json:"Description"`
	DockerEndpoint string `json:"DockerEndpoint"`
}

func cmdErr(subcmd string, out []byte, err error) error {
	return fmt.Errorf("docker %s: %w\n%s", subcmd, err, strings.TrimSpace(string(out)))
}

func runContainerCmd(id string, subcmd string, mkMsg func(error) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("docker", subcmd, id).CombinedOutput()
		if err != nil {
			return mkMsg(cmdErr(subcmd, out, err))
		}
		return mkMsg(nil)
	}
}

func execErr(err error) error {
	if err == nil {
		return nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		switch exitErr.ExitCode() {
		case 126, 127:
			return fmt.Errorf("shell not found in container (distroless/scratch image?) - press 'x' to use docker debug")
		}
	}
	return err
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

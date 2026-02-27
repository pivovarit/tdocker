package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	// LogsLineMsg carries one log line and the cmd to read the next one.
	LogsLineMsg struct {
		Line string
		Next tea.Cmd
	}
	LogsEndMsg struct{ Err error }
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

// StartLogs starts streaming docker logs for the given container.
// It returns the first tea.Cmd to kick off the read loop and a stop function
// that kills the underlying process.
func StartLogs(id string) (tea.Cmd, func()) {
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, "docker", "logs", "--follow", "--tail", "200", id)
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		cancel()
		return func() tea.Msg { return LogsEndMsg{err} }, func() {}
	}

	go func() {
		cmd.Wait()
		pw.Close()
	}()

	scanner := bufio.NewScanner(pr)

	var readNext tea.Cmd
	readNext = func() tea.Msg {
		if scanner.Scan() {
			return LogsLineMsg{Line: scanner.Text(), Next: readNext}
		}
		return LogsEndMsg{scanner.Err()}
	}

	stop := func() {
		cancel()
		pr.Close()
	}

	return readNext, stop
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

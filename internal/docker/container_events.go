package docker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	tea "charm.land/bubbletea/v2"
)

const timeoutContainerEvents = 10 * time.Second

type ContainerEventsMsg struct {
	ContainerID string
	Events      []Event
	Err         error
}

func (CLI) FetchContainerEvents(id string, since time.Duration) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutContainerEvents)
		defer cancel()

		sinceArg := fmt.Sprintf("%ds", int(since.Seconds()))
		cmd := exec.CommandContext(ctx, "docker", "events",
			"--since", sinceArg,
			"--until", "0s",
			"--filter", "container="+id,
			"--format", "{{json .}}",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			if isDaemonUnavailable(out) {
				return ContainerEventsMsg{ContainerID: id, Err: cmdErr("events", out, ErrDaemonUnavailable)}
			}
			return ContainerEventsMsg{ContainerID: id, Err: cmdErr("events", out, err)}
		}

		var events []Event
		sc := bufio.NewScanner(bytes.NewReader(out))
		for sc.Scan() {
			var ev Event
			if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
				continue
			}
			events = append(events, ev)
		}
		return ContainerEventsMsg{ContainerID: id, Events: events}
	}
}

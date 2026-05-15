package docker

import (
	"context"
	"encoding/json"
	"os/exec"
	"time"

	tea "charm.land/bubbletea/v2"
)

type EventActor struct {
	ID         string            `json:"ID"`
	Attributes map[string]string `json:"Attributes"`
}

type Event struct {
	Type   string     `json:"Type"`
	Action string     `json:"Action"`
	Actor  EventActor `json:"Actor"`
	Time   int64      `json:"time"`
}

func (e Event) Name() string {
	if n := e.Actor.Attributes["name"]; n != "" {
		return n
	}
	if id := e.Actor.ID; len(id) > 12 {
		return id[:12]
	}
	return e.Actor.ID
}

func (e Event) Timestamp() string {
	return time.Unix(e.Time, 0).Format("15:04:05")
}

type (
	EventLineMsg struct {
		Event Event
		Next  tea.Cmd
		Gen   int
	}
	EventEndMsg struct {
		Err error
		Gen int
	}
)

func (CLI) StartEvents(ctx context.Context, gen int) tea.Cmd {
	return streamCmd(ctx, exec.CommandContext(ctx, "docker", "events", "--format", "{{json .}}"),
		func(line string, next tea.Cmd) tea.Msg {
			var ev Event
			if err := json.Unmarshal([]byte(line), &ev); err != nil {
				return nil
			}
			return EventLineMsg{Event: ev, Next: next, Gen: gen}
		},
		func(err error) tea.Msg { return EventEndMsg{Err: err, Gen: gen} },
	)
}

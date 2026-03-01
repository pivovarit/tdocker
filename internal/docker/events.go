package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
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

func StartEvents(ctx context.Context, gen int) tea.Cmd {
	cmd := exec.CommandContext(ctx, "docker", "events", "--format", "{{json .}}")
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		return func() tea.Msg { return EventEndMsg{Err: err, Gen: gen} }
	}

	go func() {
		err := cmd.Wait()
		contextCancelled := ctx.Err() != nil
		if err != nil && !contextCancelled {
			if cerr := pw.CloseWithError(err); cerr != nil {
				log.Printf("pipe close: %v", cerr)
			}
		} else {
			if cerr := pw.Close(); cerr != nil {
				log.Printf("pipe close: %v", cerr)
			}
		}
	}()

	scanner := bufio.NewScanner(pr)

	var readNext tea.Cmd
	readNext = func() tea.Msg {
		if scanner.Scan() {
			var ev Event
			if err := json.Unmarshal([]byte(scanner.Text()), &ev); err != nil {
				return readNext()
			}
			return EventLineMsg{Event: ev, Next: readNext, Gen: gen}
		}
		return EventEndMsg{Err: scanner.Err(), Gen: gen}
	}

	return readNext
}

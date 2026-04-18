package docker

import (
	"testing"
	"time"
)

func TestContainerEventsMsg_Shape(t *testing.T) {
	msg := ContainerEventsMsg{
		ContainerID: "abc123",
		Events: []Event{
			{Type: "container", Action: "die", Time: time.Now().Unix()},
		},
	}
	if msg.ContainerID != "abc123" {
		t.Errorf("ContainerID = %q", msg.ContainerID)
	}
	if len(msg.Events) != 1 {
		t.Errorf("len(Events) = %d", len(msg.Events))
	}
}

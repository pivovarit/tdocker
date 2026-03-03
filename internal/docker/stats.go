package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

const timeoutStats = 10 * time.Second

type StatsEntry struct {
	CPUPerc  string `json:"CPUPerc"`
	MemUsage string `json:"MemUsage"`
	MemPerc  string `json:"MemPerc"`
	NetIO    string `json:"NetIO"`
	BlockIO  string `json:"BlockIO"`
	PIDs     string `json:"PIDs"`
}

type StatsMsg struct {
	Entry StatsEntry
	Err   error
}

func (CLI) FetchStats(id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutStats)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format", "{{json .}}", id).CombinedOutput()
		if err != nil {
			return StatsMsg{Err: fmt.Errorf("docker stats: %w\n%s", err, strings.TrimSpace(string(out)))}
		}
		line := strings.TrimSpace(string(out))
		var entry StatsEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return StatsMsg{Err: fmt.Errorf("parse stats: %w", err)}
		}
		return StatsMsg{Entry: entry}
	}
}

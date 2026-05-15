package docker

import (
	"context"
	"encoding/json"
	"os/exec"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"
)

var ansiEscRe = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

type StatsEntry struct {
	ID       string `json:"ID"`
	Name     string `json:"Name"`
	CPUPerc  string `json:"CPUPerc"`
	MemUsage string `json:"MemUsage"`
	MemPerc  string `json:"MemPerc"`
	NetIO    string `json:"NetIO"`
	BlockIO  string `json:"BlockIO"`
	PIDs     string `json:"PIDs"`
}

type (
	StatsLineMsg struct {
		Entry StatsEntry
		Next  tea.Cmd
		Gen   int
	}
	StatsEndMsg struct {
		Err error
		Gen int
	}
)

func (CLI) StartAllStats(ctx context.Context, gen int) tea.Cmd {
	return streamCmd(ctx, exec.CommandContext(ctx, "docker", "stats", "--format", "{{json .}}"),
		func(line string, next tea.Cmd) tea.Msg {
			line = strings.TrimSpace(ansiEscRe.ReplaceAllString(line, ""))
			if line == "" {
				return nil
			}
			var e StatsEntry
			if err := json.Unmarshal([]byte(line), &e); err != nil {
				return nil
			}
			return StatsLineMsg{Entry: e, Next: next, Gen: gen}
		},
		func(err error) tea.Msg { return StatsEndMsg{Err: err, Gen: gen} },
	)
}

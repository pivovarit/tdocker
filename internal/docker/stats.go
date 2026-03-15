package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
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
	cmd := exec.CommandContext(ctx, "docker", "stats", "--format", "{{json .}}")
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		return func() tea.Msg { return StatsEndMsg{Err: err, Gen: gen} }
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
		for scanner.Scan() {
			line := strings.TrimSpace(ansiEscRe.ReplaceAllString(scanner.Text(), ""))
			if line == "" {
				continue
			}
			var e StatsEntry
			if err := json.Unmarshal([]byte(line), &e); err != nil {
				continue
			}
			return StatsLineMsg{Entry: e, Next: readNext, Gen: gen}
		}
		return StatsEndMsg{Err: scanner.Err(), Gen: gen}
	}

	return readNext
}

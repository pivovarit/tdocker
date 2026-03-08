package docker

import (
	"bufio"
	"context"
	"io"
	"log"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type (
	LogsLineMsg struct {
		Line string
		Next tea.Cmd
		Gen  int
	}
	LogsEndMsg struct {
		Err error
		Gen int
	}
)

func (CLI) SupportsGrep() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutDebug)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", "logs", "--help").CombinedOutput()
		if err != nil {
			return GrepSupportMsg{Available: false}
		}
		return GrepSupportMsg{Available: strings.Contains(string(out), "--grep")}
	}
}

func (CLI) StartLogs(ctx context.Context, id string, tail string, timestamps bool, grep string, gen int) tea.Cmd {
	args := []string{"logs", "--follow", "--tail", tail}
	if timestamps {
		args = append(args, "--timestamps")
	}
	if grep != "" {
		args = append(args, "--grep", grep)
	}
	args = append(args, id)
	cmd := exec.CommandContext(ctx, "docker", args...)
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		return func() tea.Msg { return LogsEndMsg{Err: err, Gen: gen} }
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
			return LogsLineMsg{Line: scanner.Text(), Next: readNext, Gen: gen}
		}
		return LogsEndMsg{Err: scanner.Err(), Gen: gen}
	}

	return readNext
}

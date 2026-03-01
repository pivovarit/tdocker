package docker

import (
	"bufio"
	"context"
	"io"
	"log"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
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

func StartLogs(ctx context.Context, id string, tail string, gen int) tea.Cmd {
	cmd := exec.CommandContext(ctx, "docker", "logs", "--follow", "--tail", tail, id)
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

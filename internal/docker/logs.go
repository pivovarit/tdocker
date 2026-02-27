package docker

import (
	"bufio"
	"context"
	"io"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	LogsLineMsg struct {
		Line string
		Next tea.Cmd
	}
	LogsEndMsg struct{ Err error }
)

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

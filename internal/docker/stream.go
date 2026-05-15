package docker

import (
	"bufio"
	"context"
	"io"
	"log"
	"os/exec"

	tea "charm.land/bubbletea/v2"
)

func streamCmd(
	ctx context.Context,
	cmd *exec.Cmd,
	parse func(line string, next tea.Cmd) tea.Msg,
	endMsg func(err error) tea.Msg,
) tea.Cmd {
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		return func() tea.Msg { return endMsg(err) }
	}

	go func() {
		err := cmd.Wait()
		if err != nil && ctx.Err() == nil {
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
			if msg := parse(scanner.Text(), readNext); msg != nil {
				return msg
			}
		}
		return endMsg(scanner.Err())
	}

	return readNext
}

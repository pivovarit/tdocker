package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/pivovarit/tdocker/internal/docker"
)

func buildTable(containers []docker.Container, width int) table.Model {
	const (
		idW         = 13
		stateW      = 9
		runningForW = 13
		overhead    = 16
	)

	nameW, imageW, statusW, portsW := 4, 5, 6, 5
	for _, c := range containers {
		if w := len([]rune(strings.TrimPrefix(c.Names, "/"))); w > nameW {
			nameW = w
		}
		if w := len([]rune(c.Image)); w > imageW {
			imageW = w
		}
		if w := len([]rune(c.Status)); w > statusW {
			statusW = w
		}
		if w := len([]rune(c.Ports)); w > portsW {
			portsW = w
		}
	}

	nameW = min(nameW, 30)
	imageW = min(imageW, 40)
	statusW = min(statusW, 20)
	portsW = min(portsW, 60)

	remaining := width - idW - stateW - runningForW - overhead
	if remaining < 40 {
		remaining = 40
	}

	total := nameW + imageW + statusW + portsW
	nameW = max(remaining*nameW/total, 4)
	imageW = max(remaining*imageW/total, 5)
	statusW = max(remaining*statusW/total, 6)
	portsW = max(remaining-nameW-imageW-statusW, 5)

	cols := []table.Column{
		{Title: "ID", Width: idW},
		{Title: "Name", Width: nameW},
		{Title: "Image", Width: imageW},
		{Title: "State", Width: stateW},
		{Title: "Status", Width: statusW},
		{Title: "Running for", Width: runningForW},
		{Title: "Ports", Width: portsW},
	}

	rows := make([]table.Row, len(containers))
	for i, c := range containers {
		rows[i] = table.Row{
			trunc(c.ID, idW),
			trunc(c.Names, nameW),
			trunc(c.Image, imageW),
			trunc(c.State, stateW),
			trunc(c.Status, statusW),
			trunc(c.RunningFor, runningForW),
			trunc(c.Ports, portsW),
		}
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#0369A1")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("#38BDF8"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#F0F9FF")).
		Background(lipgloss.Color("#0369A1")).
		Bold(false)
	t.SetStyles(s)

	return t
}

func trunc(s string, max int) string {
	if max < 1 {
		return ""
	}
	s = strings.TrimPrefix(s, "/")
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	return string(runes[:max-1]) + "…"
}

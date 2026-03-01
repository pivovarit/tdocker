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

	nameW, imageW, statusW, portsW := 4, 5, 6, 0
	hasPorts := false
	for i, c := range containers {
		if w := len([]rune(buildTableName(containers, i))); w > nameW {
			nameW = w
		}
		if w := len([]rune(c.Image)); w > imageW {
			imageW = w
		}
		if w := len([]rune(c.Status)); w > statusW {
			statusW = w
		}
		if c.Ports != "" {
			hasPorts = true
			if w := len([]rune(c.Ports)); w > portsW {
				portsW = w
			}
		}
	}

	nameW = min(nameW, 50)
	imageW = min(imageW, 40)
	statusW = min(statusW, 20)

	remaining := width - idW - stateW - runningForW - overhead
	if remaining < 40 {
		remaining = 40
	}

	leftover := max(remaining-nameW, 20)
	if hasPorts {
		portsW = min(portsW, 60)
		total := imageW + statusW + portsW
		imageW = max(leftover*imageW/total, 5)
		statusW = max(leftover*statusW/total, 6)
		portsW = max(leftover-imageW-statusW, 5)
	} else {
		total := imageW + statusW
		imageW = max(leftover*imageW/total, 5)
		statusW = max(leftover-imageW, 6)
	}

	cols := []table.Column{
		{Title: "ID", Width: idW},
		{Title: "Name", Width: nameW},
		{Title: "Image", Width: imageW},
		{Title: "State", Width: stateW},
		{Title: "Status", Width: statusW},
		{Title: "Running for", Width: runningForW},
	}
	if hasPorts {
		cols = append(cols, table.Column{Title: "Ports", Width: portsW})
	}

	rows := make([]table.Row, len(containers))
	for i, c := range containers {
		row := table.Row{
			trunc(c.ID, idW),
			trunc(buildTableName(containers, i), nameW),
			trunc(c.Image, imageW),
			trunc(c.State, stateW),
			trunc(c.Status, statusW),
			trunc(c.RunningFor, runningForW),
		}
		if hasPorts {
			row = append(row, trunc(c.Ports, portsW))
		}
		rows[i] = row
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

func buildTableName(containers []docker.Container, i int) string {
	c := containers[i]
	p := c.ComposeProject()
	if p == "" {
		return strings.TrimPrefix(c.Names, "/")
	}
	s := c.ComposeService()
	if s == "" {
		s = strings.TrimPrefix(c.Names, "/")
	}
	label := p + "/" + s
	prev := i > 0 && containers[i-1].ComposeProject() == p
	next := i < len(containers)-1 && containers[i+1].ComposeProject() == p
	switch {
	case !prev && next:
		return "┬ " + label
	case prev && next:
		return "├ " + label
	case prev:
		return "└ " + label
	default:
		return label
	}
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

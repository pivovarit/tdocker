package ui

import (
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

func buildTable(containers []docker.Container, width int) table.Model {
	const (
		idW         = 13
		commandW    = 20
		runningForW = 13
		overhead    = 15
	)

	names := make([]string, len(containers))
	nameW, imageW, statusW, portsW := 5, 5, 6, 0
	hasPorts := false
	for i, c := range containers {
		names[i] = buildTableName(containers, i)
		if w := len([]rune(names[i])); w > nameW {
			nameW = w
		}
		if w := len([]rune(c.Image)); w > imageW {
			imageW = w
		}
		if w := len([]rune(c.Status)); w > statusW {
			statusW = w
		}
		if fp := formatPorts(c.Ports); fp != "" {
			hasPorts = true
			if w := len([]rune(fp)); w > portsW {
				portsW = w
			}
		}
	}

	remaining := width - idW - commandW - runningForW - overhead

	if hasPorts {
		minR := 5 + 5 + 6 + 5
		if remaining < minR {
			remaining = minR
		}
		total := imageW + statusW + portsW + nameW
		imageW = max(remaining*imageW/total, 5)
		statusW = max(remaining*statusW/total, 6)
		portsW = max(remaining*portsW/total, 5)
		nameW = max(remaining*nameW/total, 5)
		if leftover := remaining - imageW - statusW - portsW - nameW; leftover > 0 {
			nameW += leftover
		}
	} else {
		minR := 5 + 6 + 5
		if remaining < minR {
			remaining = minR
		}
		total := imageW + statusW + nameW
		imageW = max(remaining*imageW/total, 5)
		statusW = max(remaining*statusW/total, 6)
		nameW = max(remaining-imageW-statusW, 5)
	}

	cols := []table.Column{
		{Title: "ID", Width: idW},
		{Title: "Names", Width: nameW},
		{Title: "Image", Width: imageW},
		{Title: "Command", Width: commandW},
		{Title: "Created", Width: runningForW},
		{Title: "Status", Width: statusW},
	}
	if hasPorts {
		cols = append(cols, table.Column{Title: "Ports", Width: portsW})
	}

	rows := make([]table.Row, len(containers))
	for i, c := range containers {
		row := table.Row{
			trunc(c.ID, idW),
			trunc(names[i], nameW),
			trunc(c.Image, imageW),
			trunc(c.Command, commandW),
			trunc(c.RunningFor, runningForW),
			trunc(c.Status, statusW),
		}
		if hasPorts {
			row = append(row, trunc(formatPorts(c.Ports), portsW))
		}
		rows[i] = row
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithWidth(width),
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
	if c.State == "collapsed" {
		return c.Names
	}
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

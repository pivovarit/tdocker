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

	treeChars := make([]string, len(containers))
	hasTree := false
	for i := range containers {
		treeChars[i] = composeTreeChar(containers, i)
		if treeChars[i] != "" {
			hasTree = true
		}
	}

	actualIDW := idW
	if hasTree {
		actualIDW = idW + 2
	}

	remaining := width - actualIDW - commandW - runningForW - overhead

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
		{Title: "ID", Width: actualIDW},
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
		id := trunc(c.ID, actualIDW)
		if ch := treeChars[i]; ch != "" {
			var short string
			if c.State == "collapsed" {
				const placeholder = "—"
				pad := idW - len([]rune(placeholder))
				left := pad / 2
				right := pad - left
				short = strings.Repeat(" ", left) + placeholder + strings.Repeat(" ", right)
			} else {
				short = trunc(c.ID, idW)
				if pad := idW - len([]rune(short)); pad > 0 {
					short += strings.Repeat(" ", pad)
				}
			}
			id = short + " " + ch
		} else if hasTree && c.State == "detail" {
			if ch := detailTreeChar(containers, i); ch != "" {
				id = strings.Repeat(" ", idW) + " " + ch
			}
		}
		row := table.Row{
			id,
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
	if c.State == "detail" {
		return c.Names
	}
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
	return p + "/" + s
}

func composeTreeChar(containers []docker.Container, i int) string {
	c := containers[i]
	if c.State == "detail" {
		return ""
	}
	if c.State == "collapsed" {
		tree := func(ch string) string { return "\x1b[38;2;100;116;139m" + ch + "\x1b[39m" }
		return tree("▸")
	}
	if c.ComposeProject() == "" {
		return ""
	}
	if !inComposeGroup(containers, i) {
		return ""
	}
	p := c.ComposeProject()
	tree := func(ch string) string { return "\x1b[38;2;100;116;139m" + ch + "\x1b[39m" }

	prevIdx := i - 1
	for prevIdx >= 0 && containers[prevIdx].State == "detail" {
		prevIdx--
	}
	hasPrev := prevIdx >= 0 && containers[prevIdx].ComposeProject() == p

	nextIdx := i + 1
	for nextIdx < len(containers) && containers[nextIdx].State == "detail" {
		nextIdx++
	}
	hasNext := nextIdx < len(containers) && containers[nextIdx].ComposeProject() == p

	switch {
	case !hasPrev:
		return tree("┬")
	case !hasNext:
		if i+1 < len(containers) && containers[i+1].State == "detail" {
			return tree("├")
		}
		return tree("└")
	default:
		return tree("├")
	}
}

func detailTreeChar(containers []docker.Container, i int) string {
	parentIdx := i - 1
	for parentIdx >= 0 && containers[parentIdx].State == "detail" {
		parentIdx--
	}
	if parentIdx < 0 || containers[parentIdx].ComposeProject() == "" {
		return ""
	}
	if !inComposeGroup(containers, parentIdx) {
		return ""
	}
	proj := containers[parentIdx].ComposeProject()
	nextIdx := i + 1
	for nextIdx < len(containers) && containers[nextIdx].State == "detail" {
		nextIdx++
	}
	if nextIdx < len(containers) && containers[nextIdx].ComposeProject() == proj {
		return "│"
	}
	isLastDetail := i+1 >= len(containers) || containers[i+1].State != "detail"
	if isLastDetail {
		return "└"
	}
	return "│"
}

func inComposeGroup(containers []docker.Container, i int) bool {
	p := containers[i].ComposeProject()
	if p == "" {
		return false
	}
	prevIdx := i - 1
	for prevIdx >= 0 && containers[prevIdx].State == "detail" {
		prevIdx--
	}
	if prevIdx >= 0 && containers[prevIdx].ComposeProject() == p {
		return true
	}
	nextIdx := i + 1
	for nextIdx < len(containers) && containers[nextIdx].State == "detail" {
		nextIdx++
	}
	return nextIdx < len(containers) && containers[nextIdx].ComposeProject() == p
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

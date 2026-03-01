package docker

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

type Client interface {
	FetchContainers(all bool) tea.Cmd
	StopContainer(id string) tea.Cmd
	StartContainer(id string) tea.Cmd
	DeleteContainer(id string) tea.Cmd
	ExecContainer(id string) tea.Cmd
	CheckDebugAvailable(id string) tea.Cmd
	DebugContainer(id string) tea.Cmd
	InspectContainer(id string) tea.Cmd
	FetchStats(id string) tea.Cmd
	StartLogs(ctx context.Context, id string, tail string, gen int) tea.Cmd
}

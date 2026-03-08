package docker

import (
	"context"

	tea "charm.land/bubbletea/v2"
)

type Client interface {
	FetchContainers(all bool) tea.Cmd
	StopContainer(id string) tea.Cmd
	StartContainer(id string) tea.Cmd
	RestartContainer(id string) tea.Cmd
	DeleteContainer(id string) tea.Cmd
	CheckShellAvailable(id string) tea.Cmd
	ExecContainer(id, shell string) tea.Cmd
	CheckDebugAvailable(id string) tea.Cmd
	DebugContainer(id string) tea.Cmd
	InspectContainer(id string) tea.Cmd
	FetchStats(id string) tea.Cmd
	StartLogs(ctx context.Context, id string, tail string, timestamps bool, gen int) tea.Cmd
	StartEvents(ctx context.Context, gen int) tea.Cmd
	FetchContexts() tea.Cmd
	SwitchContext(name string) tea.Cmd
	PauseContainer(id string) tea.Cmd
	UnpauseContainer(id string) tea.Cmd
}

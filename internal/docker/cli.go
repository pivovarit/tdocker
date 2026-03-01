package docker

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

type CLI struct{}

func (CLI) FetchContainers(all bool) tea.Cmd      { return FetchContainers(all) }
func (CLI) StopContainer(id string) tea.Cmd       { return StopContainer(id) }
func (CLI) StartContainer(id string) tea.Cmd      { return StartContainer(id) }
func (CLI) DeleteContainer(id string) tea.Cmd     { return DeleteContainer(id) }
func (CLI) ExecContainer(id string) tea.Cmd       { return ExecContainer(id) }
func (CLI) CheckDebugAvailable(id string) tea.Cmd { return CheckDebugAvailable(id) }
func (CLI) DebugContainer(id string) tea.Cmd      { return DebugContainer(id) }
func (CLI) InspectContainer(id string) tea.Cmd    { return InspectContainer(id) }
func (CLI) FetchStats(id string) tea.Cmd          { return FetchStats(id) }
func (CLI) StartLogs(ctx context.Context, id string, tail string, gen int) tea.Cmd {
	return StartLogs(ctx, id, tail, gen)
}

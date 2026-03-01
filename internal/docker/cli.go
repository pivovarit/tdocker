package docker

import (
	"context"

	tea "charm.land/bubbletea/v2"
)

func (CLI) InspectContainer(id string) tea.Cmd { return InspectContainer(id) }
func (CLI) FetchStats(id string) tea.Cmd       { return FetchStats(id) }
func (CLI) StartLogs(ctx context.Context, id string, tail string, gen int) tea.Cmd {
	return StartLogs(ctx, id, tail, gen)
}
func (CLI) StartEvents(ctx context.Context, gen int) tea.Cmd { return StartEvents(ctx, gen) }

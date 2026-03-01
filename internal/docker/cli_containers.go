package docker

import tea "charm.land/bubbletea/v2"

type CLI struct{}

func (CLI) FetchContainers(all bool) tea.Cmd   { return FetchContainers(all) }
func (CLI) StopContainer(id string) tea.Cmd    { return StopContainer(id) }
func (CLI) StartContainer(id string) tea.Cmd   { return StartContainer(id) }
func (CLI) RestartContainer(id string) tea.Cmd { return RestartContainer(id) }
func (CLI) DeleteContainer(id string) tea.Cmd  { return DeleteContainer(id) }

package docker

import tea "charm.land/bubbletea/v2"

func (CLI) FetchContexts() tea.Cmd            { return FetchContexts() }
func (CLI) SwitchContext(name string) tea.Cmd { return SwitchContext(name) }

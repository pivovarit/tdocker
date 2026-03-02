package docker

import tea "charm.land/bubbletea/v2"

func (CLI) CheckShellAvailable(id string) tea.Cmd { return CheckShellAvailable(id) }
func (CLI) ExecContainer(id string) tea.Cmd       { return ExecContainer(id) }
func (CLI) CheckDebugAvailable(id string) tea.Cmd { return CheckDebugAvailable(id) }
func (CLI) DebugContainer(id string) tea.Cmd      { return DebugContainer(id) }

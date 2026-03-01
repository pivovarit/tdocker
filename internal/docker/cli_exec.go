package docker

import tea "github.com/charmbracelet/bubbletea"

func (CLI) ExecContainer(id string) tea.Cmd       { return ExecContainer(id) }
func (CLI) CheckDebugAvailable(id string) tea.Cmd { return CheckDebugAvailable(id) }
func (CLI) DebugContainer(id string) tea.Cmd      { return DebugContainer(id) }

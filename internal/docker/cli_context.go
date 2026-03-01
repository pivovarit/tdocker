package docker

import tea "github.com/charmbracelet/bubbletea"

func (CLI) FetchContexts() tea.Cmd            { return FetchContexts() }
func (CLI) SwitchContext(name string) tea.Cmd { return SwitchContext(name) }

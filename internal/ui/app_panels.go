package ui

func (m App) tableHeight() int {
	reserved := tableChrome
	if m.logs.visible {
		reserved += logsPanelHeight
	}
	if m.inspect.visible {
		reserved += inspectPanelHeight
	}
	if m.stats.visible {
		reserved += statsPanelHeight
	}
	if m.events.visible {
		reserved += eventsPanelHeight
	}
	h := m.height - reserved
	if h < 3 {
		return 3
	}
	return h
}

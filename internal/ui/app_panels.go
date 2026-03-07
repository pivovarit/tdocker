package ui

func (m App) tableHeight() int {
	reserved := tableChrome
	if m.inspect.visible {
		reserved += m.inspectPanelHeight()
	}
	if m.stats.visible {
		reserved += statsPanelHeight
	}
	if m.events.visible {
		reserved += m.eventsPanelHeight()
	}
	if m.ctxPicker.visible {
		reserved += min(len(m.ctxPicker.contexts), ctxPanelMaxRows) + 2 // +2 for divider + title
	}
	h := m.height - reserved
	if h < 3 {
		return 3
	}
	return h
}

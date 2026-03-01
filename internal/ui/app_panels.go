package ui

func (m App) closeLogs() App {
	if m.logs.cancel != nil {
		m.logs.cancel()
	}
	m.logs = logsState{scroll: scrollState{autoScroll: true}}
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) closeInspect() App {
	m.inspect = inspectState{}
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) closeStats() App {
	m.stats = statsState{}
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) closeEvents() App {
	m.events = eventsState{scroll: scrollState{autoScroll: true}}
	m.table.SetHeight(m.tableHeight())
	return m
}

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

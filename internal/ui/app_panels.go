package ui

func (m App) closeLogs() App {
	if m.logsCancel != nil {
		m.logsCancel()
		m.logsCancel = nil
	}
	m.logsVisible = false
	m.logsLines = nil
	m.logsContainer = ""
	m.logsContainerID = ""
	m.logsScrollOffset = 0
	m.logsAutoScroll = true
	m.logsAllMode = false
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) closeInspect() App {
	m.inspectVisible = false
	m.inspectLines = nil
	m.inspectContainer = ""
	m.inspectOffset = 0
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) closeStats() App {
	m.statsVisible = false
	m.statsEntry = nil
	m.statsPrevEntry = nil
	m.statsContainer = ""
	m.statsContainerID = ""
	m.statsFetching = false
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) closeEvents() App {
	if m.eventsCancel != nil {
		m.eventsCancel()
		m.eventsCancel = nil
	}
	m.eventsVisible = false
	m.eventsEvents = nil
	m.eventsScrollOffset = 0
	m.eventsAutoScroll = true
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) tableHeight() int {
	reserved := tableChrome
	if m.logsVisible {
		reserved += logsPanelHeight
	}
	if m.inspectVisible {
		reserved += inspectPanelHeight
	}
	if m.statsVisible {
		reserved += statsPanelHeight
	}
	if m.eventsVisible {
		reserved += eventsPanelHeight
	}
	h := m.height - reserved
	if h < 3 {
		return 3
	}
	return h
}

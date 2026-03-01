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
	m.statsContainer = ""
	m.statsContainerID = ""
	m.statsFetching = false
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) tableHeight() int {
	reserved := 8
	if m.logsVisible {
		reserved += logsPanelHeight
	}
	if m.inspectVisible {
		reserved += inspectPanelHeight
	}
	if m.statsVisible {
		reserved += statsPanelHeight
	}
	h := m.height - reserved
	if h < 3 {
		return 3
	}
	return h
}

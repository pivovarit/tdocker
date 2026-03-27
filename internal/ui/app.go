package ui

import (
	"context"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

func trimLastRune(s string) string {
	_, size := utf8.DecodeLastRuneInString(s)
	return s[:len(s)-size]
}

const (
	statsRows        = 5
	statsPanelHeight = statsRows + 3
	ctxPanelMaxRows  = 8
	logsTailDefault  = "200"

	chromeTitle        = 1
	chromeTitleMargin  = 1
	chromeTitleNewline = 1
	chromeBorderTop    = 1
	chromeBorderBottom = 1
	chromeHelpNewline  = 1
	chromeHelpMargin   = 1
	chromeHelp         = 1

	tableChrome = chromeTitle + chromeTitleMargin + chromeTitleNewline +
		chromeBorderTop + chromeBorderBottom +
		chromeHelpNewline + chromeHelpMargin + chromeHelp
)

type Operation int

const (
	OpNone Operation = iota
	OpConfirming
	OpStopping
	OpStarting
	OpRestarting
	OpDeleting
	OpPausing
	OpUnpausing
	OpRenaming
)

type operationState struct {
	kind    Operation
	visible bool
	gen     int
	action  string
	id      string
	name    string
}

type fetchState struct {
	start   time.Time
	gen     int
	slow    bool
	loading bool
	visible bool
}

type renameState struct {
	active bool
	id     string
	input  string
}

type App struct {
	client         docker.Client
	table          table.Model
	containers     []docker.Container
	sorted         []docker.Container
	containersByID map[string]docker.Container
	viewportStart  int
	showAll        bool
	filtering      bool
	filterQuery    string
	err            error
	width          int
	height         int

	collapsedProjects  map[string]bool
	expandedContainers map[string]*docker.InspectData

	op     operationState
	fetch  fetchState
	rename renameState

	logs      logsState
	inspect   inspectState
	stats     statsState
	events    eventsState
	ctxPicker ctxPickerState

	inlineStats       map[string]docker.StatsEntry
	showInlineStats   bool
	bgStatsGen        int
	statsDirty        bool
	statsPendingFlush bool
	copiedName        string
	warnMsg           string
	version           string
	updateAvailable   string
	bgEventsGen       int
	pendingRefresh    bool
	helpVisible       bool
	grepSupported     bool
}

func New(version string) App {
	return newWithClient(docker.CLI{}, version)
}

func newWithClient(c docker.Client, version string) App {
	return App{
		client:             c,
		version:            version,
		showAll:            true,
		collapsedProjects:  map[string]bool{},
		expandedContainers: map[string]*docker.InspectData{},
		inlineStats:        map[string]docker.StatsEntry{},
		bgStatsGen:         1,
		table:              buildTable(nil, 120, nil),
		fetch: fetchState{
			loading: true,
			start:   time.Now(),
			gen:     1,
		},
		logs:        logsState{scroll: scrollState{autoScroll: true}},
		events:      eventsState{scroll: scrollState{autoScroll: true}},
		bgEventsGen: 1,
	}
}

func (m App) Init() tea.Cmd {
	return tea.Batch(
		m.client.FetchContainers(m.showAll),
		m.client.StartAllStats(context.Background(), m.bgStatsGen),
		m.client.FetchContexts(),
		m.client.StartEvents(context.Background(), m.bgEventsGen),
		m.client.SupportsGrep(),
		fetchTimerCmd(),
		fetchSlowCmd(m.fetch.gen),
		checkUpdateCmd(m.version),
		periodicRefreshCmd(),
	)
}

func matchesFilter(c docker.Container, q string) bool {
	return strings.Contains(strings.ToLower(c.Names), q) ||
		strings.Contains(strings.ToLower(c.Image), q) ||
		strings.Contains(strings.ToLower(c.ID), q) ||
		strings.Contains(strings.ToLower(c.ComposeProject()), q) ||
		strings.Contains(strings.ToLower(c.ComposeService()), q)
}

func (m App) filtered() []docker.Container {
	if m.filterQuery != "" {
		q := strings.ToLower(m.filterQuery)
		out := make([]docker.Container, 0, len(m.sorted))
		for _, c := range m.sorted {
			if matchesFilter(c, q) {
				out = append(out, c)
			}
		}
		return out
	}

	if len(m.collapsedProjects) == 0 && len(m.expandedContainers) == 0 {
		return m.sorted
	}
	if len(m.collapsedProjects) == 0 {
		out := make([]docker.Container, 0, len(m.sorted)+len(m.expandedContainers)*4)
		for _, c := range m.sorted {
			out = append(out, c)
			if data, expanded := m.expandedContainers[c.ID]; expanded {
				out = append(out, detailRows(data)...)
			}
		}
		return out
	}

	out := make([]docker.Container, 0, len(m.sorted))
	var pendingProj string
	var pendingGroup []docker.Container

	flush := func() {
		if pendingProj != "" {
			out = append(out, collapseSummary(pendingProj, pendingGroup))
			pendingProj = ""
			pendingGroup = pendingGroup[:0]
		}
	}

	for _, c := range m.sorted {
		proj := c.ComposeProject()
		if proj != "" && m.collapsedProjects[proj] {
			if proj != pendingProj {
				flush()
				pendingProj = proj
			}
			pendingGroup = append(pendingGroup, c)
		} else {
			flush()
			out = append(out, c)
			if data, expanded := m.expandedContainers[c.ID]; expanded {
				out = append(out, detailRows(data)...)
			}
		}
	}
	flush()

	return out
}

func (m App) logsPanelHeight() int    { return max(5, m.height-tableChrome) }
func (m App) inspectPanelHeight() int { return max(5, m.height-tableChrome) }
func (m App) eventsPanelHeight() int  { return max(5, min(12, m.height/3)) }

func (m App) currentSelectedID() string {
	if c, ok := m.selectedContainer(); ok {
		return c.ID
	}
	return ""
}

func (m App) selectedContainer() (docker.Container, bool) {
	return m.selectedContainerFrom(m.filtered())
}

func (m App) selectedContainerFrom(filtered []docker.Container) (docker.Container, bool) {
	if c := m.table.Cursor(); c >= 0 && c < len(filtered) {
		return filtered[c], true
	}
	return docker.Container{}, false
}

func (m App) projectHasRunning(project string) bool {
	for _, c := range m.containers {
		if c.ComposeProject() == project && c.State == docker.StateRunning {
			return true
		}
	}
	return false
}

func (m App) containerByID(id string) (docker.Container, bool) {
	c, ok := m.containersByID[id]
	return c, ok
}

func indexContainers(cs []docker.Container) map[string]docker.Container {
	idx := make(map[string]docker.Container, len(cs))
	for _, c := range cs {
		idx[c.ID] = c
	}
	return idx
}

func (m App) ensureCursorVisible() App {
	cursor := m.table.Cursor()
	height := m.tableHeight()
	if cursor < m.viewportStart {
		m.viewportStart = cursor
	} else if height > 0 && cursor >= m.viewportStart+height {
		m.viewportStart = cursor - height + 1
	}
	return m
}

func (m App) rebuildTable(selectedID string) App {
	filtered := m.filtered()

	var statsForTable map[string]docker.StatsEntry
	if m.showInlineStats {
		statsForTable = m.inlineStats
	}
	m.table = buildTable(filtered, m.width-2, statsForTable)
	m.table.SetHeight(m.tableHeight())
	m.viewportStart = 0

	if selectedID != "" {
		if _, ok := m.containersByID[selectedID]; ok {
			if i := slices.IndexFunc(filtered, func(c docker.Container) bool { return c.ID == selectedID }); i >= 0 {
				m.table.SetCursor(i)
				lastRow := i
				for j := i + 1; j < len(filtered) && filtered[j].State == docker.StateDetail; j++ {
					lastRow = j
				}
				if h := m.tableHeight(); h > 0 && lastRow >= h {
					m.viewportStart = lastRow - h + 1
				}
			}
		}
	}
	return m
}

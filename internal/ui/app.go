package ui

import (
	"context"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

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
)

type App struct {
	client        docker.Client
	table         table.Model
	containers    []docker.Container
	sorted        []docker.Container
	viewportStart int
	showAll       bool
	loading       bool
	op            Operation
	confirmAction string
	confirmID     string
	confirmName   string
	filtering     bool
	filterQuery   string
	err           error
	width         int
	height        int

	logs    logsState
	inspect inspectState

	fetchStart time.Time
	fetchGen   int
	fetchSlow  bool

	copiedName      string
	warnMsg         string
	opGen           int
	opVisible       bool
	loadingVisible  bool
	version         string
	updateAvailable string

	ctxPicker ctxPickerState

	stats statsState

	containersByID map[string]docker.Container

	events         eventsState
	bgEventsGen    int
	pendingRefresh bool

	helpVisible bool
}

func New(version string) App {
	return newWithClient(docker.CLI{}, version)
}

func newWithClient(c docker.Client, version string) App {
	return App{
		client:      c,
		version:     version,
		loading:     true,
		fetchStart:  time.Now(),
		fetchGen:    1,
		showAll:     true,
		table:       buildTable(nil, 120),
		logs:        logsState{scroll: scrollState{autoScroll: true}},
		events:      eventsState{scroll: scrollState{autoScroll: true}},
		bgEventsGen: 1,
	}
}

func (m App) Init() tea.Cmd {
	return tea.Batch(
		m.client.FetchContainers(m.showAll),
		m.client.FetchContexts(),
		m.client.StartEvents(context.Background(), m.bgEventsGen),
		fetchTimerCmd(),
		fetchSlowCmd(m.fetchGen),
		checkUpdateCmd(m.version),
	)
}

func (m App) filtered() []docker.Container {
	if m.filterQuery == "" {
		return m.sorted
	}
	q := strings.ToLower(m.filterQuery)
	var out []docker.Container
	for _, c := range m.sorted {
		if strings.Contains(strings.ToLower(c.Names), q) ||
			strings.Contains(strings.ToLower(c.Image), q) ||
			strings.Contains(strings.ToLower(c.ID), q) ||
			strings.Contains(strings.ToLower(c.ComposeProject()), q) ||
			strings.Contains(strings.ToLower(c.ComposeService()), q) {
			out = append(out, c)
		}
	}
	return out
}

func (m App) logsPanelHeight() int    { return max(5, m.height-tableChrome) }
func (m App) inspectPanelHeight() int { return max(5, m.height-tableChrome) }
func (m App) eventsPanelHeight() int  { return max(5, min(12, m.height/3)) }

func (m App) currentSelectedID() string {
	filtered := m.filtered()
	if c := m.table.Cursor(); c >= 0 && c < len(filtered) {
		return filtered[c].ID
	}
	return ""
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

func (m App) rebuildTable(selectedID string) App {
	filtered := m.filtered()

	m.table = buildTable(filtered, m.width-2)
	m.table.SetHeight(m.tableHeight())
	m.viewportStart = 0

	if selectedID != "" {
		if _, ok := m.containersByID[selectedID]; ok {
			if i := slices.IndexFunc(filtered, func(c docker.Container) bool { return c.ID == selectedID }); i >= 0 {
				m.table.SetCursor(i)
				if h := m.tableHeight(); h > 0 && i >= h {
					m.viewportStart = i - h + 1
				}
			}
		}
	}
	return m
}

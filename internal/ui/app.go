package ui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pivovarit/tdocker/internal/docker"
)

const (
	logsPanelHeight    = 15
	inspectPanelHeight = 20
	statsPanelHeight   = 9
)

type Operation int

const (
	OpNone Operation = iota
	OpConfirming
	OpStopping
	OpStarting
	OpDeleting
)

type App struct {
	client             docker.Client
	table              table.Model
	containers         []docker.Container
	sorted             []docker.Container
	filteredContainers []docker.Container
	filteredQuery      string
	viewportStart      int
	showAll            bool
	loading            bool
	op                 Operation
	confirmAction      string
	confirmID          string
	confirmName        string
	filtering          bool
	filterQuery        string
	err                error
	width              int
	height             int
	logsVisible        bool
	logsLines          []string
	logsContainer      string
	logsContainerID    string
	logsScrollOffset   int
	logsAutoScroll     bool
	logsAllMode        bool
	logsGen            int
	logsCancel         context.CancelFunc

	inspectVisible   bool
	inspectLines     []string
	inspectContainer string
	inspectOffset    int

	copiedName string

	statsVisible     bool
	statsContainer   string
	statsContainerID string
	statsEntry       *docker.StatsEntry
	statsFetching    bool
}

func New() App {
	return newWithClient(docker.CLI{})
}

func newWithClient(c docker.Client) App {
	return App{
		client:  c,
		loading: true,
		showAll: true,
		table:   buildTable(nil, 120),
	}
}

func (m App) Init() tea.Cmd {
	return m.client.FetchContainers(m.showAll)
}

func (m App) filtered() []docker.Container {
	if m.filteredQuery == m.filterQuery && m.filteredContainers != nil {
		return m.filteredContainers
	}
	return m.computeFilter().filteredContainers
}

func (m App) computeFilter() App {
	m.filteredQuery = m.filterQuery
	if m.filterQuery == "" {
		m.filteredContainers = m.sorted
		return m
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
	m.filteredContainers = out
	return m
}

func (m App) rebuildTable() App {
	var selectedID string
	if prev := m.filteredContainers; len(prev) > 0 {
		if c := m.table.Cursor(); c >= 0 && c < len(prev) {
			selectedID = prev[c].ID
		}
	}

	m = m.computeFilter()
	m.table = buildTable(m.filteredContainers, m.width)
	m.table.SetHeight(m.tableHeight())
	m.viewportStart = 0

	if selectedID != "" {
		for i, c := range m.filteredContainers {
			if c.ID == selectedID {
				m.table.SetCursor(i)
				if h := m.tableHeight(); h > 0 && i >= h {
					m.viewportStart = i - h + 1
				}
				break
			}
		}
	}
	return m
}

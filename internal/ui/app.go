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
	OpDeleting
)

type App struct {
	client           docker.Client
	table            table.Model
	containers       []docker.Container
	sorted           []docker.Container
	viewportStart    int
	showAll          bool
	loading          bool
	op               Operation
	confirmAction    string
	confirmID        string
	confirmName      string
	filtering        bool
	filterQuery      string
	err              error
	width            int
	height           int
	logsVisible      bool
	logsLines        []string
	logsContainer    string
	logsContainerID  string
	logsScrollOffset int
	logsAutoScroll   bool
	logsAllMode      bool
	logsGen          int
	logsCancel       context.CancelFunc

	inspectVisible   bool
	inspectLines     []string
	inspectContainer string
	inspectOffset    int

	copiedName string

	contextPickerVisible   bool
	contextPickerRequested bool
	contexts               []docker.DockerContext
	contextCursor          int
	currentContext         string

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
	return tea.Batch(m.client.FetchContainers(m.showAll), m.client.FetchContexts())
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

func (m App) currentSelectedID() string {
	filtered := m.filtered()
	if c := m.table.Cursor(); c >= 0 && c < len(filtered) {
		return filtered[c].ID
	}
	return ""
}

func (m App) rebuildTable(selectedID string) App {
	filtered := m.filtered()

	m.table = buildTable(filtered, m.width)
	m.table.SetHeight(m.tableHeight())
	m.viewportStart = 0

	if selectedID != "" {
		for i, c := range filtered {
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

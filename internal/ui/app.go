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
	statsRows          = 5
	statsPanelHeight   = statsRows + 3
	eventsPanelHeight  = 12
	logsTailDefault    = "200"

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
)

type scrollState struct {
	offset     int
	autoScroll bool
}

func (s scrollState) up() scrollState {
	if s.offset > 0 {
		s.offset--
		s.autoScroll = false
	}
	return s
}

func (s scrollState) down(contentLen, viewHeight int) scrollState {
	maxOff := max(0, contentLen-viewHeight)
	if s.offset < maxOff {
		s.offset++
	}
	if s.offset >= maxOff {
		s.autoScroll = true
	}
	return s
}

func (s scrollState) top() scrollState {
	s.offset = 0
	s.autoScroll = false
	return s
}

func (s scrollState) bottom(contentLen, viewHeight int) scrollState {
	s.offset = max(0, contentLen-viewHeight)
	s.autoScroll = true
	return s
}

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

	copiedName string

	ctxPicker ctxPickerState

	stats statsState

	containersByID map[string]docker.Container

	events         eventsState
	bgEventsGen    int
	pendingRefresh bool
}

func New() App {
	return newWithClient(docker.CLI{})
}

func newWithClient(c docker.Client) App {
	return App{
		client:      c,
		loading:     true,
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

	m.table = buildTable(filtered, m.width)
	m.table.SetHeight(m.tableHeight())
	m.viewportStart = 0

	if selectedID != "" {
		if _, ok := m.containersByID[selectedID]; ok {
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
	}
	return m
}

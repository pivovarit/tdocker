package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pivovarit/tdocker/internal/docker"
)

type eventsState struct {
	visible bool
	events  []docker.Event
	scroll  scrollState
}

func (m App) closeEvents() App {
	m.events = eventsState{scroll: scrollState{autoScroll: true}}
	m.table.SetHeight(m.tableHeight())
	return m
}

func (m App) handleEventsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "v":
		m = m.closeEvents()
	case "up", "k":
		m.events.scroll = m.events.scroll.up()
	case "down", "j":
		m.events.scroll = m.events.scroll.down(len(m.events.events), eventsPanelHeight-2)
	case "g", "home":
		m.events.scroll = m.events.scroll.top()
	case "G", "end":
		m.events.scroll = m.events.scroll.bottom(len(m.events.events), eventsPanelHeight-2)
	}
	return m, nil
}

func (m App) renderEventsPanel() string {
	return m.renderPanel(" Events  (live)", func(b *strings.Builder) {
		maxLines := eventsPanelHeight - 2
		if len(m.events.events) == 0 {
			b.WriteString(emptyStyle.Render("Waiting for events…"))
			b.WriteString("\n")
			panelPad(b, 1, maxLines)
			return
		}
		start := m.events.scroll.offset
		end := start + maxLines
		if end > len(m.events.events) {
			end = len(m.events.events)
		}
		shown := 0
		for i := start; i < end; i++ {
			ev := m.events.events[i]
			actionStyle := eventDimStyle
			switch ev.Action {
			case "start", "unpause", "create":
				actionStyle = eventStartStyle
			case "die", "stop", "kill", "destroy", "oom":
				actionStyle = eventStopStyle
			case "pause":
				actionStyle = eventWarnStyle
			}
			line := eventTimeStyle.Render(ev.Timestamp()) + "  " +
				eventTypeStyle.Render(fmt.Sprintf("%-11s", ev.Type)) +
				actionStyle.Render(fmt.Sprintf("%-12s", ev.Action)) +
				eventNameStyle.Render(ev.Name())
			b.WriteString("  " + line + "\n")
			shown++
		}
		panelPad(b, shown, maxLines)
	})
}

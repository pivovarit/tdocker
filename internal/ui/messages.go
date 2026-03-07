package ui

type statsTickMsg struct{}
type autoRefreshMsg struct{}
type bgEventsRestartMsg struct{ gen int }
type fetchTimerTickMsg struct{}
type fetchSlowMsg struct{ gen int }
type opSlowMsg struct{ gen int }
type opDisplayMsg struct{ gen int }
type updateAvailableMsg struct{ version string }

type clipboardMsg struct {
	name string
	err  error
}

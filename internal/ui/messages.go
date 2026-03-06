package ui

type statsTickMsg struct{}
type autoRefreshMsg struct{}
type bgEventsRestartMsg struct{ gen int }
type fetchTimerTickMsg struct{}
type fetchSlowMsg struct{ gen int }

type clipboardMsg struct {
	name string
	err  error
}

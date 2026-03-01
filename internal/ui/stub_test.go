package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

type stubClient struct {
	fetchContainers  func(bool) tea.Cmd
	stopContainer    func(string) tea.Cmd
	startContainer   func(string) tea.Cmd
	deleteContainer  func(string) tea.Cmd
	execContainer    func(string) tea.Cmd
	checkDebugAvail  func(string) tea.Cmd
	debugContainer   func(string) tea.Cmd
	inspectContainer func(string) tea.Cmd
	fetchStats       func(string) tea.Cmd
	startLogs        func(context.Context, string, string, int) tea.Cmd
}

func newStubClient() *stubClient {
	noop := func() tea.Msg { return nil }
	noopStr := func(_ string) tea.Cmd { return noop }
	return &stubClient{
		fetchContainers:  func(_ bool) tea.Cmd { return noop },
		stopContainer:    noopStr,
		startContainer:   noopStr,
		deleteContainer:  noopStr,
		execContainer:    noopStr,
		checkDebugAvail:  noopStr,
		debugContainer:   noopStr,
		inspectContainer: noopStr,
		fetchStats:       noopStr,
		startLogs:        func(_ context.Context, _ string, _ string, _ int) tea.Cmd { return noop },
	}
}

func (c *stubClient) FetchContainers(all bool) tea.Cmd      { return c.fetchContainers(all) }
func (c *stubClient) StopContainer(id string) tea.Cmd       { return c.stopContainer(id) }
func (c *stubClient) StartContainer(id string) tea.Cmd      { return c.startContainer(id) }
func (c *stubClient) DeleteContainer(id string) tea.Cmd     { return c.deleteContainer(id) }
func (c *stubClient) ExecContainer(id string) tea.Cmd       { return c.execContainer(id) }
func (c *stubClient) CheckDebugAvailable(id string) tea.Cmd { return c.checkDebugAvail(id) }
func (c *stubClient) DebugContainer(id string) tea.Cmd      { return c.debugContainer(id) }
func (c *stubClient) InspectContainer(id string) tea.Cmd    { return c.inspectContainer(id) }
func (c *stubClient) FetchStats(id string) tea.Cmd          { return c.fetchStats(id) }
func (c *stubClient) StartLogs(ctx context.Context, id string, tail string, gen int) tea.Cmd {
	return c.startLogs(ctx, id, tail, gen)
}

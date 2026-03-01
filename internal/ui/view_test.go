package ui

import (
	"strings"
	"testing"

	"github.com/pivovarit/tdocker/internal/docker"
)

func viewApp() App {
	m := modelWithSorted(nil)
	m.width = 120
	return m
}

func statsEntry(cpu, mem, net, blk, pids string) docker.StatsEntry {
	return docker.StatsEntry{
		CPUPerc:  cpu,
		MemUsage: mem + " / 1GiB",
		MemPerc:  mem,
		NetIO:    net + " / 0B",
		BlockIO:  blk + " / 0B",
		PIDs:     pids,
	}
}

func TestRenderStatsPanel_LoadingShowsMessage(t *testing.T) {
	m := statsPanel()
	out := m.renderStatsPanel()
	if !strings.Contains(out, "Loading") {
		t.Errorf("want 'Loading' when entry is nil, got:\n%s", out)
	}
}

func TestRenderStatsPanel_TitleContainsContainerName(t *testing.T) {
	m := statsPanel()
	m.stats.entry = &docker.StatsEntry{CPUPerc: "1.00%"}
	out := m.renderStatsPanel()
	if !strings.Contains(out, runningContainer.Names) {
		t.Errorf("want container name %q in title, got:\n%s", runningContainer.Names, out)
	}
}

func TestRenderStatsPanel_AllMetricLabelsPresent(t *testing.T) {
	m := statsPanel()
	e := statsEntry("1.00%", "1.00%", "1kB", "0B", "3")
	m.stats.entry = &e
	out := m.renderStatsPanel()
	for _, label := range []string{"CPU", "Memory", "Net I/O", "Block I/O", "PIDs"} {
		if !strings.Contains(out, label) {
			t.Errorf("want label %q in stats panel output, got:\n%s", label, out)
		}
	}
}

func TestRenderStatsPanel_MetricValuesPresent(t *testing.T) {
	m := statsPanel()
	e := docker.StatsEntry{
		CPUPerc:  "3.14%",
		MemUsage: "42MiB / 1GiB",
		MemPerc:  "4.10%",
		NetIO:    "1.5kB / 900B",
		BlockIO:  "8MB / 0B",
		PIDs:     "7",
	}
	m.stats.entry = &e
	out := m.renderStatsPanel()
	for _, want := range []string{"3.14%", "42MiB", "4.10%", "7"} {
		if !strings.Contains(out, want) {
			t.Errorf("want %q in stats panel output, got:\n%s", want, out)
		}
	}
}

func TestRenderStatsPanel_TrendUpWhenMetricsRise(t *testing.T) {
	m := statsPanel()
	prev := statsEntry("1.00%", "1.00%", "1kB", "1kB", "1")
	curr := statsEntry("90.00%", "90.00%", "900MB", "900MB", "50")
	m.stats.prevEntry = &prev
	m.stats.entry = &curr
	out := m.renderStatsPanel()
	if !strings.Contains(out, "↑") {
		t.Errorf("want ↑ trend when all metrics rise significantly, got:\n%s", out)
	}
}

func TestRenderStatsPanel_TrendDownWhenMetricsDrop(t *testing.T) {
	m := statsPanel()
	prev := statsEntry("90.00%", "90.00%", "900MB", "900MB", "50")
	curr := statsEntry("1.00%", "1.00%", "1kB", "1kB", "1")
	m.stats.prevEntry = &prev
	m.stats.entry = &curr
	out := m.renderStatsPanel()
	if !strings.Contains(out, "↓") {
		t.Errorf("want ↓ trend when all metrics drop significantly, got:\n%s", out)
	}
}

func TestRenderStatsPanel_TrendSteadyWhenUnchanged(t *testing.T) {
	m := statsPanel()
	same := statsEntry("5.00%", "5.00%", "1kB", "1kB", "4")
	m.stats.prevEntry = &same
	m.stats.entry = &same
	out := m.renderStatsPanel()
	if !strings.Contains(out, "·") {
		t.Errorf("want · steady indicator when metrics unchanged, got:\n%s", out)
	}
}

func TestRenderStatsPanel_NoPrevEntryNoTrend(t *testing.T) {
	m := statsPanel()
	e := statsEntry("5.00%", "5.00%", "1kB", "1kB", "4")
	m.stats.entry = &e
	out := m.renderStatsPanel()
	if strings.Contains(out, "↑") || strings.Contains(out, "↓") {
		t.Errorf("want no trend arrows without prevEntry, got:\n%s", out)
	}
}

func TestRenderLogsPanel_TitleContainsContainerName(t *testing.T) {
	m := logsOpenWithLines(nil)
	out := m.renderLogsPanel()
	if !strings.Contains(out, runningContainer.Names) {
		t.Errorf("want container name %q in logs title, got:\n%s", runningContainer.Names, out)
	}
}

func TestRenderLogsPanel_DefaultModeLabelLast200(t *testing.T) {
	m := logsOpenWithLines(nil)
	out := m.renderLogsPanel()
	if !strings.Contains(out, "200") {
		t.Errorf("want 'last 200' label in default mode, got:\n%s", out)
	}
}

func TestRenderLogsPanel_AllModeLabelAll(t *testing.T) {
	m := logsOpenWithLines(nil)
	m.logs.allMode = true
	out := m.renderLogsPanel()
	if !strings.Contains(out, "all") {
		t.Errorf("want 'all' label in all-mode, got:\n%s", out)
	}
}

func TestRenderLogsPanel_LinesAppearInOutput(t *testing.T) {
	lines := []string{"first log line", "second log line", "third log line"}
	m := logsOpenWithLines(lines)
	out := m.renderLogsPanel()
	for _, line := range lines {
		if !strings.Contains(out, line) {
			t.Errorf("want log line %q in output, got:\n%s", line, out)
		}
	}
}

func TestRenderInspectPanel_LoadingShowsMessage(t *testing.T) {
	m := viewApp()
	m.inspect.visible = true
	m.inspect.container = runningContainer.Names
	out := m.renderInspectPanel()
	if !strings.Contains(out, "Loading") {
		t.Errorf("want 'Loading' when inspect lines empty, got:\n%s", out)
	}
}

func TestRenderInspectPanel_TitleContainsContainerName(t *testing.T) {
	m := viewApp()
	m.inspect.container = "my-service"
	out := m.renderInspectPanel()
	if !strings.Contains(out, "my-service") {
		t.Errorf("want container name in inspect title, got:\n%s", out)
	}
}

func TestRenderEventsPanel_EmptyShowsWaiting(t *testing.T) {
	m := viewApp()
	m.events.visible = true
	out := m.renderEventsPanel()
	if !strings.Contains(out, "Waiting") {
		t.Errorf("want 'Waiting' when no events yet, got:\n%s", out)
	}
}

func TestRenderEventsPanel_EventTypeAppearsInOutput(t *testing.T) {
	m := viewApp()
	m.events.visible = true
	m.events.events = []docker.Event{
		{Type: "container", Action: "start", Actor: docker.EventActor{Attributes: map[string]string{"name": "my-app"}}, Time: 1000},
	}
	out := m.renderEventsPanel()
	if !strings.Contains(out, "container") {
		t.Errorf("want event type 'container' in output, got:\n%s", out)
	}
}

func TestRenderEventsPanel_EventActionAppearsInOutput(t *testing.T) {
	m := viewApp()
	m.events.visible = true
	m.events.events = []docker.Event{
		{Type: "container", Action: "die", Actor: docker.EventActor{Attributes: map[string]string{"name": "my-app"}}, Time: 1000},
	}
	out := m.renderEventsPanel()
	if !strings.Contains(out, "die") {
		t.Errorf("want event action 'die' in output, got:\n%s", out)
	}
}

func TestRenderEventsPanel_ContainerNameAppearsInOutput(t *testing.T) {
	m := viewApp()
	m.events.visible = true
	m.events.events = []docker.Event{
		{Type: "container", Action: "start", Actor: docker.EventActor{Attributes: map[string]string{"name": "special-name"}}, Time: 1000},
	}
	out := m.renderEventsPanel()
	if !strings.Contains(out, "special-name") {
		t.Errorf("want container name 'special-name' in output, got:\n%s", out)
	}
}

func TestRenderEventsPanel_MultipleEventsAllAppear(t *testing.T) {
	m := viewApp()
	m.events.visible = true
	m.events.events = []docker.Event{
		{Type: "container", Action: "start", Actor: docker.EventActor{Attributes: map[string]string{"name": "alpha"}}, Time: 1000},
		{Type: "network", Action: "connect", Actor: docker.EventActor{Attributes: map[string]string{"name": "bridge"}}, Time: 1001},
		{Type: "container", Action: "die", Actor: docker.EventActor{Attributes: map[string]string{"name": "beta"}}, Time: 1002},
	}
	out := m.renderEventsPanel()
	for _, want := range []string{"alpha", "beta", "network", "die"} {
		if !strings.Contains(out, want) {
			t.Errorf("want %q in events output, got:\n%s", want, out)
		}
	}
}

func TestStatsTrend_UpWhenValueRises(t *testing.T) {
	got := statsTrend("1.00%", "50.00%", parsePercent)
	if !strings.Contains(got, "↑") {
		t.Errorf("want ↑ for 1%%→50%%, got %q", got)
	}
}

func TestStatsTrend_DownWhenValueDrops(t *testing.T) {
	got := statsTrend("50.00%", "1.00%", parsePercent)
	if !strings.Contains(got, "↓") {
		t.Errorf("want ↓ for 50%%→1%%, got %q", got)
	}
}

func TestStatsTrend_SteadyWhenUnchanged(t *testing.T) {
	got := statsTrend("5.00%", "5.00%", parsePercent)
	if !strings.Contains(got, "·") {
		t.Errorf("want · for identical values, got %q", got)
	}
}

func TestStatsTrend_EmptyOnUnparseable(t *testing.T) {
	got := statsTrend("n/a", "n/a", parsePercent)
	if got != "" {
		t.Errorf("want empty string for unparseable values, got %q", got)
	}
}

func TestParsePercent_Valid(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"0%", 0},
		{"0.00%", 0},
		{"1.00%", 1.0},
		{"0.42%", 0.42},
		{"100%", 100.0},
		{"100.00%", 100.0},
		{" 3.14% ", 3.14},
	}
	for _, tc := range cases {
		v, ok := parsePercent(tc.in)
		if !ok {
			t.Errorf("parsePercent(%q): want ok=true", tc.in)
		}
		if abs(v-tc.want) > 0.001 {
			t.Errorf("parsePercent(%q): want %v, got %v", tc.in, tc.want, v)
		}
	}
}

func TestParsePercent_InvalidReturnsNotOk(t *testing.T) {
	for _, s := range []string{"n/a", "", "abc%", "--", "-- / --"} {
		if _, ok := parsePercent(s); ok {
			t.Errorf("parsePercent(%q): want ok=false", s)
		}
	}
}

func TestParseByteSize_Units(t *testing.T) {
	const TiB = 1024 * 1024 * 1024 * 1024
	cases := []struct {
		in   string
		want float64
	}{
		// SI
		{"1B", 1},
		{"0B", 0},
		{"1kB", 1000},
		{"1MB", 1e6},
		{"1GB", 1e9},
		{"1TB", 1e12},
		// IEC
		{"1KiB", 1024},
		{"1MiB", 1024 * 1024},
		{"1GiB", 1024 * 1024 * 1024},
		{"1TiB", TiB},
		// fractional
		{"1.5MiB", 1.5 * 1024 * 1024},
		{"2.5kB", 2500},
	}
	for _, tc := range cases {
		v, ok := parseByteSize(tc.in)
		if !ok {
			t.Errorf("parseByteSize(%q): want ok=true", tc.in)
		}
		if abs(v-tc.want) > 0.001 {
			t.Errorf("parseByteSize(%q): want %v, got %v", tc.in, tc.want, v)
		}
	}
}

func TestParseByteSize_InvalidReturnsNotOk(t *testing.T) {
	for _, s := range []string{"", "abc", "n/a", "--", "-- / --"} {
		if _, ok := parseByteSize(s); ok {
			t.Errorf("parseByteSize(%q): want ok=false", s)
		}
	}
}

func TestParseSizeFirst_Valid(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"500kB / 1MB", 500_000},
		{"0B / 0B", 0},
		{"1.2GiB / 800MB", 1.2 * 1024 * 1024 * 1024},
		{"2MiB", float64(2 * 1024 * 1024)}, // no slash
	}
	for _, tc := range cases {
		v, ok := parseSizeFirst(tc.in)
		if !ok {
			t.Fatalf("parseSizeFirst(%q): want ok=true", tc.in)
		}
		if abs(v-tc.want) > 0.001 {
			t.Errorf("parseSizeFirst(%q): want %v, got %v", tc.in, tc.want, v)
		}
	}
}

func TestParseSizeFirst_StoppedContainerDashDash(t *testing.T) {
	// docker stats outputs "-- / --" for stopped containers
	if _, ok := parseSizeFirst("-- / --"); ok {
		t.Error("parseSizeFirst(\"-- / --\"): want ok=false for stopped-container placeholder")
	}
}

func TestParseNumber_Valid(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"0", 0},
		{"1", 1},
		{"42", 42},
		{" 7 ", 7},
	}
	for _, tc := range cases {
		v, ok := parseNumber(tc.in)
		if !ok {
			t.Errorf("parseNumber(%q): want ok=true", tc.in)
		}
		if v != tc.want {
			t.Errorf("parseNumber(%q): want %v, got %v", tc.in, tc.want, v)
		}
	}
}

func TestParseNumber_InvalidReturnsNotOk(t *testing.T) {
	for _, s := range []string{"abc", "", "--", "n/a"} {
		if _, ok := parseNumber(s); ok {
			t.Errorf("parseNumber(%q): want ok=false", s)
		}
	}
}

func TestView_LoadingState(t *testing.T) {
	m := viewApp()
	m.loading = true
	out := m.View()
	if !strings.Contains(out, "Fetching") {
		t.Errorf("want 'Fetching' in loading view, got:\n%s", out)
	}
}

func TestView_ErrorState(t *testing.T) {
	m := viewApp()
	m.loading = false
	m.err = errSentinel("something went wrong")
	out := m.View()
	if !strings.Contains(out, "something went wrong") {
		t.Errorf("want error message in view, got:\n%s", out)
	}
}

func TestView_EmptyRunningContainers(t *testing.T) {
	m := viewApp()
	m.loading = false
	m.showAll = false
	out := m.View()
	if !strings.Contains(out, "No running") {
		t.Errorf("want 'No running containers' message, got:\n%s", out)
	}
}

func TestView_EmptyAllContainers(t *testing.T) {
	m := viewApp()
	m.loading = false
	m.showAll = true
	out := m.View()
	if !strings.Contains(out, "No containers") {
		t.Errorf("want 'No containers found' message, got:\n%s", out)
	}
}

func TestView_HelpBarContainsKeyBindings(t *testing.T) {
	m := modelWithSorted([]docker.Container{runningContainer})
	out := m.View()
	for _, key := range []string{"l", "i", "t", "v", "s", "S", "r"} {
		if !strings.Contains(out, key) {
			t.Errorf("want key %q in help bar, got:\n%s", key, out)
		}
	}
}

type errSentinel string

func (e errSentinel) Error() string { return string(e) }

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

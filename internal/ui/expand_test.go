package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/pivovarit/tdocker/internal/docker"
)

func TestDetailRows_NilData_ReturnsLoadingRow(t *testing.T) {
	rows := detailRows(nil)
	if len(rows) != 1 {
		t.Fatalf("want 1 loading row, got %d", len(rows))
	}
	if rows[0].State != "detail" {
		t.Errorf("want state=detail, got %q", rows[0].State)
	}
	if !strings.Contains(rows[0].Names, "loading") {
		t.Errorf("want loading indicator in Names, got %q", rows[0].Names)
	}
	if !strings.HasPrefix(rows[0].Names, "└") {
		t.Errorf("want └ prefix on single loading row, got %q", rows[0].Names)
	}
}

func TestDetailRows_LastRowUsesCornerPrefix(t *testing.T) {
	data := &docker.InspectData{
		Ports: map[string][]docker.PortBinding{
			"80/tcp": {{HostIP: "0.0.0.0", HostPort: "80"}},
		},
		Networks: []docker.NetworkInfo{
			{Name: "bridge", IPAddress: "172.17.0.2"},
		},
	}
	rows := detailRows(data)
	if len(rows) == 0 {
		t.Fatal("want at least one row")
	}
	last := rows[len(rows)-1]
	if !strings.HasPrefix(last.Names, "└") {
		t.Errorf("want last row to start with └, got %q", last.Names)
	}
	for _, r := range rows[:len(rows)-1] {
		if !strings.HasPrefix(r.Names, "│") {
			t.Errorf("want non-last rows to start with │, got %q", r.Names)
		}
	}
}

func TestDetailRows_AllStateIsDetail(t *testing.T) {
	data := &docker.InspectData{
		Networks: []docker.NetworkInfo{{Name: "bridge", IPAddress: "172.17.0.2"}},
	}
	for _, r := range detailRows(data) {
		if r.State != "detail" {
			t.Errorf("want state=detail, got %q", r.State)
		}
		if r.ID != "" || r.Image != "" || r.Command != "" {
			t.Errorf("want all other fields empty, got ID=%q Image=%q Command=%q", r.ID, r.Image, r.Command)
		}
	}
}

func TestDetailRows_PortsFormatted(t *testing.T) {
	data := &docker.InspectData{
		Ports: map[string][]docker.PortBinding{
			"80/tcp": {{HostIP: "0.0.0.0", HostPort: "8080"}},
		},
	}
	rows := detailRows(data)
	found := false
	for _, r := range rows {
		if strings.Contains(r.Names, "80/tcp") && strings.Contains(r.Names, "8080") {
			found = true
		}
	}
	if !found {
		t.Errorf("want port mapping in Names, rows: %v", rows)
	}
}

func TestDetailRows_NetworksFormatted(t *testing.T) {
	data := &docker.InspectData{
		Networks: []docker.NetworkInfo{
			{Name: "bridge", IPAddress: "172.17.0.2"},
		},
	}
	rows := detailRows(data)
	found := false
	for _, r := range rows {
		if strings.Contains(r.Names, "bridge") && strings.Contains(r.Names, "172.17.0.2") {
			found = true
		}
	}
	if !found {
		t.Errorf("want network in Names, rows: %v", rows)
	}
}

func TestDetailRows_EmptyData_ReturnsNoRows(t *testing.T) {
	data := &docker.InspectData{}
	rows := detailRows(data)
	if len(rows) != 0 {
		t.Errorf("want 0 rows for empty data, got %d", len(rows))
	}
}

func TestFiltered_ExpandedContainerInjectsDetailRows(t *testing.T) {
	containers := []docker.Container{
		{ID: "s1", Names: "solo", State: "running"},
	}
	m := modelWithSorted(containers)
	m.expandedContainers["s1"] = &docker.InspectData{
		Networks: []docker.NetworkInfo{{Name: "bridge", IPAddress: "172.17.0.2"}},
	}

	got := m.filtered()
	if len(got) != 2 {
		t.Fatalf("want 2 rows (container + detail), got %d", len(got))
	}
	if got[0].ID != "s1" {
		t.Errorf("first row should be the container, got %q", got[0].ID)
	}
	if got[1].State != "detail" {
		t.Errorf("second row should be detail, got state=%q", got[1].State)
	}
}

func TestFiltered_ExpandedContainerNilData_ShowsLoadingRow(t *testing.T) {
	containers := []docker.Container{
		{ID: "s1", Names: "solo", State: "running"},
	}
	m := modelWithSorted(containers)
	m.expandedContainers["s1"] = nil

	got := m.filtered()
	if len(got) != 2 {
		t.Fatalf("want 2 rows, got %d", len(got))
	}
	if got[1].State != "detail" {
		t.Errorf("second row should be detail, got %q", got[1].State)
	}
	if !strings.Contains(got[1].Names, "loading") {
		t.Errorf("want loading indicator, got %q", got[1].Names)
	}
}

func TestFiltered_ExpandedContainerNotShownDuringFilter(t *testing.T) {
	containers := []docker.Container{
		{ID: "s1", Names: "solo", State: "running"},
	}
	m := modelWithSorted(containers)
	m.expandedContainers["s1"] = &docker.InspectData{
		Networks: []docker.NetworkInfo{{Name: "bridge", IPAddress: "172.17.0.2"}},
	}
	m.filterQuery = "solo"

	got := m.filtered()
	if len(got) != 1 {
		t.Fatalf("want 1 row (no detail rows during filter), got %d", len(got))
	}
}

func TestRightArrow_ExpandsStandaloneContainer(t *testing.T) {
	mc := newStubClient()
	var gotID string
	mc.inspectContainerExpand = func(id string) tea.Cmd {
		gotID = id
		return func() tea.Msg { return nil }
	}
	containers := []docker.Container{
		{ID: "s1", Names: "solo", State: "running"},
	}
	m := modelWithMock(mc, containers)
	got := update(m, tea.KeyPressMsg{Code: tea.KeyRight})

	if _, ok := got.expandedContainers["s1"]; !ok {
		t.Fatal("want s1 in expandedContainers after right arrow")
	}
	if gotID != "s1" {
		t.Errorf("want InspectContainerExpand(%q), got %q", "s1", gotID)
	}
}

func TestRightArrow_NoopOnAlreadyExpandedContainer(t *testing.T) {
	mc := newStubClient()
	callCount := 0
	mc.inspectContainerExpand = func(id string) tea.Cmd {
		callCount++
		return func() tea.Msg { return nil }
	}
	containers := []docker.Container{
		{ID: "s1", Names: "solo", State: "running"},
	}
	m := modelWithMock(mc, containers)
	m.expandedContainers["s1"] = nil
	m = m.rebuildTable("")
	update(m, tea.KeyPressMsg{Code: tea.KeyRight})
	if callCount != 0 {
		t.Errorf("want no InspectContainerExpand call on already-expanded container, got %d calls", callCount)
	}
}

func TestLeftArrow_CollapsesExpandedStandaloneContainer(t *testing.T) {
	containers := []docker.Container{
		{ID: "s1", Names: "solo", State: "running"},
	}
	m := modelWithSorted(containers)
	m.expandedContainers["s1"] = &docker.InspectData{
		Networks: []docker.NetworkInfo{{Name: "bridge", IPAddress: "172.17.0.2"}},
	}
	m = m.rebuildTable("")
	got := update(m, tea.KeyPressMsg{Code: tea.KeyLeft})
	if _, ok := got.expandedContainers["s1"]; ok {
		t.Fatal("want s1 removed from expandedContainers after left arrow")
	}
	if len(got.filtered()) != 1 {
		t.Errorf("want 1 row after collapse, got %d", len(got.filtered()))
	}
}

func TestLeftArrow_NoopOnNonExpandedStandaloneContainer(t *testing.T) {
	containers := []docker.Container{
		{ID: "s1", Names: "solo", State: "running"},
	}
	m := modelWithSorted(containers)
	got := update(m, tea.KeyPressMsg{Code: tea.KeyLeft})
	if len(got.expandedContainers) != 0 {
		t.Errorf("want expandedContainers unchanged, got %v", got.expandedContainers)
	}
}

func TestExpandInspectMsg_UpdatesExpandedContainers(t *testing.T) {
	containers := []docker.Container{
		{ID: "s1", Names: "solo", State: "running"},
	}
	m := modelWithSorted(containers)
	m.expandedContainers["s1"] = nil

	data := &docker.InspectData{
		Networks: []docker.NetworkInfo{{Name: "bridge", IPAddress: "172.17.0.2"}},
	}
	got := update(m, docker.ExpandInspectMsg{ContainerID: "s1", Data: data})

	if got.expandedContainers["s1"] == nil {
		t.Fatal("want expandedContainers[s1] populated after ExpandInspectMsg")
	}
	filtered := got.filtered()
	if len(filtered) != 2 {
		t.Fatalf("want 2 rows after data arrives, got %d", len(filtered))
	}
	if filtered[1].State != "detail" {
		t.Errorf("want detail row, got state=%q", filtered[1].State)
	}
}

func TestExpandInspectMsg_ErrorClearsExpansion(t *testing.T) {
	containers := []docker.Container{
		{ID: "s1", Names: "solo", State: "running"},
	}
	m := modelWithSorted(containers)
	m.expandedContainers["s1"] = nil

	got := update(m, docker.ExpandInspectMsg{ContainerID: "s1", Err: fmt.Errorf("inspect failed")})
	if _, ok := got.expandedContainers["s1"]; ok {
		t.Fatal("want s1 removed from expandedContainers on error")
	}
}

func TestNavigation_DownArrow_SkipsDetailRows(t *testing.T) {
	containers := []docker.Container{
		{ID: "s1", Names: "solo", State: "running"},
		{ID: "s2", Names: "other", State: "running"},
	}
	m := modelWithSorted(containers)
	m.expandedContainers["s1"] = &docker.InspectData{
		Networks: []docker.NetworkInfo{{Name: "bridge", IPAddress: "172.17.0.2"}},
	}
	m = m.rebuildTable("")
	m.table.SetCursor(0)

	got := update(m, tea.KeyPressMsg{Code: tea.KeyDown})
	cursor := got.table.Cursor()
	filtered := got.filtered()
	if filtered[cursor].State == "detail" {
		t.Errorf("cursor should not rest on detail row, filtered[%d]=%q", cursor, filtered[cursor].Names)
	}
	if filtered[cursor].ID != "s2" {
		t.Errorf("want cursor on s2, got ID=%q", filtered[cursor].ID)
	}
}

func TestNavigation_UpArrow_SkipsDetailRows(t *testing.T) {
	containers := []docker.Container{
		{ID: "s1", Names: "solo", State: "running"},
		{ID: "s2", Names: "other", State: "running"},
	}
	m := modelWithSorted(containers)
	m.expandedContainers["s1"] = &docker.InspectData{
		Networks: []docker.NetworkInfo{{Name: "bridge", IPAddress: "172.17.0.2"}},
	}
	m = m.rebuildTable("")
	m.table.SetCursor(2)

	got := update(m, tea.KeyPressMsg{Code: tea.KeyUp})
	cursor := got.table.Cursor()
	filtered := got.filtered()
	if filtered[cursor].State == "detail" {
		t.Errorf("cursor should not rest on detail row, filtered[%d]=%q", cursor, filtered[cursor].Names)
	}
	if filtered[cursor].ID != "s1" {
		t.Errorf("want cursor on s1, got ID=%q", filtered[cursor].ID)
	}
}

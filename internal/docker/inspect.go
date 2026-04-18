package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

const timeoutInspect = 10 * time.Second

type PortBinding struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

type Mount struct {
	Type        string `json:"Type"`
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	RW          bool   `json:"RW"`
}

type NetworkInfo struct {
	Name      string
	IPAddress string
}

type ContainerState struct {
	Status     string
	ExitCode   int
	Error      string
	OOMKilled  bool
	StartedAt  time.Time
	FinishedAt time.Time
	Health     *Health
}

type Health struct {
	Status        string
	FailingStreak int
	Log           []HealthLogEntry
}

type HealthLogEntry struct {
	Start    time.Time
	End      time.Time
	ExitCode int
	Output   string
}

type RestartPolicy struct {
	Name              string
	MaximumRetryCount int
}

type Healthcheck struct {
	Test        []string
	Interval    time.Duration
	Timeout     time.Duration
	Retries     int
	StartPeriod time.Duration
}

type InspectData struct {
	ImageDigest   string
	Ports         map[string][]PortBinding
	Env           []string
	Mounts        []Mount
	Networks      []NetworkInfo
	State         ContainerState
	RestartCount  int
	RestartPolicy RestartPolicy
	Healthcheck   *Healthcheck
}

type InspectMsg struct {
	Data *InspectData
	Err  error
}

type InspectLineKind int

const (
	InspectLineSection InspectLineKind = iota
	InspectLineKeyValue
	InspectLineValue
	InspectLineBlank
)

type InspectLine struct {
	Kind  InspectLineKind
	Key   string
	Value string
}

func (d *InspectData) Lines(width int) []InspectLine {
	var out []InspectLine
	section := func(title string) { out = append(out, InspectLine{Kind: InspectLineSection, Key: title}) }
	kv := func(key, value string) {
		out = append(out, InspectLine{Kind: InspectLineKeyValue, Key: key, Value: value})
	}
	val := func(value string) { out = append(out, InspectLine{Kind: InspectLineValue, Value: value}) }
	blank := func() { out = append(out, InspectLine{Kind: InspectLineBlank}) }

	section("Image")
	digest := d.ImageDigest
	if width > 4 && len(digest) > width-4 {
		digest = digest[:width-5] + "…"
	}
	val(digest)
	blank()

	section("Ports")
	if len(d.Ports) == 0 {
		val("(none)")
	} else {
		portKeys := make([]string, 0, len(d.Ports))
		for k := range d.Ports {
			portKeys = append(portKeys, k)
		}
		slices.Sort(portKeys)
		for _, containerPort := range portKeys {
			bindings := d.Ports[containerPort]
			if len(bindings) == 0 {
				kv(containerPort, "→  (not published)")
			} else {
				for _, b := range bindings {
					kv(containerPort, "→  "+b.HostIP+":"+b.HostPort)
				}
			}
		}
	}
	blank()

	section("Environment")
	if len(d.Env) == 0 {
		val("(none)")
	} else {
		for _, e := range d.Env {
			if idx := strings.Index(e, "="); idx > 0 {
				kv(e[:idx]+"=", e[idx+1:])
			} else {
				val(e)
			}
		}
	}
	blank()

	section("Mounts")
	if len(d.Mounts) == 0 {
		val("(none)")
	} else {
		for _, mount := range d.Mounts {
			rw := "ro"
			if mount.RW {
				rw = "rw"
			}
			src := mount.Source
			if src == "" {
				src = "(" + mount.Type + ")"
			}
			kv(src, "→  "+mount.Destination+"  ("+rw+")")
		}
	}
	blank()

	section("Networks")
	if len(d.Networks) == 0 {
		val("(none)")
	} else {
		for _, n := range d.Networks {
			ip := n.IPAddress
			if ip == "" {
				ip = "—"
			}
			kv(n.Name, ip)
		}
	}
	blank()

	return out
}

type inspectRaw struct {
	Image string `json:"Image"`
	State struct {
		Status     string `json:"Status"`
		ExitCode   int    `json:"ExitCode"`
		Error      string `json:"Error"`
		OOMKilled  bool   `json:"OOMKilled"`
		StartedAt  string `json:"StartedAt"`
		FinishedAt string `json:"FinishedAt"`
		Health     *struct {
			Status        string `json:"Status"`
			FailingStreak int    `json:"FailingStreak"`
			Log           []struct {
				Start    string `json:"Start"`
				End      string `json:"End"`
				ExitCode int    `json:"ExitCode"`
				Output   string `json:"Output"`
			} `json:"Log"`
		} `json:"Health"`
	} `json:"State"`
	RestartCount int `json:"RestartCount"`
	Config       struct {
		Env         []string `json:"Env"`
		Healthcheck *struct {
			Test        []string `json:"Test"`
			Interval    int64    `json:"Interval"`
			Timeout     int64    `json:"Timeout"`
			Retries     int      `json:"Retries"`
			StartPeriod int64    `json:"StartPeriod"`
		} `json:"Healthcheck"`
	} `json:"Config"`
	HostConfig struct {
		RestartPolicy struct {
			Name              string `json:"Name"`
			MaximumRetryCount int    `json:"MaximumRetryCount"`
		} `json:"RestartPolicy"`
	} `json:"HostConfig"`
	Mounts          []Mount `json:"Mounts"`
	NetworkSettings struct {
		Ports    map[string][]PortBinding `json:"Ports"`
		Networks map[string]struct {
			IPAddress string `json:"IPAddress"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
}

func parseDockerTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}
	}
	if t.Year() == 1 {
		return time.Time{}
	}
	return t
}

func parseInspectData(out []byte) (*InspectData, error) {
	var raw []inspectRaw
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse inspect output: %w", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("no inspect data returned")
	}
	r := raw[0]
	nets := make([]NetworkInfo, 0, len(r.NetworkSettings.Networks))
	for name, n := range r.NetworkSettings.Networks {
		nets = append(nets, NetworkInfo{Name: name, IPAddress: n.IPAddress})
	}
	slices.SortFunc(nets, func(a, b NetworkInfo) int { return strings.Compare(a.Name, b.Name) })

	state := ContainerState{
		Status:     r.State.Status,
		ExitCode:   r.State.ExitCode,
		Error:      r.State.Error,
		OOMKilled:  r.State.OOMKilled,
		StartedAt:  parseDockerTime(r.State.StartedAt),
		FinishedAt: parseDockerTime(r.State.FinishedAt),
	}
	if r.State.Health != nil {
		h := &Health{
			Status:        r.State.Health.Status,
			FailingStreak: r.State.Health.FailingStreak,
		}
		for _, le := range r.State.Health.Log {
			h.Log = append(h.Log, HealthLogEntry{
				Start:    parseDockerTime(le.Start),
				End:      parseDockerTime(le.End),
				ExitCode: le.ExitCode,
				Output:   le.Output,
			})
		}
		state.Health = h
	}

	var hc *Healthcheck
	if r.Config.Healthcheck != nil {
		hc = &Healthcheck{
			Test:        r.Config.Healthcheck.Test,
			Interval:    time.Duration(r.Config.Healthcheck.Interval),
			Timeout:     time.Duration(r.Config.Healthcheck.Timeout),
			Retries:     r.Config.Healthcheck.Retries,
			StartPeriod: time.Duration(r.Config.Healthcheck.StartPeriod),
		}
	}

	return &InspectData{
		ImageDigest:   r.Image,
		Ports:         r.NetworkSettings.Ports,
		Env:           r.Config.Env,
		Mounts:        r.Mounts,
		Networks:      nets,
		State:         state,
		RestartCount:  r.RestartCount,
		RestartPolicy: RestartPolicy{Name: r.HostConfig.RestartPolicy.Name, MaximumRetryCount: r.HostConfig.RestartPolicy.MaximumRetryCount},
		Healthcheck:   hc,
	}, nil
}

func (CLI) InspectContainer(id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutInspect)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", "inspect", id).CombinedOutput()
		if err != nil {
			return InspectMsg{Err: cmdErr("inspect", out, err)}
		}
		data, err := parseInspectData(out)
		if err != nil {
			return InspectMsg{Err: err}
		}
		return InspectMsg{Data: data}
	}
}

func (CLI) InspectContainerExpand(id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutInspect)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", "inspect", id).CombinedOutput()
		if err != nil {
			return ExpandInspectMsg{ContainerID: id, Err: cmdErr("inspect", out, err)}
		}
		data, err := parseInspectData(out)
		if err != nil {
			return ExpandInspectMsg{ContainerID: id, Err: err}
		}
		return ExpandInspectMsg{ContainerID: id, Data: data}
	}
}

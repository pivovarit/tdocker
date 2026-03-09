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

type InspectData struct {
	ImageDigest string
	Ports       map[string][]PortBinding
	Env         []string
	Mounts      []Mount
	Networks    []NetworkInfo
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
	Image  string `json:"Image"`
	Config struct {
		Env []string `json:"Env"`
	} `json:"Config"`
	Mounts          []Mount `json:"Mounts"`
	NetworkSettings struct {
		Ports    map[string][]PortBinding `json:"Ports"`
		Networks map[string]struct {
			IPAddress string `json:"IPAddress"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
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
	return &InspectData{
		ImageDigest: r.Image,
		Ports:       r.NetworkSettings.Ports,
		Env:         r.Config.Env,
		Mounts:      r.Mounts,
		Networks:    nets,
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

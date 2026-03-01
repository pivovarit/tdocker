package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

type InspectData struct {
	ImageDigest string
	Ports       map[string][]PortBinding
	Env         []string
	Mounts      []Mount
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

	return out
}

type inspectRaw struct {
	Image  string `json:"Image"`
	Config struct {
		Env []string `json:"Env"`
	} `json:"Config"`
	Mounts          []Mount `json:"Mounts"`
	NetworkSettings struct {
		Ports map[string][]PortBinding `json:"Ports"`
	} `json:"NetworkSettings"`
}

func InspectContainer(id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutInspect)
		defer cancel()
		out, err := exec.CommandContext(ctx, "docker", "inspect", id).CombinedOutput()
		if err != nil {
			return InspectMsg{Err: fmt.Errorf("docker inspect: %w\n%s", err, strings.TrimSpace(string(out)))}
		}
		var raw []inspectRaw
		if err := json.Unmarshal(out, &raw); err != nil {
			return InspectMsg{Err: fmt.Errorf("parse inspect output: %w", err)}
		}
		if len(raw) == 0 {
			return InspectMsg{Err: fmt.Errorf("no inspect data returned")}
		}
		r := raw[0]
		return InspectMsg{Data: &InspectData{
			ImageDigest: r.Image,
			Ports:       r.NetworkSettings.Ports,
			Env:         r.Config.Env,
			Mounts:      r.Mounts,
		}}
	}
}

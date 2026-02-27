package docker

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

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
		out, err := exec.Command("docker", "inspect", id).CombinedOutput()
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

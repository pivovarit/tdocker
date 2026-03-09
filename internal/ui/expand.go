package ui

import (
	"fmt"
	"slices"

	"github.com/pivovarit/tdocker/internal/docker"
)

func detailRows(data *docker.InspectData) []docker.Container {
	if data == nil {
		return []docker.Container{
			{State: "detail", Names: "└  loading…"},
		}
	}

	var lines []string

	portKeys := make([]string, 0, len(data.Ports))
	for k := range data.Ports {
		portKeys = append(portKeys, k)
	}
	slices.Sort(portKeys)
	for _, containerPort := range portKeys {
		bindings := data.Ports[containerPort]
		if len(bindings) == 0 {
			lines = append(lines, fmt.Sprintf("Ports    %s (not published)", containerPort))
		} else {
			for _, b := range bindings {
				lines = append(lines, fmt.Sprintf("Ports    %s → %s:%s", containerPort, b.HostIP, b.HostPort))
			}
		}
	}

	for _, n := range data.Networks {
		ip := n.IPAddress
		if ip == "" {
			ip = "—"
		}
		lines = append(lines, fmt.Sprintf("Network  %s (%s)", n.Name, ip))
	}

	if len(lines) == 0 {
		return nil
	}

	rows := make([]docker.Container, len(lines))
	for i, l := range lines {
		prefix := "│  "
		if i == len(lines)-1 {
			prefix = "└  "
		}
		rows[i] = docker.Container{State: "detail", Names: prefix + l}
	}
	return rows
}

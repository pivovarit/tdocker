package ui

import (
	"strconv"
	"strings"
)

func formatPorts(raw string) string {
	if raw == "" {
		return ""
	}

	seen := make(map[string]bool)
	var bindings []string

	for _, part := range strings.Split(raw, ", ") {
		part = strings.TrimSpace(part)
		arrowIdx := strings.Index(part, "->")
		if arrowIdx < 0 {
			continue
		}

		hostPart := part[:arrowIdx]
		containerPart := part[arrowIdx+2:]

		if i := strings.LastIndex(containerPart, "/"); i >= 0 {
			containerPart = containerPart[:i]
		}

		hostPort := hostPart
		if i := strings.LastIndex(hostPart, ":"); i >= 0 {
			hostPort = hostPart[i+1:]
		}

		key := hostPort + "→" + containerPart
		if !seen[key] {
			seen[key] = true
			bindings = append(bindings, key)
		}
	}

	switch len(bindings) {
	case 0:
		return ""
	case 1, 2:
		return strings.Join(bindings, " ")
	default:
		return bindings[0] + " +" + strconv.Itoa(len(bindings)-1)
	}
}

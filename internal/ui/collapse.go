package ui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pivovarit/tdocker/internal/docker"
)

func collapseSummary(project string, containers []docker.Container) docker.Container {
	counts := map[string]int{}
	for _, c := range containers {
		counts[c.State]++
	}

	states := make([]string, 0, len(counts))
	for s := range counts {
		states = append(states, s)
	}

	slices.SortFunc(states, func(a, b string) int {
		rank := func(s string) int {
			switch s {
			case "running":
				return 0
			case "paused":
				return 1
			default:
				return 2
			}
		}
		ra, rb := rank(a), rank(b)
		if ra != rb {
			return ra - rb
		}
		return strings.Compare(a, b)
	})

	parts := make([]string, len(states))
	for i, s := range states {
		parts[i] = fmt.Sprintf("%d %s", counts[s], s)
	}

	return docker.Container{
		Names:  fmt.Sprintf("%s (%s)", project, strings.Join(parts, ", ")),
		State:  "collapsed",
		Labels: docker.Labels{"com.docker.compose.project": project},
	}
}

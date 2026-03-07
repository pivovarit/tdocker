package ui

import "testing"

func TestFormatPorts(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"empty", "", ""},
		{"single binding", "0.0.0.0:8080->80/tcp", "8080→80"},
		{"ipv6 dedup", "0.0.0.0:8080->80/tcp, :::8080->80/tcp", "8080→80"},
		{"two distinct bindings", "0.0.0.0:8080->80/tcp, 0.0.0.0:8443->443/tcp", "8080→80 8443→443"},
		{"two bindings with ipv6 dedup", "0.0.0.0:8080->80/tcp, :::8080->80/tcp, 0.0.0.0:8443->443/tcp, :::8443->443/tcp", "8080→80 8443→443"},
		{"three bindings collapses", "0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp, 0.0.0.0:8080->8080/tcp", "80→80 +2"},
		{"four bindings collapses", "0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp, 0.0.0.0:8080->8080/tcp, 0.0.0.0:9090->9090/tcp", "80→80 +3"},
		{"exposed only no arrow", "80/tcp", ""},
		{"same port ipv4 ipv6 only one shown", "0.0.0.0:5432->5432/tcp, :::5432->5432/tcp", "5432→5432"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatPorts(tc.raw); got != tc.want {
				t.Errorf("formatPorts(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

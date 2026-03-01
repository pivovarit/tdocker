package ui

import (
	"encoding/base64"
	"os"
	"os/exec"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func copyToClipboard(name, text string) tea.Cmd {
	switch runtime.GOOS {
	case "darwin":
		return clipExec(name, text, "pbcopy")
	case "windows":
		return clipExec(name, text, "clip")
	default:
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			return clipExec(name, text, "wl-copy")
		}
		if os.Getenv("DISPLAY") != "" {
			return clipExec(name, text, "xclip", "-selection", "clipboard")
		}
		return clipOSC52(name, text)
	}
}

func clipExec(name, text string, args ...string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			return clipboardMsg{err: err}
		}
		return clipboardMsg{name: name}
	}
}

func clipOSC52(name, text string) tea.Cmd {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	return tea.Batch(
		tea.Printf("\033]52;c;%s\007", encoded),
		func() tea.Msg { return clipboardMsg{name: name} },
	)
}

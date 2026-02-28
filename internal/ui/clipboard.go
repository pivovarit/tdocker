package ui

import (
	"encoding/base64"
	"os"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func copyToClipboard(text string) tea.Cmd {
	switch runtime.GOOS {
	case "darwin":
		return clipExec(text, "pbcopy")
	case "windows":
		return clipExec(text, "clip")
	default:
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			return clipExec(text, "wl-copy")
		}
		if os.Getenv("DISPLAY") != "" {
			return clipExec(text, "xclip", "-selection", "clipboard")
		}
		return clipOSC52(text)
	}
}

func clipExec(text string, args ...string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = strings.NewReader(text)
		_ = cmd.Run()
		return nil
	}
}

func clipOSC52(text string) tea.Cmd {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	return tea.Printf("\033]52;c;%s\007", encoded)
}
